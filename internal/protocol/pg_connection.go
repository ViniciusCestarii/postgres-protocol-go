package protocol

import (
	"fmt"
	"net"
	"postgres-protocol-go/internal/messages"
	"postgres-protocol-go/internal/pool"
	"postgres-protocol-go/pkg/models"
	"postgres-protocol-go/pkg/utils"
)

type PgConnection struct {
	conn   net.Conn
	config models.ConnConfig
}

func NewPgConnection(config models.ConnConfig, conn net.Conn) (*PgConnection, error) {
	if conn == nil {
		url := fmt.Sprintf("%s:%d", config.Hostname, config.Port)
		if config.Verbose != nil && *config.Verbose {
			fmt.Printf("Connecting to PostgreSQL at %s\n", url)
		}
		var err error
		conn, err = net.Dial("tcp", url)
		if err != nil {
			return nil, fmt.Errorf("failed to establish a TCP connection to PostgreSQL: %w", err)
		}
	}

	pgConnection := PgConnection{conn: conn, config: config}

	SendStartup(pgConnection)
	err := ProcessAuth(pgConnection)

	if err != nil {
		pgConnection.Close()
		return nil, err
	}

	return &pgConnection, nil
}

// todo support Extended Query
func (pg *PgConnection) Query(query string, params ...interface{}) (*models.QueryResult, error) {

	if len(params) > 0 {
		return ProcessExtendedQuery(*pg, query, params...)
	}

	return ProcessSimpleQuery(*pg, query)
}

func (pg *PgConnection) sendMessage(buf *pool.WriteBuffer) error {
	message := buf.Bytes

	if pg.isVerbose() {
		utils.LogFrontendRequest(message)
	}

	_, err := pg.conn.Write(message)
	if err != nil {
		return fmt.Errorf("error sending message: %w", err)
	}
	return nil
}

func (pg *PgConnection) readMessage() ([]byte, error) {
	header := make([]byte, 5)
	_, err := pg.conn.Read(header)
	if err != nil {
		return nil, fmt.Errorf("error reading from connection: %w", err)
	}

	identifier := utils.ParseIdentifier(header)
	messageLength := utils.ParseMessageLength(header)

	remaining := messageLength - 4

	message := make([]byte, remaining)
	_, err = pg.conn.Read(message)
	if err != nil {
		return nil, fmt.Errorf("error reading from connection: %w", err)
	}

	fullMessage := append(header, message...)

	if identifier == string(messages.Error) {
		return nil, fmt.Errorf("error from backend: %s", utils.ParseBackendErrorMessage(message))
	}

	if pg.isVerbose() {
		utils.LogBackendAnswer(fullMessage)
	}

	return fullMessage, nil
}

func (pg *PgConnection) Close() {
	buf := pool.NewWriteBuffer(5)
	buf.StartMessage(messages.Terminate)
	buf.FinishMessage()

	pg.sendMessage(buf)
	pg.conn.Close()
}

func (pg *PgConnection) isVerbose() bool {
	return pg.config.Verbose != nil && *pg.config.Verbose
}
