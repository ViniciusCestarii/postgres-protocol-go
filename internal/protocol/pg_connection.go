package protocol

import (
	"fmt"
	"net"
	"postgres-protocol-go/internal/messages"
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

func (pg *PgConnection) Query(query string) (string, error) {
	return ProcessSimpleQuery(*pg, query)
}

func (pg *PgConnection) sendMessage(buf WriteBuffer) error {
	message := buf.Bytes

	if pg.config.Verbose != nil && *pg.config.Verbose {
		utils.LogFrontendRequest(message, buf.IsStartupMessage)
	}

	_, err := pg.conn.Write(message)
	if err != nil {
		return fmt.Errorf("error sending message: %w", err)
	}
	return nil
}

func (pg *PgConnection) readMessage() ([]byte, error) {
	// Read the first 5 bytes to get the message type and length
	header := make([]byte, 5)

	_, err := pg.conn.Read(header)
	if err != nil {
		return nil, fmt.Errorf("error reading from connection: %w", err)
	}

	identifier := utils.ParseIdentifier(header)

	messageLength := utils.ParseMessageLength(header)

	message := make([]byte, messageLength-1)

	// Read the rest of the message
	_, err = pg.conn.Read(message)

	if err != nil {
		return nil, fmt.Errorf("error reading from connection: %w", err)
	}

	fullMessage := append(header, message...)

	if identifier == string(messages.Error) {
		utils.LogBackendAnswer(fullMessage)
		return nil, fmt.Errorf("error from backend: %s", utils.ParseBackendErrorMessage(message))
	}

	if pg.config.Verbose != nil && *pg.config.Verbose {
		utils.LogBackendAnswer(fullMessage)
	}

	return fullMessage, nil
}
