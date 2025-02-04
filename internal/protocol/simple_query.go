package protocol

import (
	"postgres-protocol-go/internal/pool"
	"postgres-protocol-go/internal/protocol/messages"
	"postgres-protocol-go/pkg/models"
)

func ProcessSimpleQuery(pgConnection PgConnection, query string) (*models.QueryResult, error) {
	buf := pool.NewWriteBuffer(1024)
	buf.StartMessage(messages.SimpleQuery)
	buf.WriteString(query)
	buf.FinishMessage()

	err := pgConnection.sendMessage(buf)
	if err != nil {
		return nil, err
	}

	return processQueryResult(pgConnection, query)
}
