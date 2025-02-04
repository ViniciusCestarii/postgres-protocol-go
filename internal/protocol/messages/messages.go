package messages

import "postgres-protocol-go/internal/pool"

// https://www.postgresql.org/docs/current/protocol-message-formats.html
const (
	Startup         = 0 // No identifier
	Auth            = 'R'
	Password        = 'p'
	Error           = 'E'
	SimpleQuery     = 'Q'
	Parse           = 'P'
	Describe        = 'D'
	ParseComplete   = '1'
	Bind            = 'B'
	Sync            = 'S'
	Terminate       = 'X'
	ReadyForQuery   = 'Z'
	RowDescription  = 'T'
	DataRow         = 'D'
	CommandComplete = 'C'
	Notice          = 'N'
	Execute         = 'E'
)

func WriteSyncMsg(buf *pool.WriteBuffer) {
	buf.StartMessage(Sync)
	buf.FinishMessage()
}
