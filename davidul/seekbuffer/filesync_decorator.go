package seekbuffer

import (
	"io"
	"os"
)

// FileSyncDecorator wraps any SeekableBuffer implementation and adds file synchronization
// This decorator keeps the buffer and file perfectly synchronized:
// - Write operations write to both buffer and file
// - Seek operations update both buffer offset and file position
// - Rewind operations reset both to the beginning
type FileSyncDecorator struct {
	buffer       SeekableBuffer
	file         *os.File
	filename     string
	writtenBytes int
	syncEnabled  bool
}

// NewFileSyncDecorator creates a decorator that wraps the given buffer
// with file synchronization functionality
func NewFileSyncDecorator(buffer SeekableBuffer) *FileSyncDecorator {
	return &FileSyncDecorator{
		buffer:       buffer,
		file:         nil,
		filename:     "",
		writtenBytes: 0,
		syncEnabled:  false,
	}
}

// EnableFileSync enables synchronization with a file. All writes to the buffer
// will be automatically written to the file, and seek operations will update
// the file position as well.
func (d *FileSyncDecorator) EnableFileSync(filename string) error {
	// If already syncing to a different file, close it first
	if d.file != nil && d.filename != filename {
		d.file.Close()
		d.file = nil
	}

	// Open or create the file
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	d.file = f
	d.filename = filename
	d.syncEnabled = true
	d.writtenBytes = 0

	// Write existing buffer content to file if any
	bufferData := d.buffer.Bytes()
	if len(bufferData) > 0 {
		// Truncate file first
		if err := f.Truncate(0); err != nil {
			return err
		}
		if err := d.syncNewDataToFile(); err != nil {
			return err
		}
	}

	return nil
}

// DisableFileSync disables file synchronization. The file handle is closed
// but the buffer contents remain in memory.
func (d *FileSyncDecorator) DisableFileSync() error {
	if d.file != nil {
		err := d.file.Close()
		d.file = nil
		d.filename = ""
		d.writtenBytes = 0
		d.syncEnabled = false
		return err
	}
	d.syncEnabled = false
	return nil
}

// Write delegates to the underlying buffer and syncs to file
func (d *FileSyncDecorator) Write(src []byte) (int, error) {
	n, err := d.buffer.Write(src)
	if err != nil {
		return n, err
	}

	if d.syncEnabled && d.file != nil {
		if err := d.syncNewDataToFile(); err != nil {
			return 0, err
		}
	}

	return n, nil
}

// Append delegates to the underlying buffer and syncs to file
func (d *FileSyncDecorator) Append(src []byte) {
	d.buffer.Append(src)
	if d.syncEnabled && d.file != nil {
		d.syncNewDataToFile()
	}
}

// Read delegates to the underlying buffer
func (d *FileSyncDecorator) Read(dst []byte) (int, error) {
	return d.buffer.Read(dst)
}

// Rewind delegates to the underlying buffer and syncs file position
func (d *FileSyncDecorator) Rewind() {
	d.buffer.Rewind()
	if d.syncEnabled && d.file != nil {
		d.file.Seek(0, io.SeekStart)
	}
}

// Seek delegates to the underlying buffer and syncs file position
func (d *FileSyncDecorator) Seek(offset int) {
	d.buffer.Seek(offset)
	if d.syncEnabled && d.file != nil {
		d.file.Seek(int64(offset), io.SeekStart)
	}
}

// Bytes delegates to the underlying buffer
func (d *FileSyncDecorator) Bytes() []byte {
	return d.buffer.Bytes()
}

// Len delegates to the underlying buffer
func (d *FileSyncDecorator) Len() int {
	return d.buffer.Len()
}

// ReadBytes delegates to the underlying buffer
func (d *FileSyncDecorator) ReadBytes(c byte) ([]byte, error) {
	return d.buffer.ReadBytes(c)
}

// Close closes both the file and the underlying buffer
func (d *FileSyncDecorator) Close() error {
	var bufferErr error
	if d.buffer != nil {
		bufferErr = d.buffer.Close()
	}

	var fileErr error
	if d.file != nil {
		fileErr = d.file.Close()
		d.file = nil
		d.filename = ""
		d.writtenBytes = 0
		d.syncEnabled = false
	}

	if fileErr != nil {
		return fileErr
	}
	return bufferErr
}

// syncNewDataToFile writes any new data in the buffer that hasn't been written to file yet
func (d *FileSyncDecorator) syncNewDataToFile() error {
	if d.file == nil || !d.syncEnabled {
		return nil
	}

	bufferData := d.buffer.Bytes()
	if d.writtenBytes < len(bufferData) {
		// Save current file position
		currentPos, err := d.file.Seek(0, io.SeekCurrent)
		if err != nil {
			return err
		}

		// Seek to end of written data
		_, err = d.file.Seek(int64(d.writtenBytes), io.SeekStart)
		if err != nil {
			return err
		}

		// Write new data
		newData := bufferData[d.writtenBytes:]
		n, err := d.file.Write(newData)
		if err != nil {
			return err
		}
		d.writtenBytes += n

		// Restore file position to where it was (matches buffer offset)
		_, err = d.file.Seek(currentPos, io.SeekStart)
		if err != nil {
			return err
		}
	}

	return nil
}

// IsSyncEnabled returns true if file synchronization is currently enabled
func (d *FileSyncDecorator) IsSyncEnabled() bool {
	return d.syncEnabled
}

// GetSyncFilename returns the filename being synced to, or empty string if not syncing
func (d *FileSyncDecorator) GetSyncFilename() string {
	return d.filename
}
