package protocol

import (
	"encoding/binary"
	"fmt"
	"postgres-protocol-go/internal/messages"
	"postgres-protocol-go/pkg/utils"
)

func ProcessSimpleQuery(pgConnection PgConnection, query string) (string, error) {
	buf := NewWriteBuffer(1024)
	buf.StartMessage(messages.SimpleQuery)
	buf.WriteString(query)
	buf.FinishMessage()

	err := pgConnection.sendMessage(buf)
	if err != nil {
		return "", err
	}

	answer, err := pgConnection.readMessage()
	if err != nil {
		return "", err
	}

	identifier := utils.ParseIdentifier(answer)
	if identifier != string(messages.RowDescription) {
		return "", fmt.Errorf("expected RowDescription message, got %s", identifier)
	}

	numberOfFields := parseNumberOfFields(answer)
	idxRead := 7

	for i := uint16(0); i < numberOfFields; i++ {
		fieldName := utils.ExtractNullTerminatedString(answer[idxRead:])
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

		fmt.Printf("Field: %s, TableOID: %d, AttrNum: %d, DataTypeOID: %d, Size: %d, TypeModifier: %d, FormatCode: %d\n",
			fieldName, tableOID, columnAttrNum, dataTypeOID, dataTypeSize, typeModifier, formatCode)
	}

	return string(answer), nil
}

func parseNumberOfFields(message []byte) uint16 {
	return binary.BigEndian.Uint16(message[5:7])
}
