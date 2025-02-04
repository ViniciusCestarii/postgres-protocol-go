package protocol

import "postgres-protocol-go/internal/pool"

func SendStartup(pgConnection PgConnection) {

	protocolVersion := int32(3 << 16) // 3 << 16 = 196608 version 3.0

	buf := pool.NewWriteBuffer(1024)

	buf.StartMessage(0)
	buf.WriteInt32(protocolVersion)
	buf.WriteString("user")
	buf.WriteString(pgConnection.config.Username)
	buf.WriteString("database")
	buf.WriteString(*pgConnection.config.Database)
	buf.WriteString("") // must be null-terminated
	buf.FinishMessage()

	pgConnection.sendMessage(buf)
}
