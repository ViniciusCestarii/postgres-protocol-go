package utils

// Length of message (including length itself = 4)
func AppendMessageLength(message []byte) []byte {
	return append(Int32ToBytes(int32(len(message)+4)), message...)
}
