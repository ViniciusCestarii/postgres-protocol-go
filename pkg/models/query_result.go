package models

// todo: make rows generic
type QueryResult struct {
	Command  string
	Fields   []Field
	RowCount int
	Rows     []map[string]interface{}
}

// todo: parse these int for compreensible values
type Field struct {
	Name         string
	TableOID     uint32
	AttrNum      uint16
	DataTypeOID  uint32
	Size         uint16
	TypeModifier uint32
	Format       string // text | binary
}
