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
	buffer       []byte
	offset       int
	file         *os.File // associated file handle for sync
	filename     string   // filename for the associated file
	writtenBytes int      // tracks how many bytes have been written to file
	syncEnabled  bool     // whether file sync is enabled
}

// NewEmptySeekBuffer empty buffer, offset is 0
func NewEmptySeekBuffer() *SeekBuffer {
	return &SeekBuffer{
		buffer:       make([]byte, 0),
		offset:       0,
		file:         nil,
		filename:     "",
		writtenBytes: 0,
		syncEnabled:  false,
	}
}

// NewSeekBuffer buffer with initial content. Copy src into buffer.
func NewSeekBuffer(src []byte) *SeekBuffer {
	a := make([]byte, len(src))
	copy(a, src)
	return &SeekBuffer{
		buffer:       a,
		offset:       0,
		file:         nil,
		filename:     "",
		writtenBytes: 0,
		syncEnabled:  false,
	}
}

// Bytes returns current content of the buffer
func (s *SeekBuffer) Bytes() []byte {
	return s.buffer
}

// Append appends content to the buffer
func (s *SeekBuffer) Append(src []byte) {
	s.buffer = append(s.buffer, src...)
	if s.syncEnabled && s.file != nil {
		s.syncNewDataToFile()
	}
}

// writes content to the buffer, alias for Append
func (s *SeekBuffer) Write(src []byte) (int, error) {
	s.buffer = append(s.buffer, src...)
	if s.syncEnabled && s.file != nil {
		if err := s.syncNewDataToFile(); err != nil {
			return 0, err
		}
	}
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
	if s.syncEnabled && s.file != nil {
		s.file.Seek(0, io.SeekStart)
	}
}

// Seek seeks to the offset
func (s *SeekBuffer) Seek(offset int) {
	s.offset = offset
	if s.syncEnabled && s.file != nil {
		s.file.Seek(int64(offset), io.SeekStart)
	}
}

// Close closes the buffer and associated file
func (s *SeekBuffer) Close() error {
	s.offset = 0
	s.buffer = nil
	if s.file != nil {
		err := s.file.Close()
		s.file = nil
		s.filename = ""
		s.writtenBytes = 0
		s.syncEnabled = false
		return err
	}
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

// EnableFileSync enables synchronization with a file. All writes to the buffer
// will be automatically written to the file, and seek operations will update
// the file position as well.
func (s *SeekBuffer) EnableFileSync(filename string) error {
	// If already syncing to a different file, close it first
	if s.file != nil && s.filename != filename {
		s.file.Close()
		s.file = nil
	}

	// Open or create the file
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	s.file = f
	s.filename = filename
	s.syncEnabled = true
	s.writtenBytes = 0

	// Write existing buffer content to file if any
	if len(s.buffer) > 0 {
		// Truncate file first
		if err := f.Truncate(0); err != nil {
			return err
		}
		if err := s.syncNewDataToFile(); err != nil {
			return err
		}
	}

	// Set file position to match current buffer offset
	if s.offset > 0 {
		_, err = f.Seek(int64(s.offset), io.SeekStart)
		if err != nil {
			return err
		}
	}

	return nil
}

// DisableFileSync disables file synchronization. The file handle is closed
// but the buffer contents remain in memory.
func (s *SeekBuffer) DisableFileSync() error {
	if s.file != nil {
		err := s.file.Close()
		s.file = nil
		s.filename = ""
		s.writtenBytes = 0
		s.syncEnabled = false
		return err
	}
	s.syncEnabled = false
	return nil
}

// syncNewDataToFile writes any new data in the buffer that hasn't been written to file yet
func (s *SeekBuffer) syncNewDataToFile() error {
	if s.file == nil || !s.syncEnabled {
		return nil
	}

	// Check if there's new data to write
	if s.writtenBytes < len(s.buffer) {
		// Save current file position
		currentPos, err := s.file.Seek(0, io.SeekCurrent)
		if err != nil {
			return err
		}

		// Seek to end of written data
		_, err = s.file.Seek(int64(s.writtenBytes), io.SeekStart)
		if err != nil {
			return err
		}

		// Write new data
		newData := s.buffer[s.writtenBytes:]
		n, err := s.file.Write(newData)
		if err != nil {
			return err
		}
		s.writtenBytes += n

		// Restore file position to where it was (matches buffer offset)
		_, err = s.file.Seek(currentPos, io.SeekStart)
		if err != nil {
			return err
		}
	}

	return nil
}

// IsSyncEnabled returns true if file synchronization is currently enabled
func (s *SeekBuffer) IsSyncEnabled() bool {
	return s.syncEnabled
}

// GetSyncFilename returns the filename being synced to, or empty string if not syncing
func (s *SeekBuffer) GetSyncFilename() string {
	return s.filename
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
