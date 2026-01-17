package randombuffer

import "testing"

func TestEmptySBuffer(t *testing.T) {
	buffer := NewEmptyRandomBuffer()
	if buffer.Len() != 0 {
		t.Errorf("buffer should be empty, but got %d", buffer.Len())
	}
	if buffer.AbsLen() != 0 {
		t.Errorf("buffer should be empty, but got %d", buffer.AbsLen())
	}
	if buffer.Cap() != 0 {
		t.Errorf("buffer should be empty, but got %d", buffer.Cap())
	}

	if buffer.ReadOffset() != 0 {
		t.Errorf("buffer should be empty, but got %d", buffer.ReadOffset())
	}
	if buffer.WriteOffset() != 0 {
		t.Errorf("buffer should be empty, but got %d", buffer.WriteOffset())
	}
}

func TestNewSBuffer(t *testing.T) {
	buffer := NewRandomBuffer([]byte{1, 2, 3})
	if buffer.Len() != 3 {
		t.Errorf("buffer should be empty, but got %d", buffer.Len())
	}

	if buffer.AbsLen() != 3 {
		t.Errorf("buffer should be empty, but got %d", buffer.AbsLen())
	}

	if buffer.ReadOffset() != 0 {
		t.Errorf("readOffset should be 0, but got %d", buffer.ReadOffset())
	}

	if buffer.WriteOffset() != 3 {
		t.Errorf("writeOffset should be 3, but got %d", buffer.WriteOffset())
	}
}

func TestNewSBufferWithCapacity(t *testing.T) {
	buffer := NewRandomBufferWithCapacity(10)
	if buffer.Len() != 0 {
		t.Errorf("buffer should be empty, but got %d", buffer.Len())
	}
	if buffer.AbsLen() != 0 {
		t.Errorf("buffer should be empty, but got %d", buffer.AbsLen())
	}
}

func TestAppend(t *testing.T) {
	buffer := NewRandomBufferWithCapacity(10)
	buffer.Append([]byte{1, 2, 3, 5, 6, 7, 8, 9, 10})
	if buffer.ReadOffset() != 0 {
		t.Errorf("readOffset should be 0, but got %d", buffer.ReadOffset())
	}

	if buffer.WriteOffset() != 9 {
		t.Errorf("writeOffset should be 9, but got %d", buffer.WriteOffset())
	}

	buffer.Append([]byte{4, 5, 6, 7})
	if buffer.WriteOffset() != 13 {
		t.Errorf("readOffset should be 13, but got %d", buffer.ReadOffset())
	}
}

func TestAppend_OverCapacity(t *testing.T) {
	buffer := NewRandomBufferWithCapacity(4)
	buffer.Append([]byte{1, 2, 3})
	if buffer.ReadOffset() != 0 {
		t.Errorf("readOffset should be 0, but got %d", buffer.ReadOffset())
	}
	if buffer.WriteOffset() != 3 {
		t.Errorf("writeOffset should be 3, but got %d", buffer.WriteOffset())
	}

	buffer.Append([]byte{4, 5, 6, 7})
	if buffer.WriteOffset() != 7 {
		t.Errorf("readOffset should be 7, but got %d", buffer.WriteOffset())
	}

	buffer2 := NewEmptyRandomBuffer()
	i := buffer2.Cap()
	if i != 0 {
		t.Errorf("cap should be 0, but got %d", i)
	}
	buffer2.Append([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

}

func TestWrite(t *testing.T) {
	buffer := NewRandomBufferWithCapacity(10)
	buffer.Write([]byte{1, 2, 3})
	if buffer.WriteOffset() != 3 {
		t.Errorf("readOffset should be 3, but got %d", buffer.WriteOffset())
	}

	buffer.Write([]byte{4, 5, 6, 7})
	if buffer.WriteOffset() != 7 {
		t.Errorf("readOffset should be 7, but got %d", buffer.WriteOffset())
	}
}

//func TestWrite_OverCapacity(t *testing.T) {
//	buffer := NewRandomBufferWithCapacity(4)
//	buffer.Write([]byte{1, 2, 3})
//	if buffer.ReadOffset() != 3 {
//		t.Errorf("readOffset should be 3, but got %d", buffer.ReadOffset())
//	}
//
//	buffer.Write([]byte{4, 5, 6, 7})
//	if buffer.ReadOffset() != 7 {
//		t.Errorf("readOffset should be 7, but got %d", buffer.ReadOffset())
//	}
//}
//
//func TestWrite_Long(t *testing.T) {
//	buffer := NewRandomBufferWithCapacity(3)
//	for i := 0; i < 1024; i++ {
//		buffer.Write([]byte{1, 2, 3})
//	}
//	if buffer.ReadOffset() != 3072 {
//		t.Errorf("readOffset should be 3072, but got %d", buffer.ReadOffset())
//	}
//
//	if buffer.AbsLen() != 3072 {
//		t.Errorf("len should be 3072, but got %d", buffer.AbsLen())
//	}
//}

func TestWrite_Random(t *testing.T) {

	tt := []struct {
		writeOffset int
		data        []byte
	}{
		{3, []byte{1, 2, 3}},
		{7, []byte{4, 5, 6, 7}},
		{12, []byte{8, 9, 10, 11, 12}},
	}

	buffer := NewRandomBufferWithCapacity(3)

	for _, tc := range tt {
		buffer.Write(tc.data)
		if buffer.WriteOffset() != tc.writeOffset {
			t.Errorf("readOffset should be %d, but got %d", tc.writeOffset, buffer.WriteOffset())
		}

	}
}

//
//func TestRewind(t *testing.T) {
//	buffer := NewRandomBufferWithCapacity(3)
//	buffer.Write([]byte{1, 2, 3})
//	dst := make([]byte, 2)
//	buffer.Read(dst)
//}
