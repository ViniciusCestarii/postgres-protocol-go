package protocol

import (
	"encoding/binary"
	"fmt"
	"postgres-protocol-go/internal/messages"
	"postgres-protocol-go/pkg/models"
	"postgres-protocol-go/pkg/utils"
	"strings"
)

func ProcessSimpleQuery(pgConnection PgConnection, query string) (*models.QueryResult, error) {
	buf := NewWriteBuffer(1024)
	buf.StartMessage(messages.SimpleQuery)
	buf.WriteString(query)
	buf.FinishMessage()

	err := pgConnection.sendMessage(buf)
	if err != nil {
		return nil, err
	}

	queryResult := &models.QueryResult{
		Command: strings.Fields(query)[0], // Default to first word of the query
		Rows:    make([]map[string]interface{}, 0),
	}

	var fields []models.Field

	for {
		message, err := pgConnection.readMessage()
		if err != nil {
			return nil, err
		}

		switch utils.ParseIdentifier(message) {
		case string(messages.RowDescription):
			fields, err = parseField(message)
			if err != nil {
				return nil, err
			}
			queryResult.Fields = fields

		case string(messages.DataRow):
			row := parseDataRow(message, fields)
			queryResult.Rows = append(queryResult.Rows, row)

		case string(messages.CommandComplete):
			parts := strings.Fields(string(message))
			if len(parts) > 0 {
				queryResult.Command = parts[0] // Extract command name
			}
			queryResult.RowCount = len(queryResult.Rows)

		case string(messages.Error):
			identifierFieldType := string(message[5:6])
			if identifierFieldType == "0" {
				return nil, fmt.Errorf("PostgreSQL error: %s", utils.ParseNullTerminatedString(message[6:]))
			}
			return nil, fmt.Errorf("PostgreSQL error: %s: %s", identifierFieldType, utils.ParseNullTerminatedString(message[6:]))

		case string(messages.Notice):
			identifierFieldType := string(message[5:6])
			if pgConnection.isVerbose() {
				continue
			}

			if identifierFieldType == "0" {
				fmt.Printf("PostgreSQL notice: %s", utils.ParseNullTerminatedString(message[6:]))
			}

			fmt.Printf("PostgreSQL notice: %s: %s", identifierFieldType, utils.ParseNullTerminatedString(message[6:]))

		case string(messages.ReadyForQuery):
			return queryResult, nil

		default:
			if pgConnection.isVerbose() {
				fmt.Printf("Unknown message: %s\n", string(message))
			}
		}
	}
}

func parseDataRow(answer []byte, fields []models.Field) map[string]interface{} {
	row := make(map[string]interface{})
	idxRead := 7 // Skip Header

	for _, field := range fields {
		value := parseColumnValue(answer, field, idxRead)
		row[field.Name] = value
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

		format := "binary"

		if formatCode == 0 {
			format = "text"
		}

		fields[i] = models.Field{
			Name:         fieldName,
			TableOID:     tableOID,
			AttrNum:      columnAttrNum,
			DataTypeOID:  dataTypeOID,
			Size:         dataTypeSize,
			TypeModifier: typeModifier,
			Format:       format,
		}
	}

	return fields, nil
}

func parseNumberOfFields(message []byte) uint16 {
	return binary.BigEndian.Uint16(message[5:7])
}

func parseColumnValue(answer []byte, field models.Field, idxRead int) any {
	columnValueLength := int32(binary.BigEndian.Uint32(answer[idxRead:]))
	idxRead += 4

	if columnValueLength == -1 {
		return nil
	}

	value := answer[idxRead : idxRead+int(columnValueLength)]

	switch field.Format {
	case "text":
		return string(value)
	case "binary":
		return value
	}

	return nil
}
