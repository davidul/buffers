package seekbuffer

import (
	"io"
	"os"
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

// TestSaveToFile tests saving buffer content to a file
func TestSaveToFile(t *testing.T) {
	buffer := NewSeekBuffer([]byte("Hello, World!"))
	filename := "test_save.txt"
	defer os.Remove(filename)

	err := buffer.SaveToFile(filename)
	if err != nil {
		t.Errorf("SaveToFile failed: %v", err)
	}

	// Verify the file was created and contains the correct data
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Errorf("Failed to read file: %v", err)
	}

	if string(data) != "Hello, World!" {
		t.Errorf("Expected 'Hello, World!', got '%s'", string(data))
	}
}

// TestLoadFromFile tests loading buffer content from a file
func TestLoadFromFile(t *testing.T) {
	filename := "test_load.txt"
	testData := []byte("Test data from file")

	// Create a test file
	err := os.WriteFile(filename, testData, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(filename)

	// Load from file
	buffer := NewEmptySeekBuffer()
	err = buffer.LoadFromFile(filename)
	if err != nil {
		t.Errorf("LoadFromFile failed: %v", err)
	}

	if string(buffer.Bytes()) != string(testData) {
		t.Errorf("Expected '%s', got '%s'", string(testData), string(buffer.Bytes()))
	}

	if buffer.offset != 0 {
		t.Errorf("Expected offset 0, got %d", buffer.offset)
	}
}

// TestNewSeekBufferFromFile tests creating a new buffer from a file
func TestNewSeekBufferFromFile(t *testing.T) {
	filename := "test_new_from_file.txt"
	testData := []byte("Content loaded directly")

	// Create a test file
	err := os.WriteFile(filename, testData, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(filename)

	// Create buffer from file
	buffer, err := NewSeekBufferFromFile(filename)
	if err != nil {
		t.Errorf("NewSeekBufferFromFile failed: %v", err)
	}

	if string(buffer.Bytes()) != string(testData) {
		t.Errorf("Expected '%s', got '%s'", string(testData), string(buffer.Bytes()))
	}

	if buffer.offset != 0 {
		t.Errorf("Expected offset 0, got %d", buffer.offset)
	}
}

// TestSaveLoadRoundtrip tests saving and loading a buffer
func TestSaveLoadRoundtrip(t *testing.T) {
	filename := "test_roundtrip.txt"
	defer os.Remove(filename)

	// Create original buffer with data
	original := NewSeekBuffer([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})

	// Move offset to simulate reading
	original.Seek(5)

	// Save to file (should save entire buffer)
	err := original.SaveToFile(filename)
	if err != nil {
		t.Errorf("SaveToFile failed: %v", err)
	}

	// Load into new buffer
	loaded := NewEmptySeekBuffer()
	err = loaded.LoadFromFile(filename)
	if err != nil {
		t.Errorf("LoadFromFile failed: %v", err)
	}

	// Verify data matches
	if len(loaded.Bytes()) != len(original.Bytes()) {
		t.Errorf("Expected length %d, got %d", len(original.Bytes()), len(loaded.Bytes()))
	}

	for i := 0; i < len(original.Bytes()); i++ {
		if original.Bytes()[i] != loaded.Bytes()[i] {
			t.Errorf("Data mismatch at index %d: expected %d, got %d",
				i, original.Bytes()[i], loaded.Bytes()[i])
		}
	}

	// Verify offset was reset on load
	if loaded.offset != 0 {
		t.Errorf("Expected offset 0 after load, got %d", loaded.offset)
	}
}

// TestLoadFromFile_NonExistent tests loading from a non-existent file
func TestLoadFromFile_NonExistent(t *testing.T) {
	buffer := NewEmptySeekBuffer()
	err := buffer.LoadFromFile("non_existent_file.txt")
	if err == nil {
		t.Errorf("Expected error when loading non-existent file, got nil")
	}
}

// TestNewSeekBufferFromFile_NonExistent tests creating buffer from non-existent file
func TestNewSeekBufferFromFile_NonExistent(t *testing.T) {
	_, err := NewSeekBufferFromFile("non_existent_file.txt")
	if err == nil {
		t.Errorf("Expected error when loading non-existent file, got nil")
	}
}

// TestAppendToFile tests appending buffer content to a file
func TestAppendToFile(t *testing.T) {
	filename := "test_append.txt"
	defer os.Remove(filename)

	// Create initial file with some data
	buffer1 := NewSeekBuffer([]byte("First line\n"))
	err := buffer1.SaveToFile(filename)
	if err != nil {
		t.Errorf("SaveToFile failed: %v", err)
	}

	// Append more data
	buffer2 := NewSeekBuffer([]byte("Second line\n"))
	err = buffer2.AppendToFile(filename)
	if err != nil {
		t.Errorf("AppendToFile failed: %v", err)
	}

	// Append even more data
	buffer3 := NewSeekBuffer([]byte("Third line\n"))
	err = buffer3.AppendToFile(filename)
	if err != nil {
		t.Errorf("AppendToFile failed: %v", err)
	}

	// Verify the file contains all data
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Errorf("Failed to read file: %v", err)
	}

	expected := "First line\nSecond line\nThird line\n"
	if string(data) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, string(data))
	}
}

