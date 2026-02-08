package seekbuffer

import (
	"fmt"
	"io"
	"log"
	"time"
)

// LoggingDecorator wraps any SeekableBuffer implementation and adds logging.
// It logs all operations performed on the buffer including reads, writes, seeks, etc.
type LoggingDecorator struct {
	buffer SeekableBuffer
	logger *log.Logger
	name   string
}

// NewLoggingDecorator creates a decorator that wraps the given buffer with logging
func NewLoggingDecorator(buffer SeekableBuffer, logger *log.Logger, name string) *LoggingDecorator {
	if logger == nil {
		logger = log.Default()
	}
	if name == "" {
		name = "SeekBuffer"
	}
	return &LoggingDecorator{
		buffer: buffer,
		logger: logger,
		name:   name,
	}
}

// Write logs the write operation and delegates to the underlying buffer
func (d *LoggingDecorator) Write(p []byte) (int, error) {
	start := time.Now()
	n, err := d.buffer.Write(p)
	duration := time.Since(start)

	if err != nil {
		d.logger.Printf("[%s] Write: %d bytes, error: %v, duration: %v", d.name, n, err, duration)
	} else {
		d.logger.Printf("[%s] Write: %d bytes, duration: %v", d.name, n, duration)
	}

	return n, err
}

// Read logs the read operation and delegates to the underlying buffer
func (d *LoggingDecorator) Read(p []byte) (int, error) {
	start := time.Now()
	n, err := d.buffer.Read(p)
	duration := time.Since(start)

	if err != nil && err != io.EOF {
		d.logger.Printf("[%s] Read: %d bytes, error: %v, duration: %v", d.name, n, err, duration)
	} else if err == io.EOF {
		d.logger.Printf("[%s] Read: %d bytes, EOF, duration: %v", d.name, n, duration)
	} else {
		d.logger.Printf("[%s] Read: %d bytes, duration: %v", d.name, n, duration)
	}

	return n, err
}

// Append logs the append operation and delegates to the underlying buffer
func (d *LoggingDecorator) Append(src []byte) {
	start := time.Now()
	d.buffer.Append(src)
	duration := time.Since(start)

	d.logger.Printf("[%s] Append: %d bytes, duration: %v", d.name, len(src), duration)
}

// Bytes logs the bytes retrieval and delegates to the underlying buffer
func (d *LoggingDecorator) Bytes() []byte {
	start := time.Now()
	result := d.buffer.Bytes()
	duration := time.Since(start)

	d.logger.Printf("[%s] Bytes: retrieved %d bytes, duration: %v", d.name, len(result), duration)

	return result
}

// Rewind logs the rewind operation and delegates to the underlying buffer
func (d *LoggingDecorator) Rewind() {
	start := time.Now()
	d.buffer.Rewind()
	duration := time.Since(start)

	d.logger.Printf("[%s] Rewind: offset reset to 0, duration: %v", d.name, duration)
}

// Seek logs the seek operation and delegates to the underlying buffer
func (d *LoggingDecorator) Seek(offset int) {
	start := time.Now()
	d.buffer.Seek(offset)
	duration := time.Since(start)

	d.logger.Printf("[%s] Seek: moved to offset %d, duration: %v", d.name, offset, duration)
}

// Len logs the length query and delegates to the underlying buffer
func (d *LoggingDecorator) Len() int {
	start := time.Now()
	length := d.buffer.Len()
	duration := time.Since(start)

	d.logger.Printf("[%s] Len: %d unread bytes, duration: %v", d.name, length, duration)

	return length
}

// ReadBytes logs the read bytes operation and delegates to the underlying buffer
func (d *LoggingDecorator) ReadBytes(c byte) ([]byte, error) {
	start := time.Now()
	result, err := d.buffer.ReadBytes(c)
	duration := time.Since(start)

	if err != nil && err != io.EOF {
		d.logger.Printf("[%s] ReadBytes: delimiter='%c' (0x%02x), read %d bytes, error: %v, duration: %v",
			d.name, c, c, len(result), err, duration)
	} else if err == io.EOF {
		d.logger.Printf("[%s] ReadBytes: delimiter='%c' (0x%02x), read %d bytes, EOF, duration: %v",
			d.name, c, c, len(result), duration)
	} else {
		d.logger.Printf("[%s] ReadBytes: delimiter='%c' (0x%02x), read %d bytes, duration: %v",
			d.name, c, c, len(result), duration)
	}

	return result, err
}

// Close logs the close operation and delegates to the underlying buffer
func (d *LoggingDecorator) Close() error {
	start := time.Now()
	err := d.buffer.Close()
	duration := time.Since(start)

	if err != nil {
		d.logger.Printf("[%s] Close: error: %v, duration: %v", d.name, err, duration)
	} else {
		d.logger.Printf("[%s] Close: success, duration: %v", d.name, duration)
	}

	return err
}

// SetLogger updates the logger used by this decorator
func (d *LoggingDecorator) SetLogger(logger *log.Logger) {
	if logger != nil {
		d.logger = logger
	}
}

// SetName updates the name prefix used in log messages
func (d *LoggingDecorator) SetName(name string) {
	if name != "" {
		d.name = name
	}
}

// GetLogger returns the current logger
func (d *LoggingDecorator) GetLogger() *log.Logger {
	return d.logger
}

// LogSummary logs a summary of the current buffer state
func (d *LoggingDecorator) LogSummary() {
	totalBytes := len(d.buffer.Bytes())
	unreadBytes := d.buffer.Len()
	readBytes := totalBytes - unreadBytes

	d.logger.Printf("[%s] Summary: total=%d bytes, read=%d bytes, unread=%d bytes",
		d.name, totalBytes, readBytes, unreadBytes)
}

// LogWithMessage logs a custom message with the buffer name
func (d *LoggingDecorator) LogWithMessage(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	d.logger.Printf("[%s] %s", d.name, message)
}
