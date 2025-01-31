package protocol

import (
	"encoding/binary"
	"fmt"
	"net"
	"postgres-protocol-go/pkg/messages"
	"postgres-protocol-go/pkg/models"
	"postgres-protocol-go/pkg/utils"
)

type PgConnection struct {
	conn   net.Conn
	config models.ConnConfig
}

func NewPgConnection(config models.ConnConfig, conn net.Conn) (*PgConnection, error) {
	if conn == nil {
		var err error
		conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", config.Hostname, config.Port))
		if err != nil {
			return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
		}
	}

	pgConnection := PgConnection{conn: conn, config: config}

	ProcessStartup(pgConnection)
	err := ProcessAuth(pgConnection)

	if err != nil {
		return nil, err
	}

	return &pgConnection, nil
}

func (pg *PgConnection) SendMessage(identifier byte, message []byte) error {

	message = append(message, 0) // Ensure single null terminator
	message = utils.AppendMessageLength(message)

	isStartupMessage := identifier == messages.Startup

	if !isStartupMessage {
		message = append([]byte{identifier}, message...)
	}

	if pg.config.Verbose != nil && *pg.config.Verbose {
		utils.LogFrontendRequest(message, isStartupMessage)
	}

	_, err := pg.conn.Write(message)
	if err != nil {
		return fmt.Errorf("error sending message: %w", err)
	}
	return nil
}

func (pg *PgConnection) ReadMessage() ([]byte, error) {
	// Read the first 5 bytes to get the message type and length
	header := make([]byte, 5)
	_, err := pg.conn.Read(header)
	if err != nil {
		return nil, fmt.Errorf("error reading from connection: %w", err)
	}

	messageLength := binary.BigEndian.Uint32(header[1:5])

	message := make([]byte, messageLength-1)

	// Read the rest of the message
	_, err = pg.conn.Read(message)

	if err != nil {
		return nil, fmt.Errorf("error reading from connection: %w", err)
	}

	fullMessage := append(header, message...)

	if pg.config.Verbose != nil && *pg.config.Verbose {
		utils.LogBackendAnswer(fullMessage)
	}

	return fullMessage, nil
}
