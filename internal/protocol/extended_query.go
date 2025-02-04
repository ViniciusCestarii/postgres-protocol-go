package protocol

import (
	"postgres-protocol-go/internal/pool"
	"postgres-protocol-go/internal/protocol/messages"
	"postgres-protocol-go/pkg/models"
	"postgres-protocol-go/pkg/types"
)

func ProcessExtendedQuery(pgConnection PgConnection, query string, params ...interface{}) (*models.QueryResult, error) {
	buf := pool.NewWriteBuffer(1024)
	buf.StartMessage(messages.Parse)
	buf.WriteString("") // unnamed statement
	buf.WriteString(query)
	buf.WriteInt16(0) // don't want to prespecify types for parameters
	buf.FinishMessage()

	buf.StartMessage(messages.Describe)
	buf.WriteByte('S')
	buf.WriteString("") // unnamed statement
	buf.FinishMessage()

	buf.StartMessage(messages.Bind)
	buf.WriteString("") // unnamed portal
	buf.WriteString("") // unnamed statement
	buf.WriteInt16(0)   // don't want to prespecify types for parameters
	buf.WriteInt16(int16(len(params)))
	for _, param := range params {
		buf.StartParam()
		bytes := types.Append(buf.Bytes, param, 0)
		if bytes != nil {
			buf.Bytes = bytes
			buf.FinishParam()
		} else {
			buf.FinishNullParam()
		}
	}
	buf.WriteInt16(0)
	buf.FinishMessage()

	buf.StartMessage(messages.Execute)
	buf.WriteString("")
	buf.WriteInt32(0)
	buf.FinishMessage()

	messages.WriteSyncMsg(buf)

	err := pgConnection.sendMessage(buf)

	if err != nil {
		return nil, err
	}

	return processQueryResult(pgConnection, query)
}
