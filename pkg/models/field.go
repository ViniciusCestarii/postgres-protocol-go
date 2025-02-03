package models

type Field struct {
	Name         string
	TableOID     uint32
	AttrNum      uint16
	DataTypeOID  uint32
	Size         uint16
	TypeModifier uint32
	FormatCode   uint16
}
