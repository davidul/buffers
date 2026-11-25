package plain

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWrite(t *testing.T) {
	tt := []struct {
		dst      []byte
		src      []byte
		offset   int
		expected []byte
	}{
		{Allocate(10), []byte{1, 2, 3}, 0, []byte{1, 2, 3, 0, 0, 0, 0, 0, 0, 0}},
		{Allocate(10), []byte{1, 2, 3}, 3, []byte{0, 0, 0, 1, 2, 3, 0, 0, 0, 0}},
	}

	for _, tc := range tt {
		err := Write(tc.dst, tc.src, tc.offset)
		assert.Nil(t, err)
		assert.Equal(t, tc.expected, tc.dst)
	}
}

func TestWritePanic(t *testing.T) {
	dst := Allocate(10)
	src := []byte{1, 2, 3}
	err := Write(dst, src, 10)
	assert.NotNil(t, err)
	if err != nil {
		assert.Equal(t, "not enough space in dst", err.Error())
	}
}

func TestWriteNegativeOffsetReturnsErrorWithoutPanic(t *testing.T) {
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Errorf("Write panicked for a negative offset: %v", recovered)
		}
	}()
	if err := Write(make([]byte, 4), []byte("x"), -1); err == nil {
		t.Fatal("Write should reject a negative offset")
	}
}

func TestRead(t *testing.T) {
	dst := Allocate(10)
	src := []byte{1, 2, 3}
	// Writing 1, 2, 3 to dst at offset 7
	err := Write(dst, src, 7)
	assert.Nil(t, err)

	// Reading from offset 7 should give us the bytes 1, 2, 3
	dst2 := Allocate(3)
	err = Read(dst2, dst, 7)
	assert.Nil(t, err)
	assert.Equal(t, []byte{1, 2, 3}, dst2)
}

func TestReadRange(t *testing.T) {
	dst := Allocate(10)
	src := []byte{1, 2, 3}
	err := Write(dst, src, 2)
	assert.Nil(t, err)
	dst2 := Allocate(3)
	err = ReadRange(dst2, dst, 2, 3)
	assert.Nil(t, err)
	assert.Equal(t, []byte{1, 2, 3}, dst2)
}

func TestRabinKarp(t *testing.T) {
	runes := make([]rune, 10)
	runes[0] = 'a'
	runes[1] = 'b'
	runes[2] = 'c'

	RabinKarp([]rune{'a', 'b', 'c'}, []rune{'a', 'b', 'c'})
}

func TestRollingHash(t *testing.T) {
	s := "šabcž"
	r := []rune(s)
	for _, x := range r {
		fmt.Printf("%d ", x)
	}

	b := []byte(s)
	for _, x := range b {
		fmt.Printf("%d ", x)
	}
	file, err := os.ReadFile("/Users/david/Downloads/katastr/20230831_OB_539066_UKSH.xml")
	if err != nil {
		fmt.Println(err)
	}

	s2 := string(file)

	hash := RollingHash([]rune(s2))
	fmt.Println(hash)
}
