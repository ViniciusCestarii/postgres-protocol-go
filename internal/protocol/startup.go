package protocol

import (
	"postgres-protocol-go/pkg/messages"
	"postgres-protocol-go/pkg/utils"
)

func ProcessStartup(pgConnection PgConnection) {
	messageContent := make([]byte, 0)

	protocolVersion := int32(3 << 16) // 3 << 16 = 196608 version 3.0

	messageContent = append(messageContent, utils.Int32ToBytes(protocolVersion)...)

	// Parameters: user is required
	// if no database is specified, the user name will be used
	parameters := map[string]*string{
		"user":     &pgConnection.config.Username,
		"database": pgConnection.config.Database,
	}

	for key, value := range parameters {
		if value == nil {
			continue
		}
		messageContent = append(messageContent, utils.StringToBytes(key)...)
		messageContent = append(messageContent, utils.StringToBytes(*value)...)
	}

	pgConnection.SendMessage(messages.Startup, messageContent)
}
