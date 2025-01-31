package utils

import (
	"encoding/binary"
	"fmt"
	enum_auth "postgres-protocol-go/pkg/enum"
)

func ParseIdentifier(answer []byte) string {
	return string(answer[0:1])
}

func ParseMessageLength(answer []byte) uint32 {
	return binary.BigEndian.Uint32(answer[1:5])
}

func ParseAuthenticationMethod(answer []byte) (enum_auth.AuthenticationMethod, error) {
	if len(answer) < 9 {
		return 0, fmt.Errorf("invalid message length: %d", len(answer))
	}

	authMethodID := binary.BigEndian.Uint32(answer[5:9])

	authMethod := enum_auth.AuthenticationMethod(authMethodID)

	if authMethod.String() == "Unknown" {
		return 0, fmt.Errorf("unknown authentication method: %d", authMethodID)
	}

	return authMethod, nil
}

func ParseErrorMessage(answer []byte) string {
	codeIdentifier := string(answer[5:6])
	message := string(answer[7:])
	return fmt.Sprintf("Code: %s, Message: %s", codeIdentifier, message)
}
