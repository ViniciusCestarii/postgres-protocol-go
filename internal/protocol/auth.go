package protocol

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"postgres-protocol-go/internal/pool"
	"postgres-protocol-go/internal/protocol/messages"
	"postgres-protocol-go/pkg/utils"
	"strings"
)

func ProcessAuth(pgConnection PgConnection) error {
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

		for {
			message, err := pgConnection.readMessage()
			if err != nil {
				return err
			}

			// there are other useful messages that can be processed here like client_enconding, DateStyle, BackendKeyData, etc.
			switch utils.ParseIdentifier(message) {
			case string(messages.ReadyForQuery):
				return nil
			default:
				if pgConnection.isVerbose() {
					fmt.Printf("Auth: Unknown message: %s\n", string(message))
				}
			}
		}
	case authenticationSASL:
		pgConnection.saslMethod = strings.Trim(string(answer[9:]), "\x00 \n\r")
		switch pgConnection.saslMethod {
		// https://datatracker.ietf.org/doc/html/rfc7677
		case "SCRAM-SHA-256":
			initialResponse, err := buildSCRAMInitialResponse(pgConnection.connConfig.User)
			if err != nil {
				return err
			}

			initialResponseBytes := []byte(initialResponse)

			lengthBytes := make([]byte, 4)
			binary.BigEndian.PutUint32(lengthBytes, uint32(len(initialResponseBytes)))
			fmt.Println(lengthBytes)

			buff := pool.NewWriteBuffer(1024)
			buff.StartMessage(messages.SASLInitial)
			buff.WriteString(pgConnection.saslMethod)
			buff.WriteInt32(int32(len(initialResponseBytes)))
			_, err = buff.Write(initialResponseBytes)
			if err != nil {
				return err
			}
			buff.FinishMessage()

			err = pgConnection.sendMessage(buff)
			if err != nil {
				return err
			}

			return ProcessAuth(pgConnection)
		default:
			return fmt.Errorf("SASL authentication method %s is not supported", pgConnection.saslMethod)
		}
	case authenticationSASLContinue:
		fmt.Println("SASLContinue", pgConnection.saslMethod)
		switch pgConnection.saslMethod {
		case "SCRAM-SHA-256":
			return fmt.Errorf("not implemented")
		default:
			return fmt.Errorf("SASL authentication method %s is not supported", pgConnection.saslMethod)
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

		err := pgConnection.sendMessage(buf)

		if err != nil {
			return err
		}

		return ProcessAuth(pgConnection)
	case authenticationCleartextPassword:
		if pgConnection.connConfig.Password == nil {
			return fmt.Errorf("password is required for cleartext authentication")
		}

		buf := pool.NewWriteBuffer(1024)
		buf.StartMessage(messages.Password)
		buf.WriteString(*pgConnection.connConfig.Password)
		buf.FinishMessage()

		err := pgConnection.sendMessage(buf)
		if err != nil {
			return err
		}

		return ProcessAuth(pgConnection)
	default:
		return fmt.Errorf("unsupported authentication method: %d", authType)
	}
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

func buildSCRAMInitialResponse(username string) (string, error) {
	nonce, err := generateNonce()
	if err != nil {
		return "", fmt.Errorf("failed to generate nonce: %v", err)
	}
	// Format: "n,,n=<username>,r=<nonce>"
	initialResponse := fmt.Sprintf("n,,n=%s,r=%s", username, nonce)
	return initialResponse, nil
}

func generateNonce() (string, error) {
	nonceBytes := make([]byte, 16) // 16 bytes = 128 bits of randomness
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(nonceBytes), nil
}