// TestAppendToFile_NewFile tests appending to a non-existent file (should create it)
func TestAppendToFile_NewFile(t *testing.T) {
	filename := "test_append_new.txt"
	defer os.Remove(filename)

	buffer := NewSeekBuffer([]byte("New file content"))
	err := buffer.AppendToFile(filename)
	if err != nil {
		t.Errorf("AppendToFile failed: %v", err)
	}

	// Verify the file was created
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Errorf("Failed to read file: %v", err)
	}

	if string(data) != "New file content" {
		t.Errorf("Expected 'New file content', got '%s'", string(data))
	}
}

// TestAppendUnreadToFile tests appending only unread portion of buffer
func TestAppendUnreadToFile(t *testing.T) {
	filename := "test_append_unread.txt"
	defer os.Remove(filename)

	// Create initial file
	initial := NewSeekBuffer([]byte("Start: "))
	err := initial.SaveToFile(filename)
	if err != nil {
		t.Errorf("SaveToFile failed: %v", err)
	}

	// Create buffer and mark some as "read"
	buffer := NewSeekBuffer([]byte("SKIP_THIS_PART|Include this part"))
	buffer.Seek(15) // Skip "SKIP_THIS_PART|"

	// Append only unread portion
	err = buffer.AppendUnreadToFile(filename)
	if err != nil {
		t.Errorf("AppendUnreadToFile failed: %v", err)
	}

	// Verify only unread portion was appended
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Errorf("Failed to read file: %v", err)
	}

	expected := "Start: Include this part"
	if string(data) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, string(data))
	}
}

// TestAppendUnreadToFile_AllRead tests appending when entire buffer is read
func TestAppendUnreadToFile_AllRead(t *testing.T) {
	filename := "test_append_unread_empty.txt"
	defer os.Remove(filename)

	// Create initial file
	initial := NewSeekBuffer([]byte("Initial content"))
	err := initial.SaveToFile(filename)
	if err != nil {
		t.Errorf("SaveToFile failed: %v", err)
	}

	// Create buffer with offset at end
	buffer := NewSeekBuffer([]byte("All read"))
	buffer.Seek(8) // Offset at end

	// Append unread (should append nothing)
	err = buffer.AppendUnreadToFile(filename)
	if err != nil {
		t.Errorf("AppendUnreadToFile failed: %v", err)
	}

	// Verify file content unchanged
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Errorf("Failed to read file: %v", err)
	}

	expected := "Initial content"
	if string(data) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, string(data))
	}
}

// TestAppendToFile_MultipleBuffers tests appending from multiple buffers in sequence
func TestAppendToFile_MultipleBuffers(t *testing.T) {
	filename := "test_append_multiple.txt"
	defer os.Remove(filename)

	buffers := []string{
		"Line 1\n",
		"Line 2\n",
		"Line 3\n",
		"Line 4\n",
		"Line 5\n",
	}

	for _, content := range buffers {
		buffer := NewSeekBuffer([]byte(content))
		err := buffer.AppendToFile(filename)
		if err != nil {
			t.Errorf("AppendToFile failed for '%s': %v", content, err)
		}
	}

	// Verify all lines were appended
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Errorf("Failed to read file: %v", err)
	}

	expected := "Line 1\nLine 2\nLine 3\nLine 4\nLine 5\n"
	if string(data) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, string(data))
	}
}

// TestMixedSaveAndAppend tests mixing SaveToFile and AppendToFile
func TestMixedSaveAndAppend(t *testing.T) {
	filename := "test_mixed.txt"
	defer os.Remove(filename)

	// Save initial data (overwrites)
	buffer1 := NewSeekBuffer([]byte("Initial"))
	err := buffer1.SaveToFile(filename)
	if err != nil {
		t.Errorf("SaveToFile failed: %v", err)
	}

	// Append more data
	buffer2 := NewSeekBuffer([]byte(" + Appended"))
	err = buffer2.AppendToFile(filename)
	if err != nil {
		t.Errorf("AppendToFile failed: %v", err)
	}

	// Overwrite with new save
	buffer3 := NewSeekBuffer([]byte("Replaced"))
	err = buffer3.SaveToFile(filename)
	if err != nil {
		t.Errorf("SaveToFile failed: %v", err)
	}

	// Append again
	buffer4 := NewSeekBuffer([]byte(" + More"))
	err = buffer4.AppendToFile(filename)
	if err != nil {
		t.Errorf("AppendToFile failed: %v", err)
	}

	// Verify final content
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Errorf("Failed to read file: %v", err)
	}

	expected := "Replaced + More"
	if string(data) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, string(data))
	}
}
