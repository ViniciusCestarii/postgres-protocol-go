package protocol

import (
	"crypto/tls"
	"fmt"
	"postgres-protocol-go/internal/pool"
	"postgres-protocol-go/internal/protocol/messages"
)

func ProcessSSL(pgConnection *PgConnection) error {
	buff := pool.NewWriteBuffer(1024)
	buff.StartMessage(messages.SSL)
	buff.WriteInt32(80877103)
	buff.FinishMessage()

	err := pgConnection.sendMessage(buff)

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
