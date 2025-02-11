package utils

import (
	"encoding/binary"
	"fmt"
)

func LogBackendAnswer(answer []byte) {
	backendLogSeparator()
	fmt.Printf("Message: %d\n", answer)

	var identifier = string(ParseIdentifierStr(answer))
	fmt.Printf("Identifier: %s\n", identifier)

	messageLength := ParseMessageLength(answer)
	fmt.Printf("Message Length: %d\n", messageLength)

	logSeparator()
}

func LogSingleByteBackendAnswer(answer []byte) {
	backendLogSeparator()
	fmt.Printf("Message: %d\n", answer)

	var identifier = string(ParseIdentifierStr(answer))
	fmt.Printf("Identifier: %s\n", identifier)

	logSeparator()
}

func LogFrontendRequest(request []byte) {
	idx := 0

	for idx < len(request) {
		messageLength := int(binary.BigEndian.Uint32(request[idx+1 : idx+5]))
		LogOneFrontendRequest(request[idx:])
		idx += int(messageLength) + 1
	}
}

func LogOneFrontendRequest(request []byte) {
	frontedLogSeparator()
	fmt.Printf("Message: %d\n", request)

	var identifier = ParseIdentifierStr(request)
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
