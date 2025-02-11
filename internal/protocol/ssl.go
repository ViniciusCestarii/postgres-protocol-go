package protocol

import (
	"crypto/tls"
	"fmt"
	"postgres-protocol-go/internal/pool"
	"postgres-protocol-go/internal/protocol/messages"
)

func ProcessSSL(pgConnection *PgConnection) error {
	buf := pool.NewWriteBuffer(1024)
	buf.StartMessage(messages.SSL)
	buf.WriteInt32(80877103)
	buf.FinishMessage()

	err := pgConnection.sendMessage(buf)

	if err != nil {
		return err
	}

	answer, err := pgConnection.readSingleByteMessage()

	if err != nil {
		return err
	}

	if string(answer) != "S" {
		return fmt.Errorf("postgresql server is unwilling to perform SSL")
	}

	tlsConn := tls.Client(pgConnection.conn, &tls.Config{
		InsecureSkipVerify: true,
	})

	if err := tlsConn.Handshake(); err != nil {
		return fmt.Errorf("TLS handshake failed: %v", err)
	}

	pgConnection.conn = tlsConn

	if pgConnection.isVerbose() {
		fmt.Println("SSL connection established")
	}

	return nil
}
