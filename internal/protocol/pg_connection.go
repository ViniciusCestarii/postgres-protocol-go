package protocol

import (
	"fmt"
	"net"
	"net/url"
	"postgres-protocol-go/internal/pool"
	"postgres-protocol-go/internal/protocol/messages"
	"postgres-protocol-go/pkg/models"
	"postgres-protocol-go/pkg/utils"
	"strconv"
	"strings"
)

type PgConnection struct {
	conn        net.Conn
	connConfig  models.ConnConfig
	driveConfig models.DriveConfig
}

func NewPgConnection(connStr string, driveConfig models.DriveConfig) (*PgConnection, error) {
	connConfig, err := parseConnStr(connStr)

	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s:%d", connConfig.Host, connConfig.Port)

	if driveConfig.Verbose {
		fmt.Printf("Connecting to PostgreSQL at %s\n", url)
	}

	conn, err := net.Dial("tcp", url)

	if err != nil {
		return nil, fmt.Errorf("failed to establish a TCP connection to PostgreSQL: %w", err)
	}

	pgConnection := PgConnection{conn: conn, connConfig: connConfig, driveConfig: driveConfig}

	if connConfig.Secure {
		err = ProcessSSL(&pgConnection)
		if err != nil {
			pgConnection.Close()
			return nil, err
		}
	}

	SendStartup(pgConnection)
	err = ProcessAuth(pgConnection)

	if err != nil {
		pgConnection.Close()
		return nil, err
	}

	return &pgConnection, nil
}

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

func (pg *PgConnection) readSingleByteMessage() ([]byte, error) {
	message := make([]byte, 1)
	_, err := pg.conn.Read(message)
	if err != nil {
		return nil, fmt.Errorf("error reading from connection: %w", err)
	}

	if pg.isVerbose() {
		utils.LogSingleByteBackendAnswer(message)
	}

	return message, nil
}

func (pg *PgConnection) Close() {
	buf := pool.NewWriteBuffer(5)
	buf.StartMessage(messages.Terminate)
	buf.FinishMessage()

	pg.sendMessage(buf)
	pg.conn.Close()
}

func (pg *PgConnection) isVerbose() bool {
	return pg.driveConfig.Verbose
}

func parseConnStr(connUrl string) (models.ConnConfig, error) {
	connConfig := models.ConnConfig{}

	if strings.HasPrefix(connUrl, "postgres://") {
		parsedUrl, err := url.Parse(connUrl)
		if err != nil {
			return connConfig, fmt.Errorf("failed to parse connection URL: %w", err)
		}
		connConfig.Host = parsedUrl.Hostname()
		port := parsedUrl.Port()
		if port != "" {
			portInt, err := strconv.Atoi(port)
			if err == nil {
				connConfig.Port = portInt
			}
		}
		connConfig.User = parsedUrl.User.Username()
		if password, ok := parsedUrl.User.Password(); ok {
			connConfig.Password = &password
		}
		path := strings.TrimPrefix(parsedUrl.Path, "/")
		if path != "" {
			connConfig.Database = &path
		}
		if parsedUrl.Query().Get("sslmode") == "require" {
			connConfig.Secure = true
		}
		return connConfig, nil
	}

	split := strings.Fields(connUrl)

	for _, s := range split {
		if strings.HasPrefix(s, "host=") {
			host := strings.SplitN(s, "=", 2)[1]
			connConfig.Host = host
			continue
		}
		if strings.HasPrefix(s, "port=") {
			portStr := strings.SplitN(s, "=", 2)[1]
			port, err := strconv.Atoi(portStr)
			if err == nil {
				connConfig.Port = port
			}
			continue
		}
		if strings.HasPrefix(s, "user=") {
			user := strings.SplitN(s, "=", 2)[1]
			connConfig.User = user
			continue
		}
		if strings.HasPrefix(s, "dbname=") {
			dbname := strings.SplitN(s, "=", 2)[1]
			connConfig.Database = &dbname
			continue
		}
		if strings.HasPrefix(s, "password=") {
			password := strings.SplitN(s, "=", 2)[1]
			connConfig.Password = &password
			continue
		}
		if strings.HasPrefix(s, "sslmode=") {
			sslmode := strings.SplitN(s, "=", 2)[1]
			if sslmode == "require" {
				connConfig.Secure = true
			}
			continue
		}
	}

	return connConfig, nil
}
