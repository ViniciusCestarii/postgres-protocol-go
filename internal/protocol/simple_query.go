package protocol

import (
	"encoding/binary"
	"fmt"
	"postgres-protocol-go/internal/messages"
	"postgres-protocol-go/pkg/models"
	"postgres-protocol-go/pkg/utils"
)

func ProcessSimpleQuery(pgConnection PgConnection, query string) ([]models.Row, error) {
	buf := NewWriteBuffer(1024)
	buf.StartMessage(messages.SimpleQuery)
	buf.WriteString(query)
	buf.FinishMessage()

	err := pgConnection.sendMessage(buf)
	if err != nil {
		return nil, err
	}

	answer, err := pgConnection.readMessage()
	if err != nil {
		return nil, err
	}

	fields, err := parseField(answer)

	if err != nil {
		return nil, err
	}

	if pgConnection.isVerbose() {
		fmt.Printf("Fields: %+v\n", fields)
	}

	rows := make([]models.Row, 0)

	pgConnection.readMessageUntil(func(message []byte) (bool, error) {
		switch utils.ParseIdentifier(message) {
		case string(messages.CommandComplete):
			idx := 5
			tag := utils.ParseNullTerminatedString(message[idx:])

			if tag == fmt.Sprintf("SELECT %d", len(rows)) {
				return true, nil
			}
			return false, nil
		case string(messages.DataRow):
			row := parseDataRow(message, fields)
			rows = append(rows, row)
			return false, nil
		default:
			return false, nil
		}
	})

	if pgConnection.isVerbose() {
		fmt.Printf("Rows: %+v\n", rows)
	}

	return rows, nil
}

func parseDataRow(answer []byte, fields []models.Field) models.Row {
	row := models.Row{Data: make(map[string]interface{})}
	idxRead := 7 // Skip Header

	for _, field := range fields {
		value := parseColumnValue(answer, field, idxRead)
		row.Data[field.Name] = value
	}
	return row
}

func parseField(answer []byte) ([]models.Field, error) {
	identifier := utils.ParseIdentifier(answer)
	if identifier != string(messages.RowDescription) {
		return nil, fmt.Errorf("expected RowDescription message, got %s", identifier)
	}

	numberOfFields := parseNumberOfFields(answer)
	idxRead := 7 // Skip header

	fields := make([]models.Field, numberOfFields)

	for i := uint16(0); i < numberOfFields; i++ {
		fieldName := utils.ParseNullTerminatedString(answer[idxRead:])
		idxRead += len(fieldName) + 1

		tableOID := binary.BigEndian.Uint32(answer[idxRead:])
		idxRead += 4

		columnAttrNum := binary.BigEndian.Uint16(answer[idxRead:])
		idxRead += 2

		dataTypeOID := binary.BigEndian.Uint32(answer[idxRead:])
		idxRead += 4

		dataTypeSize := binary.BigEndian.Uint16(answer[idxRead:])
		idxRead += 2

		typeModifier := binary.BigEndian.Uint32(answer[idxRead:])
		idxRead += 4

		formatCode := binary.BigEndian.Uint16(answer[idxRead:])
		idxRead += 2

		fields[i] = models.Field{
			Name:         fieldName,
			TableOID:     tableOID,
			AttrNum:      columnAttrNum,
			DataTypeOID:  dataTypeOID,
			Size:         dataTypeSize,
			TypeModifier: typeModifier,
			FormatCode:   formatCode,
		}
	}

	return fields, nil
}

func parseNumberOfFields(message []byte) uint16 {
	return binary.BigEndian.Uint16(message[5:7])
}

func parseNumberOfColumns(message []byte) uint16 {
	return binary.BigEndian.Uint16(message[5:7])
}

func parseColumnValue(answer []byte, field models.Field, idxRead int) any {
	columnValueLength := int32(binary.BigEndian.Uint32(answer[idxRead:]))
	idxRead += 4

	if columnValueLength == -1 {
		return nil
	}

	value := answer[idxRead : idxRead+int(columnValueLength)]

	switch field.FormatCode {
	case 0: // text
		return string(value)
	case 1: // binary
		return value
	}

	return nil
}
