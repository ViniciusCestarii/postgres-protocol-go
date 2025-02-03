package utils

import (
	"encoding/binary"
	"fmt"
)

func LogBackendAnswer(answer []byte) {
	backendLogSeparator()
	fmt.Printf("Message: %d\n", answer)

	var identifier = ParseIdentifier(answer)
	fmt.Printf("Identifier: %s\n", identifier)

	messageLength := ParseMessageLength(answer)
	fmt.Printf("Message Length: %d\n", messageLength)

	logSeparator()
}

func LogFrontendRequest(request []byte, isStartupMessage bool) {
	frontedLogSeparator()
	fmt.Printf("Message: %d\n", request)

	if isStartupMessage {
		messageLength := binary.BigEndian.Uint32(request[0:4])
		fmt.Printf("Message Length: %d\n", messageLength)
		return
	}
	var identifier = ParseIdentifier(request)
	fmt.Printf("Identifier: %s\n", identifier)

	messageLength := ParseMessageLength(request)
	fmt.Printf("Message Length: %d\n", messageLength)

	logSeparator()
}

func frontedLogSeparator() {
	fmt.Println("--------------------------------------FRONTEND-MESSAGE")
}

func backendLogSeparator() {
	fmt.Println("---------------------------------------BACKEND-MESSAGE")
}

func logSeparator() {
	fmt.Println("------------------------------------------------------")
}
