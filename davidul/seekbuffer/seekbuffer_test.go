package seekbuffer

import (
	"io"
	"testing"
)

func TestNewEmptyBuffer(t *testing.T) {
	b := NewEmptySeekBuffer()
	if len(b.buffer) != 0 {
		t.Errorf("buffer should be empty, but got %v", b.buffer)
	}
	if b.offset != 0 {
		t.Errorf("offset should be 0, but got %d", b.offset)
	}
}

func TestNewBuffer(t *testing.T) {
	b := NewSeekBuffer([]byte{1, 2, 3})
	if len(b.buffer) != 3 {
		t.Errorf("buffer should have len 3, but got %d", len(b.buffer))
	}
	if b.offset != 0 {
		t.Errorf("offset should be 0, but got %d", b.offset)
	}
}
func TestRewind(t *testing.T) {
	b := NewEmptySeekBuffer()
	b.Append([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9})
	dst := make([]byte, 2)
	read, err := b.Read(dst)
	if err != nil {
		t.Errorf("error should be nil, but got %v", err)
		return
	}

	if b.offset != 2 {
		t.Errorf("offset should be 2, but got %d", b.offset)
	}
	if read != 2 {
		t.Errorf("read should be 2, but got %d", read)

	}

	dst2 := make([]byte, 3)
	_, err2 := b.Read(dst2)
	if err2 != nil {
		t.Errorf("error should be nil, but got %v", err2)
		return
	}
	if b.offset != 5 {
		t.Errorf("offset should be 5, but got %d", b.offset)
	}
	b.Rewind()
	dst3 := make([]byte, 2)
	_, err3 := b.Read(dst3)
	if err3 != nil {
		t.Errorf("error should be nil, but got %v", err3)
		return
	}
	if b.offset != 2 {
		t.Errorf("offset should be 0, but got %d", b.offset)
	}
}

func TestAppend(t *testing.T) {
	buffer := NewEmptySeekBuffer()
	buffer.Append([]byte{1, 2, 3})
	if len(buffer.buffer) != 3 {
		t.Errorf("buffer should be empty, but got %v", buffer.buffer)
	}
	if buffer.offset != 0 {
		t.Errorf("offset should be 0, but got %d", buffer.offset)
	}
}

func TestRead(t *testing.T) {
	buffer := NewSeekBuffer([]byte{1, 2, 3})
	dst := make([]byte, 2)
	n, err := buffer.Read(dst)
	if err != nil {
		t.Errorf("error should be nil, but got %v", err)
	}
	if n != 2 {
		t.Errorf("n should be 2, but got %d", n)
	}
	if buffer.offset != 2 {
		t.Errorf("offset should be 2, but got %d", buffer.offset)
	}
}

func TestRead_Bigger(t *testing.T) {
	buffer := NewSeekBuffer([]byte{1, 2, 3})
	dst := make([]byte, 12)
	n, err := buffer.Read(dst)
	if err != nil {
		t.Errorf("error should be nil, but got %v", err)
	}
	if n != 3 {
		t.Errorf("n should be 3, but got %d", n)
	}
}

func TestRead_Empty(t *testing.T) {
	buffer := NewEmptySeekBuffer()
	dst := make([]byte, 12)
	n, err := buffer.Read(dst)
	if err != io.EOF {
		t.Errorf("error should be nil, but got %v", err)
	}

	if n != 0 {
		t.Errorf("n should be 0, but got %d", n)
	}
}

func TestRead_Repeat(t *testing.T) {
	buffer := NewSeekBuffer([]byte{1, 2, 3})
	dst := make([]byte, 2)
	n, err := buffer.Read(dst)
	if err != nil {
		t.Errorf("error should be nil, but got %v", err)
	}
	if n != 2 {
		t.Errorf("n should be 2, but got %d", n)
	}
	if buffer.offset != 2 {
		t.Errorf("offset should be 2, but got %d", buffer.offset)
	}

	r, err1 := buffer.Read(dst)
	if err1 != nil {
		t.Errorf("error should be nil, but got %v", err)
	}
	if r != 1 {
		t.Errorf("n should be 1, but got %d", r)
	}

	p, e2 := buffer.Read(dst)
	if e2 != io.EOF {
		t.Errorf("error should be nil, but got %v", err)
	}
	if p != 0 {
		t.Errorf("n should be 0, but got %d", p)
	}
}

func TestBytes(t *testing.T) {
	buffer := NewSeekBuffer([]byte{1, 2, 3})
	if len(buffer.Bytes()) != 3 {
		t.Errorf("buffer should be empty, but got %v", buffer.buffer)
	}
}

func TestSeek(t *testing.T) {
	buffer := NewSeekBuffer([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9})
	buffer.Seek(2)
	dst := make([]byte, 2)
	read, err := buffer.Read(dst)
	if err != nil {
		t.Errorf("error should be nil, but got %v", err)
	}
	if read != 2 {
		t.Errorf("read should be 2, but got %d", read)
	}

	if buffer.offset != 4 {
		t.Errorf("offset should be 4, but got %d", buffer.offset)
	}

	if dst[0] != 3 {
		t.Errorf("dst[0] should be 3, but got %d", dst[0])
	}

}

func TestReadBytes(t *testing.T) {
	buffer := NewSeekBuffer([]byte{1, 2, 3, 4, 5, 6, '\n', 8, 9})
	dst, err := buffer.ReadBytes(4)
	if err != nil {
		t.Errorf("error should be nil, but got %v", err)
	}
	if len(dst) != 4 {
		t.Errorf("len should be 4, but got %d", len(dst))
	}
	if buffer.offset != 4 {
		t.Errorf("offset should be 4, but got %d", buffer.offset)
	}
	bytes, err := buffer.ReadBytes('\n')
	if err != nil {
		t.Errorf("error should be nil, but got %v", err)
	}
	if len(bytes) != 3 {
		t.Errorf("len should be 1, but got %d", len(bytes))
	}
	if buffer.offset != 7 {
		t.Errorf("offset should be 7, but got %d", buffer.offset)
	}
}

func TestReadBytes_NotFound(t *testing.T) {
	buffer := NewSeekBuffer([]byte{1, 2, 3, 4, 5, 6, '\n', 8, 9})
	b, err := buffer.ReadBytes(11)
	if err != io.EOF {
		t.Errorf("error should be nil, but got %v", err)
	}
	if buffer.offset != 9 {
		t.Errorf("offset should be 9, but got %d", buffer.offset)
	}

	if len(b) != 9 {
		t.Errorf("len should be 9, but got %d", len(b))
	}
}
