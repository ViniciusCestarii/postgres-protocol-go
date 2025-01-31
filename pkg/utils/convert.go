package utils

import (
	"encoding/binary"
)

func StringToBytes(s string) []byte {
	return append([]byte(s), 0) // Add a null byte at the end
}

func Int32ToBytes(n int32) []byte {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(n))
	return buf
}
