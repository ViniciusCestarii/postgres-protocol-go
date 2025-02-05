package protocol

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	pbkdf2 "postgres-protocol-go"
	"postgres-protocol-go/internal/pool"
	"postgres-protocol-go/internal/protocol/messages"
	"postgres-protocol-go/pkg/utils"
	"strconv"
	"strings"
)

func ProcessAuth(pgConnection PgConnection) error {
	var (
		saslMethod        string
		clientNonce       string
		expectedServerSig []byte
	)

	for {
		answer, err := pgConnection.readMessage()
		if err != nil {
			return err
		}

		identifier := utils.ParseIdentifier(answer)
		if identifier != string(messages.Auth) {
			return fmt.Errorf("expected auth message, got %s", identifier)
		}

		authType := parseAuthType(answer)

		switch authType {
		case authenticationOk:
			if pgConnection.isVerbose() {
				fmt.Println("Authentication successful")
				fmt.Println("Waiting for ReadyForQuery message")
			}
			return waitForReady(pgConnection)
		case authenticationSASL:
			saslMethod = strings.Trim(string(answer[9:]), "\x00 \n\r")
			switch saslMethod {
			case "SCRAM-SHA-256":
				nonce, initialResponse, err := buildSCRAMInitialResponse(pgConnection.connConfig.User)
				if err != nil {
					return err
				}
				clientNonce = nonce

				buff := pool.NewWriteBuffer(1024)
				buff.StartMessage(messages.SASLInitial)
				buff.WriteString(saslMethod)
				buff.WriteInt32(int32(len(initialResponse)))
				buff.Write(initialResponse)
				buff.FinishMessage()

				if err := pgConnection.sendMessage(buff); err != nil {
					return err
				}
			default:
				return fmt.Errorf("SASL authentication method %s is not supported", saslMethod)
			}
		case authenticationSASLContinue:
			switch saslMethod {
			case "SCRAM-SHA-256":
				if clientNonce == "" {
					return fmt.Errorf("client nonce not available")
				}
				if pgConnection.connConfig.Password == nil {
					return fmt.Errorf("password is required for SCRAM authentication")
				}

				serverMessage := string(answer[9:])
				parts := strings.Split(serverMessage, ",")
				var serverNonce, saltB64 string
				var iterations int
				for _, part := range parts {
					kv := strings.SplitN(part, "=", 2)
					if len(kv) != 2 {
						return fmt.Errorf("invalid part in SASLContinue message: %s", part)
					}
					key, value := kv[0], kv[1]
					switch key {
					case "r":
						serverNonce = value
					case "s":
						saltB64 = value
					case "i":
						i, err := strconv.Atoi(value)
						if err != nil {
							return fmt.Errorf("invalid iteration count: %v", err)
						}
						iterations = i
					default:
						return fmt.Errorf("unexpected key in SASLContinue message: %s", key)
					}
				}

				if !strings.HasPrefix(serverNonce, clientNonce) {
					return fmt.Errorf("server nonce does not start with client nonce")
				}

				salt, err := base64.StdEncoding.DecodeString(saltB64)
				if err != nil {
					return fmt.Errorf("failed to decode salt: %v", err)
				}

				password := []byte(*pgConnection.connConfig.Password)
				saltedPassword := pbkdf2.Key(password, salt, iterations, 32, sha256.New)

				clientKey := hmac.New(sha256.New, saltedPassword)
				clientKey.Write([]byte("Client Key"))
				clientKeyBytes := clientKey.Sum(nil)

				storedKey := sha256.Sum256(clientKeyBytes)

				clientFirstBare := fmt.Sprintf("n=%s,r=%s", pgConnection.connConfig.User, clientNonce)
				serverFirst := fmt.Sprintf("r=%s,s=%s,i=%d", serverNonce, saltB64, iterations)
				clientFinalWithoutProof := fmt.Sprintf("c=biws,r=%s", serverNonce)
				authMessage := strings.Join([]string{clientFirstBare, serverFirst, clientFinalWithoutProof}, ",")

				clientSignature := hmac.New(sha256.New, storedKey[:])
				clientSignature.Write([]byte(authMessage))
				clientSignatureBytes := clientSignature.Sum(nil)

				clientProof := make([]byte, len(clientKeyBytes))
				for i := 0; i < len(clientKeyBytes); i++ {
					clientProof[i] = clientKeyBytes[i] ^ clientSignatureBytes[i]
				}
				clientProofB64 := base64.StdEncoding.EncodeToString(clientProof)

				serverKey := hmac.New(sha256.New, saltedPassword)
				serverKey.Write([]byte("Server Key"))
				serverKeyBytes := serverKey.Sum(nil)

				expectedServerSigHasher := hmac.New(sha256.New, serverKeyBytes)
				expectedServerSigHasher.Write([]byte(authMessage))
				expectedServerSig = expectedServerSigHasher.Sum(nil)

				clientFinalMessage := fmt.Sprintf("c=biws,r=%s,p=%s", serverNonce, clientProofB64)

				buf := pool.NewWriteBuffer(1024)
				buf.StartMessage(messages.SASLResponse)
				buf.Write([]byte(clientFinalMessage))
				buf.FinishMessage()

				if err := pgConnection.sendMessage(buf); err != nil {
					return err
				}
			default:
				return fmt.Errorf("SASL authentication method %s is not supported", saslMethod)
			}
		case authenticationSASLFinal:
			switch saslMethod {
			case "SCRAM-SHA-256":
				serverMessage := string(answer[9:])

				fmt.Println("server messsage:", serverMessage)
				var serverSigB64 string
				for _, part := range strings.Split(serverMessage, ",") {
					kv := strings.SplitN(part, "=", 2)
					if len(kv) != 2 {
						continue
					}
					if kv[0] == "v" {
						serverSigB64 = kv[1]
						break
					}
				}

				if serverSigB64 == "" {
					return fmt.Errorf("missing server signature in SASLFinal message")
				}

				serverSig, err := base64.StdEncoding.DecodeString(serverSigB64)
				if err != nil {
					return fmt.Errorf("failed to decode server signature: %v", err)
				}

				if !bytes.Equal(serverSig, expectedServerSig) {
					return fmt.Errorf("server signature mismatch")
				}
			default:
				return fmt.Errorf("SASL authentication method %s is not supported", saslMethod)
			}

		case authenticationMD5Password:
			if pgConnection.connConfig.Password == nil {
				return fmt.Errorf("password is required for MD5 authentication")
			}
			salt := parseSalt(answer)
			hashedPassword := hashPasswordMD5(*pgConnection.connConfig.Password, pgConnection.connConfig.User, string(salt))

			buf := pool.NewWriteBuffer(1024)
			buf.StartMessage(messages.Password)
			buf.WriteString(hashedPassword)
			buf.FinishMessage()

			if err := pgConnection.sendMessage(buf); err != nil {
				return err
			}

		case authenticationCleartextPassword:
			if pgConnection.connConfig.Password == nil {
				return fmt.Errorf("password is required for cleartext authentication")
			}

			buf := pool.NewWriteBuffer(1024)
			buf.StartMessage(messages.Password)
			buf.WriteString(*pgConnection.connConfig.Password)
			buf.FinishMessage()

			if err := pgConnection.sendMessage(buf); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported authentication method: %d", authType)
		}
	}
}

