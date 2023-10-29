package seekbuffer

import (
	"bytes"
	"io"
)

// byte buffer and pointer to the current offset
type SeekBuffer struct {
	buffer []byte
	offset int
}

// empty buffer
func NewEmptySeekBuffer() *SeekBuffer {
	return &SeekBuffer{
		buffer: make([]byte, 0),
		offset: 0,
	}
}

// buffer with initial content
func NewSeekBuffer(src []byte) *SeekBuffer {
	a := make([]byte, len(src))
	copy(a, src)
	return &SeekBuffer{
		buffer: a,
		offset: 0,
	}
}

// returns current content of the buffer
func (s *SeekBuffer) Bytes() []byte {
	return s.buffer
}

// appends content to the buffer
func (s *SeekBuffer) Append(src []byte) {
	s.buffer = append(s.buffer, src...)
}

// writes content to the buffer, alias for Append
func (s *SeekBuffer) Write(src []byte) (int, error) {
	s.buffer = append(s.buffer, src...)
	return len(src), nil
}

// reads content from the buffer into dst
func (s *SeekBuffer) Read(dst []byte) (int, error) {
	if s.offset >= len(s.buffer) {
		return 0, io.EOF
	}

	n := copy(dst, s.buffer[s.offset:])
	s.offset += n
	return n, nil
}

// rewinds the buffer to the beginning
func (s *SeekBuffer) Rewind() {
	s.offset = 0
}

// seeks to the offset
func (s *SeekBuffer) Seek(offset int) {
	s.offset = offset
}

// closes the buffer
func (s *SeekBuffer) Close() error {
	s.offset = 0
	s.buffer = nil
	return nil
}

// read bytes up to the first occurrence of c
func (s *SeekBuffer) ReadBytes(c byte) ([]byte, error) {
	indexByte := bytes.IndexByte(s.buffer[s.offset:], c)
	if indexByte == -1 {
		s.offset = len(s.buffer)
		return s.buffer[s.offset:], io.EOF
	}
	end := s.offset + indexByte + 1
	b := s.buffer[s.offset:end]
	s.offset = end
	return b, nil
}

func (s *SeekBuffer) Len() int {
	return len(s.buffer) - s.offset
}
