package protocol

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"postgres-protocol-go/pkg/messages"
	"postgres-protocol-go/pkg/utils"
)

func ProcessAuth(pgConnection PgConnection) error {
	answer, err := pgConnection.ReadMessage()

	if err != nil {
		return fmt.Errorf("error reading from connection: %w", err)
	}

	identifier := utils.ParseIdentifier(answer)

	if identifier != string(messages.Auth) {
		return fmt.Errorf("expected auth message, got %s", identifier)
	}

	authType := parseAuthType(answer)

	switch authType {
	case authenticationOk:
		fmt.Println("Authentication successful")
		return nil
	case authenticationMD5Password:
		if pgConnection.config.Password == nil {
			return fmt.Errorf("password is required for MD5 authentication")
		}

		salt := parseSalt(answer)

		hashedPassword := hashPasswordMD5(*pgConnection.config.Password, pgConnection.config.Username, string(salt))

		messageContent := make([]byte, 0)
		messageContent = append(messageContent, hashedPassword...)

		err := pgConnection.SendMessage(messages.Password, messageContent)

		if err != nil {
			return err
		}

		ProcessAuth(pgConnection)

		return nil
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
