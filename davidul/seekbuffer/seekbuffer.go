package seekbuffer

import (
	"bytes"
	"io"
	"os"
	"sync"
)

// Package seekbuffer provides a SeekBuffer that implements io.Reader and io.Writer interfaces.
// It allows reading and writing bytes with a seekable offset, similar to a file.
// It is useful for scenarios where you need to read and write data in a buffered manner,
// while keeping track of the current position in the buffer, such as in network protocols or file processing.
// It is thread-safe and can be safely used from multiple goroutines.

// SeekBuffer byte buffer and pointer to the current offset
// All operations are protected by a RWMutex for thread-safety
type SeekBuffer struct {
	mu     sync.RWMutex
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
// Thread-safe: uses read lock
func (s *SeekBuffer) Bytes() []byte {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.buffer
}

// Append appends content to the buffer
// Thread-safe: uses write lock
func (s *SeekBuffer) Append(src []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.buffer = append(s.buffer, src...)
}

// Write writes content to the buffer
// Thread-safe: uses write lock
func (s *SeekBuffer) Write(src []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.buffer = append(s.buffer, src...)
	return len(src), nil
}

// Read reads content from the buffer into dst
// Thread-safe: uses write lock (modifies offset)
func (s *SeekBuffer) Read(dst []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.offset >= len(s.buffer) {
		return 0, io.EOF
	}

	n := copy(dst, s.buffer[s.offset:])
	s.offset += n
	return n, nil
}

// Rewind rewinds the buffer to the beginning
// Thread-safe: uses write lock
func (s *SeekBuffer) Rewind() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.offset = 0
}

// Seek seeks to the offset
// Thread-safe: uses write lock
func (s *SeekBuffer) Seek(offset int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.offset = offset
}

// Close closes the buffer
// Thread-safe: uses write lock
func (s *SeekBuffer) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.offset = 0
	s.buffer = nil
	return nil
}

// ReadBytes read bytes up to the first occurrence of c
// Thread-safe: uses write lock (modifies offset)
func (s *SeekBuffer) ReadBytes(c byte) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

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

// Len returns the number of unread bytes in the buffer
// Thread-safe: uses read lock
func (s *SeekBuffer) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.buffer) - s.offset
}

// SaveToFile writes the buffer content to a file (overwrites if exists)
// Thread-safe: uses read lock
func (s *SeekBuffer) SaveToFile(filename string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return os.WriteFile(filename, s.buffer, 0644)
}

// AppendToFile appends the buffer content to an existing file or creates a new one
// Thread-safe: uses read lock
func (s *SeekBuffer) AppendToFile(filename string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(s.buffer)
	return err
}

// AppendUnreadToFile appends only the unread portion of the buffer (from offset to end) to a file
// Thread-safe: uses read lock
func (s *SeekBuffer) AppendUnreadToFile(filename string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

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
// Thread-safe: uses write lock
func (s *SeekBuffer) LoadFromFile(filename string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

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
