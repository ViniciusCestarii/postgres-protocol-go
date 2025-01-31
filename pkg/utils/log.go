package utils

import (
	"fmt"
)

func LogServerAnswer(answer []byte) {
	fmt.Printf("Server Answer: %d", answer)
	var identifier = ParseIdentifier(answer)

	fmt.Println(identifier)

	messageLength := ParseMessageLength(answer)
	fmt.Printf("Message Length: %d\n", messageLength)
}
