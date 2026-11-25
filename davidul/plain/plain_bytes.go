package plain

import "fmt"

func Allocate(n int) []byte {
	b := make([]byte, n) // n bytes, slice
	return b
}

func AllocateWithCap(n int) []byte {
	b := make([]byte, 0, n) // n bytes, slice
	return b
}

// Write writes src to dst at the specified offset.
func Write(dst []byte, src []byte, offset int) error {
	if offset < 0 {
		return fmt.Errorf("offset %d cannot be negative", offset)
	}
	if len(dst) < len(src)+offset {
		return fmt.Errorf("not enough space in dst")
	}

	for i := range src {
		dst[i+offset] = src[i]
	}
	return nil
}

// Read reads bytes from src into dst starting at the specified offset.
func Read(dst []byte, src []byte, offset int) error {
	if offset < 0 || offset >= len(src) {
		return fmt.Errorf("offset %d is out of bounds for source of length %d", offset, len(src))
	}

	if len(dst) > len(src)-offset {
		return fmt.Errorf("not enough bytes in src: dst needs %d bytes but only %d available from offset %d", len(dst), len(src)-offset, offset)
	}

	for i := range dst {
		dst[i] = src[i+offset]
	}
	return nil
}

// ReadRange reads a specific range of bytes from src into dst.
func ReadRange(dst []byte, src []byte, offset int, length int) error {
	if offset < 0 || offset >= len(src) {
		return fmt.Errorf("offset %d is out of bounds for source of length %d", offset, len(src))
	}

	if offset+length > len(src) {
		return fmt.Errorf("range [%d:%d] exceeds source length %d", offset, offset+length, len(src))
	}

	if len(dst) < length {
		return fmt.Errorf("not enough space in dst: needs %d bytes but has %d", length, len(dst))
	}

	for i := 0; i < length; i++ {
		dst[i] = src[i+offset]
	}
	return nil
}

func RollingHash(value []rune) int32 {
	const p int64 = 31
	const m int64 = 1_000_000_009
	var hash_value int64 = 0
	var p_pow int64 = 1

	for i := 0; i < len(value); i++ {
		val := int64(value[i])
		hash_value = (hash_value + (val%m)*p_pow%m) % m
		p_pow = (p_pow * p) % m
	}

	return int32(hash_value)
}

func RabinKarp(pattern []rune, t []rune) {
	p := int32(31)
	m := int32(1_000_000_007)
	pHash := make([]int32, len(pattern)+1)
	pHash[0] = 1

	// Precompute pHash
	for i := 1; i < len(pHash); i++ {
		pHash[i] = (pHash[i-1] * p) % m
	}

	// Precompute tHash
	tHash := make([]int32, len(t)+1)
	for i := 0; i < len(t); i++ {
		tHash[i+1] = (tHash[i] + (t[i]-'a'+1)*pHash[i]) % m
	}

	h_s := int32(0)
	for i := 0; i < len(pattern); i++ {
		h_s = (h_s + (pattern[i]-'a'+1)*pHash[i]) % m
	}

	for i := 0; i+len(pattern) <= len(t); i++ {
		curHash := (tHash[i+len(pattern)] - tHash[i] + m) % m
		if curHash == h_s*pHash[i]%m {
			fmt.Println(i)
		}
	}
}

//func RabinKarp(pattern []byte, t []byte) {
//	p := 31
//	m := 1_000_000_007
//	pHash := make([]int, len(pattern)+1)
//	pHash[0] = 1
//
//	for i := 1; i < len(pHash); i++ {
//		pHash[i] = (pHash[i-1] * p) % m
//	}
//
//	tHash := make([]int, len(t)+1)
//	for i := 0; i < len(t); i++ {
//		x := (int(t[i]) * pHash[i])
//		fmt.Println(x)
//		tHash[i+1] = (tHash[i] + (int(t[i]) * pHash[i])) % m
//	}
//}
