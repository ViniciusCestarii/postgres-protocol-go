package protocol

import (
	"fmt"
	"net"
	"postgres-protocol-go/pkg/models"
	"postgres-protocol-go/pkg/utils"
)

func ProcessStartup(conn net.Conn, config models.ConnConfig) {
	messageContent := make([]byte, 0)

	protocolVersion := int32(3 << 16) // 3 << 16 = 196608 version 3.0

	messageContent = append(messageContent, utils.Int32ToBytes(protocolVersion)...)

	// Parameters: user is required
	// if no database is specified, the user name will be used
	parameters := map[string]*string{
		"user":     &config.Username,
		"database": config.Database,
	}

	for key, value := range parameters {
		if value == nil {
			continue
		}
		messageContent = append(messageContent, utils.StringToBytes(key)...)
		messageContent = append(messageContent, utils.StringToBytes(*value)...)
	}

	// Add a null byte at the end
	messageContent = append(messageContent, 0)

	messageContent = utils.AppendMessageLength(messageContent)

	// Send the message
	_, err := conn.Write(messageContent)
	if err != nil {
		fmt.Println("Error sending message:", err)
	}
}
