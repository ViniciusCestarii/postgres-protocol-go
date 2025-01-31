package messages

// https://www.postgresql.org/docs/current/protocol-message-formats.html
const (
	Startup  = 0 // No identifier
	Auth     = 'R'
	Password = 'p'
	Error    = 'E'
)
