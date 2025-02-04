// Credits to https://github.com/go-pg/pg/tree/v10/types

package types

type ValueAppender interface {
	AppendValue(b []byte, flags int) ([]byte, error)
}

//------------------------------------------------------------------------------

// Safe represents a safe SQL query.
type Safe string

var _ ValueAppender = (*Safe)(nil)

func (q Safe) AppendValue(b []byte, flags int) ([]byte, error) {
	return append(b, q...), nil
}