func waitForReady(pgConnection PgConnection) error {
	for {
		message, err := pgConnection.readMessage()
		if err != nil {
			return err
		}

		switch utils.ParseIdentifier(message) {
		case string(messages.ReadyForQuery):
			return nil
		default:
			if pgConnection.isVerbose() {
				fmt.Printf("Auth: Unknown message: %s\n", string(message))
			}
		}
	}
}

func buildSCRAMInitialResponse(username string) (string, []byte, error) {
	nonce, err := generateNonce()
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate nonce: %v", err)
	}
	initialResponse := fmt.Sprintf("n,,n=%s,r=%s", username, nonce)
	return nonce, []byte(initialResponse), nil
}

func generateNonce() (string, error) {
	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(nonceBytes), nil
}

func hashPasswordMD5(password, username, salt string) string {
	return "md5" + md5Hash(md5Hash(password+username)+salt)
}

func md5Hash(input string) string {
	hash := md5.Sum([]byte(input))
	return hex.EncodeToString(hash[:])
}

const (
	authenticationOk                = 0
	authenticationKerberosV5        = 2
	authenticationCleartextPassword = 3
	authenticationMD5Password       = 5
	authenticationGSS               = 7
	authenticationGSSContinue       = 8
	authenticationSSPI              = 9
	authenticationSASL              = 10
	authenticationSASLContinue      = 11
	authenticationSASLFinal         = 12
)

func parseAuthType(message []byte) uint32 {
	return binary.BigEndian.Uint32(message[5:9])
}

func parseSalt(message []byte) string {
	return string(message[9:13])
}
