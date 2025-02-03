package models

// todo: make row generic
type QueryResult struct {
	Command  string
	Fields   []Field
	Oid      uint16
	RowCount int
	Rows     map[string]interface{}
}

type Field struct {
	Name         string
	TableOID     uint32
	AttrNum      uint16
	DataTypeOID  uint32
	Size         uint16
	TypeModifier uint32
	Format       string // text | binary
}
