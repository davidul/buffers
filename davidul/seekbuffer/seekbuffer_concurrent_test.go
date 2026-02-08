package seekbuffer

import (
	"os"
	"sync"
	"testing"
	"time"
)

// TestConcurrentWrites tests concurrent write operations
func TestConcurrentWrites(t *testing.T) {
	buffer := NewEmptySeekBuffer()
	var wg sync.WaitGroup
	numGoroutines := 100
	writesPerGoroutine := 100

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < writesPerGoroutine; j++ {
				buffer.Write([]byte{byte(id)})
			}
		}(i)
	}

	wg.Wait()

	// Check total size
	expectedSize := numGoroutines * writesPerGoroutine
	if len(buffer.Bytes()) != expectedSize {
		t.Errorf("Expected buffer size %d, got %d", expectedSize, len(buffer.Bytes()))
	}
}

// TestConcurrentReads tests concurrent read operations
func TestConcurrentReads(t *testing.T) {
	data := make([]byte, 10000)
	for i := range data {
		data[i] = byte(i % 256)
	}
	buffer := NewSeekBuffer(data)

	var wg sync.WaitGroup
	numGoroutines := 50

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			dst := make([]byte, 10)
			buffer.Read(dst)
		}()
	}

	wg.Wait()
	// Test passes if no panic occurs
}

// TestConcurrentReadWrite tests concurrent reads and writes
func TestConcurrentReadWrite(t *testing.T) {
	buffer := NewSeekBuffer([]byte("Initial data"))
	var wg sync.WaitGroup
	duration := 100 * time.Millisecond
	stopTime := time.Now().Add(duration)

	// Writers
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer wg.Done()
			for time.Now().Before(stopTime) {
				buffer.Write([]byte{byte(id)})
			}
		}(i)
	}

	// Readers
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			dst := make([]byte, 10)
			for time.Now().Before(stopTime) {
				buffer.Read(dst)
			}
		}()
	}

	wg.Wait()
	// Test passes if no panic or race conditions occur
}

// TestConcurrentSeekAndRead tests concurrent seek and read operations
func TestConcurrentSeekAndRead(t *testing.T) {
	data := make([]byte, 1000)
	for i := range data {
		data[i] = byte(i % 256)
	}
	buffer := NewSeekBuffer(data)

	var wg sync.WaitGroup
	numGoroutines := 50

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(offset int) {
			defer wg.Done()
			buffer.Seek(offset * 10)
			dst := make([]byte, 10)
			buffer.Read(dst)
		}(i)
	}

	wg.Wait()
	// Test passes if no panic occurs
}

// TestConcurrentRewind tests concurrent rewind operations
func TestConcurrentRewind(t *testing.T) {
	buffer := NewSeekBuffer([]byte("Test data for rewind"))
	var wg sync.WaitGroup
	numGoroutines := 100

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			buffer.Rewind()
		}()
	}

	wg.Wait()

	// After all rewinds, offset should be 0
	if buffer.Len() != len(buffer.Bytes()) {
		t.Error("Offset not at beginning after concurrent rewinds")
	}
}

// TestConcurrentAppend tests concurrent append operations
func TestConcurrentAppend(t *testing.T) {
	buffer := NewEmptySeekBuffer()
	var wg sync.WaitGroup
	numGoroutines := 100
	appendsPerGoroutine := 100

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < appendsPerGoroutine; j++ {
				buffer.Append([]byte{byte(id)})
			}
		}(i)
	}

	wg.Wait()

	expectedSize := numGoroutines * appendsPerGoroutine
	if len(buffer.Bytes()) != expectedSize {
		t.Errorf("Expected buffer size %d, got %d", expectedSize, len(buffer.Bytes()))
	}
}

// TestConcurrentBytesAccess tests concurrent access to Bytes()
func TestConcurrentBytesAccess(t *testing.T) {
	buffer := NewSeekBuffer([]byte("Shared data"))
	var wg sync.WaitGroup
	numGoroutines := 100

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			_ = buffer.Bytes()
		}()
	}

	wg.Wait()
	// Test passes if no panic occurs
}

// TestConcurrentLen tests concurrent access to Len()
func TestConcurrentLen(t *testing.T) {
	buffer := NewSeekBuffer([]byte("Test data"))
	var wg sync.WaitGroup
	numGoroutines := 100

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			_ = buffer.Len()
		}()
	}

	wg.Wait()
	// Test passes if no panic occurs
}

// TestConcurrentReadBytes tests concurrent ReadBytes operations
func TestConcurrentReadBytes(t *testing.T) {
	data := []byte("Line1\nLine2\nLine3\nLine4\nLine5\n")
	buffer := NewSeekBuffer(data)
	var wg sync.WaitGroup
	numGoroutines := 5

	wg.Add(numGoroutines)
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			buffer.ReadBytes('\n')
		}()
	}

	wg.Wait()
	// Test passes if no panic occurs
}

// TestConcurrentFileOperations tests concurrent file operations
func TestConcurrentFileOperations(t *testing.T) {
	buffer := NewSeekBuffer([]byte("File operation data"))
	var wg sync.WaitGroup

	// Multiple concurrent SaveToFile
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer wg.Done()
			filename := "test_concurrent_" + string(rune('0'+id)) + ".txt"
			buffer.SaveToFile(filename)
		}(i)
	}

	wg.Wait()

	// Cleanup
	for i := 0; i < 10; i++ {
		filename := "test_concurrent_" + string(rune('0'+i)) + ".txt"
		_ = os.Remove(filename)
	}
}

// TestConcurrentMixedOperations tests a realistic scenario with mixed operations
func TestConcurrentMixedOperations(t *testing.T) {
	buffer := NewSeekBuffer([]byte("Initial buffer content"))
	var wg sync.WaitGroup
	duration := 200 * time.Millisecond
	stopTime := time.Now().Add(duration)

	// Writers
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func(id int) {
			defer wg.Done()
			for time.Now().Before(stopTime) {
				buffer.Write([]byte{byte(id)})
				time.Sleep(time.Microsecond)
			}
		}(i)
	}

	// Readers
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func() {
			defer wg.Done()
			dst := make([]byte, 5)
			for time.Now().Before(stopTime) {
				buffer.Read(dst)
				time.Sleep(time.Microsecond)
			}
		}()
	}

	// Seekers
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func(offset int) {
			defer wg.Done()
			for time.Now().Before(stopTime) {
				buffer.Seek(offset)
				time.Sleep(time.Microsecond)
			}
		}(i * 2)
	}

	// Bytes accessors
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func() {
			defer wg.Done()
			for time.Now().Before(stopTime) {
				_ = buffer.Bytes()
				_ = buffer.Len()
				time.Sleep(time.Microsecond)
			}
		}()
	}

	wg.Wait()
	// Test passes if no panic or race conditions occur
}

// TestRaceDetector should be run with: go test -race
func TestRaceDetector(t *testing.T) {
	buffer := NewSeekBuffer([]byte("Race detection test"))
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			buffer.Write([]byte{byte(i)})
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		dst := make([]byte, 10)
		for i := 0; i < 100; i++ {
			buffer.Read(dst)
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done
}
