package randombuffer

import "fmt"

// RandomBuffer is a simple buffer implementation that allows for random access
// to the data. It implements a basic byte buffer with read and write capabilities.
// It is not thread-safe, so it should be used in a single goroutine or with proper synchronization.
type RandomBuffer struct {
	buffer      []byte
	readOffset  int
	writeOffset int
}

// NewEmptySBuffer creates a new empty RandomBuffer.
func NewEmptySBuffer() *RandomBuffer {
	return &RandomBuffer{
		buffer:      make([]byte, 0),
		readOffset:  0,
		writeOffset: 0,
	}
}

// NewSBufferWithCapacity creates a new RandomBuffer with the specified capacity.
func NewSBufferWithCapacity(capacity int) *RandomBuffer {
	return &RandomBuffer{
		buffer:      make([]byte, 0, capacity),
		readOffset:  0,
		writeOffset: 0,
	}
}

// NewSBuffer creates a new RandomBuffer initialized with the given byte slice.
func NewSBuffer(buffer []byte) *RandomBuffer {
	return &RandomBuffer{
		buffer:      buffer,
		writeOffset: len(buffer),
		readOffset:  0,
	}
}

// Append appends the given byte slice to the end of the buffer.
func (s *RandomBuffer) Append(src []byte) {
	s.buffer = append(s.buffer, src...)
	s.writeOffset += len(src)
}

// Write writes the given byte slice to the buffer at the current write offset.
// If the buffer is not large enough, it will be resized to accommodate the new data.
func (s *RandomBuffer) Write(src []byte) {
	if len(s.buffer) == 0 {
		s.buffer = s.buffer[:cap(s.buffer)]
	}

	if cap(s.buffer) < len(src)+s.writeOffset {
		n := len(s.buffer) + len(src)
		a := make([]byte, n, n)
		copy(a, s.buffer)
		s.buffer = a
	}

	if len(s.buffer) < len(src)+s.writeOffset {
		s.buffer = s.buffer[:cap(s.buffer)]
	}

	t := s.buffer[s.writeOffset : s.writeOffset+len(src)]
	copy(t, src)

	s.writeOffset += len(src)
}

// Read reads bytes from the buffer into the provided destination slice starting at the current read offset.
func (s *RandomBuffer) Read(dst []byte) error {
	if s.readOffset < 0 || s.readOffset > len(s.buffer) {
		return fmt.Errorf("read offset %d is out of bounds for buffer of length %d", s.readOffset, len(s.buffer))
	}

	if s.readOffset+len(dst) > len(s.buffer) {
		return fmt.Errorf("not enough bytes to read: need %d bytes but only %d available", len(dst), len(s.buffer)-s.readOffset)
	}

	copy(dst, s.buffer[s.readOffset:s.readOffset+len(dst)])
	s.readOffset += len(dst)
	return nil
}

// Rewind resets the read offset to the beginning of the buffer.
func (s *RandomBuffer) Rewind() {
	s.readOffset = 0
}

// Seek sets the read offset to the specified position.
// It returns an error if the offset is out of bounds.
func (s *RandomBuffer) Seek(offset int) error {
	if offset < 0 {
		return fmt.Errorf("seek offset cannot be negative: %d", offset)
	}
	if offset > len(s.buffer) {
		return fmt.Errorf("seek offset %d exceeds buffer length %d", offset, len(s.buffer))
	}
	s.readOffset = offset
	return nil
}

// AbsLen returns the absolute length of the buffer.
func (s *RandomBuffer) AbsLen() int {
	return len(s.buffer)
}

// Len returns the number of bytes remaining to be read from the current read offset to the end of the buffer.
func (s *RandomBuffer) Len() int {
	if s.readOffset > len(s.buffer) {
		return 0
	}
	return len(s.buffer[s.readOffset:])
}

// ReadOffset returns the current read offset.
func (s *RandomBuffer) ReadOffset() int {
	return s.readOffset
}

// WriteOffset returns the current write offset.
func (s *RandomBuffer) WriteOffset() int {
	return s.writeOffset
}

// Bytes returns the unread bytes from the current read offset to the end of the buffer.
func (s *RandomBuffer) Bytes() []byte {
	if s.readOffset > len(s.buffer) {
		return []byte{}
	}
	return s.buffer[s.readOffset:]
}

// Cap returns the capacity of the underlying buffer.
func (s *RandomBuffer) Cap() int {
	return cap(s.buffer)
}
