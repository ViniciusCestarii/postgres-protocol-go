package utils

import (
	"encoding/binary"
	"fmt"
)

func LogBackendAnswer(answer []byte) {
	logSeparator()
	var identifier = ParseIdentifier(answer)
	fmt.Printf("Identifier: %s\n", identifier)

	messageLength := ParseMessageLength(answer)
	fmt.Printf("Message Length: %d\n", messageLength)
}

func LogFrontendRequest(request []byte, isStartupMessage bool) {
	logSeparator()
	fmt.Printf("Fronted Request: %d\n", request)

	if isStartupMessage {
		messageLength := binary.BigEndian.Uint32(request[0:4])
		fmt.Printf("Message Length: %d\n", messageLength)
		return
	}
	var identifier = ParseIdentifier(request)
	fmt.Printf("Identifier: %s\n", identifier)

	messageLength := ParseMessageLength(request)
	fmt.Printf("Message Length: %d\n", messageLength)
}

func logSeparator() {
	fmt.Println("------------------------------------------------------")
}
