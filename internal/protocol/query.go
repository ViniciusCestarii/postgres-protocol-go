package protocol

import (
	"encoding/binary"
	"fmt"
	"postgres-protocol-go/internal/protocol/messages"
	"postgres-protocol-go/pkg/models"
	"postgres-protocol-go/pkg/utils"
	"strings"
)

func processQueryResult(pgConnection PgConnection, query string) (*models.QueryResult, error) {
	queryResult := &models.QueryResult{
		Command: strings.Fields(query)[0], // First word of the query
		Rows:    make([]map[string]interface{}, 0),
	}

	var fields []models.Field

	for {
		message, err := pgConnection.readMessage()
		if err != nil {
			return nil, err
		}

		identifier := utils.ParseIdentifier(message)

		switch identifier {
		case messages.RowDescription:
			fields, err = parseField(message)
			if err != nil {
				return nil, err
			}
			queryResult.Fields = fields

		case messages.DataRow:
			row := parseDataRow(message, fields)
			queryResult.Rows = append(queryResult.Rows, row)

		case messages.CommandComplete:
			queryResult.RowCount = len(queryResult.Rows)

			return queryResult, nil

		case messages.Error:
			identifierFieldType := string(message[5:6])
			if identifierFieldType == "0" {
				return nil, fmt.Errorf("PostgreSQL error: %s", utils.ParseNullTerminatedString(message[6:]))
			}
			return nil, fmt.Errorf("PostgreSQL error: %s: %s", identifierFieldType, utils.ParseNullTerminatedString(message[6:]))

		case messages.Notice:
			identifierFieldType := string(message[5:6])
			if pgConnection.isVerbose() {
				continue
			}

			if identifierFieldType == "0" {
				fmt.Printf("PostgreSQL notice: %s", utils.ParseNullTerminatedString(message[6:]))
			}

			fmt.Printf("PostgreSQL notice: %s: %s", identifierFieldType, utils.ParseNullTerminatedString(message[6:]))

		case messages.ReadyForQuery:
			continue

		default:
			if pgConnection.isVerbose() {
				fmt.Printf("Query: Unknown message: %s\n", string(message))
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

func parseField(answer []byte) ([]models.Field, error) {
	identifier := utils.ParseIdentifier(answer)
	if identifier != messages.RowDescription {
		return nil, fmt.Errorf("expected RowDescription message, got %s", string(identifier))
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
