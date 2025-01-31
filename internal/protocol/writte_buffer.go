package protocol

import (
	"encoding/binary"
)

type WriteBuffer struct {
	Bytes            []byte
	IsStartupMessage bool

	msgStart int
}

func NewWriteBuffer(bufSize int) *WriteBuffer {
	return &WriteBuffer{
		Bytes:            make([]byte, 0, bufSize),
		IsStartupMessage: false,
	}
}

func (buf *WriteBuffer) Reset() {
	buf.Bytes = buf.Bytes[:0]
}

func (buf *WriteBuffer) StartMessage(c byte) {
	if c == 0 {
		buf.IsStartupMessage = true
		buf.msgStart = len(buf.Bytes)
		buf.Bytes = append(buf.Bytes, 0, 0, 0, 0)
	} else {
		buf.msgStart = len(buf.Bytes) + 1
		buf.Bytes = append(buf.Bytes, c, 0, 0, 0, 0)
	}
}

func (buf *WriteBuffer) FinishMessage() {
	binary.BigEndian.PutUint32(
		buf.Bytes[buf.msgStart:], uint32(len(buf.Bytes)-buf.msgStart))
}

func (buf *WriteBuffer) Write(b []byte) (int, error) {
	buf.Bytes = append(buf.Bytes, b...)
	return len(b), nil
}

func (buf *WriteBuffer) WriteInt16(num int16) {
	buf.Bytes = append(buf.Bytes, 0, 0)
	binary.BigEndian.PutUint16(buf.Bytes[len(buf.Bytes)-2:], uint16(num))
}

func (buf *WriteBuffer) WriteInt32(num int32) {
	buf.Bytes = append(buf.Bytes, 0, 0, 0, 0)
	binary.BigEndian.PutUint32(buf.Bytes[len(buf.Bytes)-4:], uint32(num))
}

func (buf *WriteBuffer) WriteString(s string) {
	buf.Bytes = append(buf.Bytes, s...)
	buf.Bytes = append(buf.Bytes, 0)
}

func (buf *WriteBuffer) WriteBytes(b []byte) {
	buf.Bytes = append(buf.Bytes, b...)
	buf.Bytes = append(buf.Bytes, 0)
}

func (buf *WriteBuffer) WriteByte(c byte) error {
	buf.Bytes = append(buf.Bytes, c)
	return nil
}
