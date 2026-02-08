package seekbuffer

import "io"

// SeekableBuffer defines the interface that any buffer must implement to be wrapped
// with file sync functionality or other decorators
type SeekableBuffer interface {
	io.Reader
	io.Writer
	io.Closer
	Bytes() []byte
	Append(src []byte)
	Rewind()
	Seek(offset int)
	Len() int
	ReadBytes(c byte) ([]byte, error)
}
