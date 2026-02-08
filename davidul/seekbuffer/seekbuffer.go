package seekbuffer

import (
	"bytes"
	"io"
	"os"
)

// Package seekbuffer provides a SeekBuffer that implements io.Reader and io.Writer interfaces.
// It allows reading and writing bytes with a seekable offset, similar to a file.
// It is useful for scenarios where you need to read and write data in a buffered manner,
// while keeping track of the current position in the buffer, such as in network protocols or file processing.
// It is not thread-safe, so it should be used in a single goroutine or with proper synchronization.

// SeekBuffer byte buffer and pointer to the current offset
type SeekBuffer struct {
	buffer []byte
	offset int
}

// NewEmptySeekBuffer empty buffer, offset is 0
func NewEmptySeekBuffer() *SeekBuffer {
	return &SeekBuffer{
		buffer: make([]byte, 0),
		offset: 0,
	}
}

// NewSeekBuffer buffer with initial content. Copy src into buffer.
func NewSeekBuffer(src []byte) *SeekBuffer {
	a := make([]byte, len(src))
	copy(a, src)
	return &SeekBuffer{
		buffer: a,
		offset: 0,
	}
}

// Bytes returns current content of the buffer
func (s *SeekBuffer) Bytes() []byte {
	return s.buffer
}

// Append appends content to the buffer
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

// Rewind rewinds the buffer to the beginning
func (s *SeekBuffer) Rewind() {
	s.offset = 0
}

// Seek seeks to the offset
func (s *SeekBuffer) Seek(offset int) {
	s.offset = offset
}

// Close closes the buffer
func (s *SeekBuffer) Close() error {
	s.offset = 0
	s.buffer = nil
	return nil
}

// ReadBytes read bytes up to the first occurrence of c
func (s *SeekBuffer) ReadBytes(c byte) ([]byte, error) {
	indexByte := bytes.IndexByte(s.buffer[s.offset:], c)
	if indexByte == -1 {
		b := s.buffer[s.offset:]
		s.offset = len(s.buffer)
		return b, io.EOF
	}
	end := s.offset + indexByte + 1
	b := s.buffer[s.offset:end]
	s.offset = end
	return b, nil
}

func (s *SeekBuffer) Len() int {
	return len(s.buffer) - s.offset
}

// SaveToFile writes the buffer content to a file (overwrites if exists)
func (s *SeekBuffer) SaveToFile(filename string) error {
	return os.WriteFile(filename, s.buffer, 0644)
}

// AppendToFile appends the buffer content to an existing file or creates a new one
func (s *SeekBuffer) AppendToFile(filename string) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(s.buffer)
	return err
}

// AppendUnreadToFile appends only the unread portion of the buffer (from offset to end) to a file
func (s *SeekBuffer) AppendUnreadToFile(filename string) error {
	if s.offset >= len(s.buffer) {
		return nil // Nothing to append
	}

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(s.buffer[s.offset:])
	return err
}

// LoadFromFile reads the file content into the buffer and resets the offset to 0
func (s *SeekBuffer) LoadFromFile(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	s.buffer = data
	s.offset = 0
	return nil
}

// NewSeekBufferFromFile creates a new SeekBuffer by reading from a file
func NewSeekBufferFromFile(filename string) (*SeekBuffer, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &SeekBuffer{
		buffer: data,
		offset: 0,
	}, nil
}
