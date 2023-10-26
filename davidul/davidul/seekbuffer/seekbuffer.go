package seekbuffer

import "io"

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

func NewSeekBuffer(src []byte) *SeekBuffer {
	a := make([]byte, len(src))
	copy(a, src)
	return &SeekBuffer{
		buffer: a,
		offset: 0,
	}
}

func (s *SeekBuffer) Bytes() []byte {
	return s.buffer
}

func (s *SeekBuffer) Append(src []byte) {
	s.buffer = append(s.buffer, src...)
}

func (s *SeekBuffer) Write(src []byte) {
	s.buffer = append(s.buffer, src...)
}

func (s *SeekBuffer) Read(dst []byte) (int, error) {

	if s.offset >= len(s.buffer) {
		return 0, io.EOF
	}

	n := copy(dst, s.buffer[s.offset:])

	//var i = 0
	//for ; i < len(dst) && i+s.offset < len(s.buffer); i++ {
	//	dst[i] = s.buffer[i+s.offset]
	//}
	s.offset += n //len(dst)
	return n, nil
}

func (s *SeekBuffer) Rewind() {
	s.offset = 0
}

func (s *SeekBuffer) Seek(offset int) {
	s.offset = offset
}

func (s *SeekBuffer) Close() error {
	s.offset = 0
	s.buffer = nil
	return nil
}
