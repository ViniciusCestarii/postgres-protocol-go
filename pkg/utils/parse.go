package utils

import (
	"bytes"
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

func ExtractNullTerminatedString(data []byte) string {
	idx := bytes.IndexByte(data, 0)
	if idx == -1 {
		return string(data)
	}
	return string(data[:idx])
}
