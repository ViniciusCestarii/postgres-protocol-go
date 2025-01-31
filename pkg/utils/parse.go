package utils

import (
	"encoding/binary"
	"fmt"
)

func ParseIdentifier(message []byte) string {
	return string(message[0:1])
}

func ParseMessageLength(message []byte) uint32 {
	return binary.BigEndian.Uint32(message[1:5])
}

func ParseBackendErrorMessage(answer []byte) string {
	codeIdentifier := string(answer[5:6])
	message := string(answer[7:])
	return fmt.Sprintf("Code: %s, Message: %s", codeIdentifier, message)
}
