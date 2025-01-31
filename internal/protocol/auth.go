package protocol

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	enum_auth "postgres-protocol-go/pkg/enum"
	"postgres-protocol-go/pkg/models"
	"postgres-protocol-go/pkg/utils"
)

func ProcessAuth(conn net.Conn, answer []byte, config models.ConnConfig) error {
	var authType, err = utils.ParseAuthenticationMethod(answer)

	if err != nil {
		fmt.Println("Error parsing authentication method:", err)
		return nil
	}

	fmt.Println("Auth type:", authType.String())

	switch authType {
	case enum_auth.AuthenticationMD5Password:
		if config.Password == nil {
			return fmt.Errorf("password is required for MD5 authentication")
		}
		salt := answer[9:13]

		hashedPassword := hashPasswordMD5(*config.Password, config.Username, string(salt))

		messageContent := make([]byte, 0)
		messageContent = append(messageContent, utils.StringToBytes(hashedPassword)...)
		messageContent = append(messageContent, 0) // Ensure single null terminator

		messageContent = append(utils.Int32ToBytes(int32(len(messageContent)+3)), messageContent...)

		finalMessage := make([]byte, 0)
		finalMessage = append(finalMessage, 'p')
		finalMessage = append(finalMessage, messageContent...)

		fmt.Printf(string(finalMessage[0:1]))
		fmt.Printf("Message Length: %d\n", binary.BigEndian.Uint32(finalMessage[1:5]))

		fmt.Println("Sending authentication message:", finalMessage)

		_, err := conn.Write(finalMessage)
		if err != nil {
			fmt.Println("Error sending message:", err)
		}
		conn.Read(answer)

		utils.LogServerAnswer(answer)

		identifier := utils.ParseIdentifier(answer)

		switch identifier {
		case "E":
			return fmt.Errorf("error authenticating: %s", utils.ParseErrorMessage(answer))
		default:
			fmt.Println(utils.ParseAuthenticationMethod(answer))
		}

		return nil
	default:
		return fmt.Errorf("unsupported authentication method: %s", authType)
	}
}

func hashPasswordMD5(password, username, salt string) string {
	return "md5" + md5Hash(md5Hash(password+username)+salt)
}

func md5Hash(input string) string {
	hash := md5.Sum([]byte(input))
	return hex.EncodeToString(hash[:])
}
