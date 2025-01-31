package protocol

import "postgres-protocol-go/internal/messages"

func ProcessSimpleQuery(pgConnection PgConnection, query string) (string, error) {
	buf := NewWriteBuffer(1024)
	buf.StartMessage(messages.SimpleQuery)
	buf.WriteString(query)
	buf.FinishMessage()

	err := pgConnection.sendMessage(buf)

	if err != nil {
		return "", err
	}

	answer, err := pgConnection.readMessage()

	if err != nil {
		return "", err
	}

	return string(answer), nil
}
