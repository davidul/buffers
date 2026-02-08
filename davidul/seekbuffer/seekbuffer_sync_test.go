package seekbuffer

import (
	"os"
	"testing"
)

func TestEnableFileSync_BasicWrite(t *testing.T) {
	filename := "test_sync_write.dat"
	defer os.Remove(filename)

	buffer := NewEmptySeekBuffer()
	decorator := NewFileSyncDecorator(buffer)
	err := decorator.EnableFileSync(filename)
	if err != nil {
		t.Fatalf("EnableFileSync failed: %v", err)
	}
	defer decorator.Close()

	// Write some data
	decorator.Write([]byte("Hello, World!"))

	// Verify data is in buffer
	if string(decorator.Bytes()) != "Hello, World!" {
		t.Errorf("SeekableBuffer content mismatch: got '%s'", string(decorator.Bytes()))
	}

	// Verify data is in file
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(data) != "Hello, World!" {
		t.Errorf("File content mismatch: got '%s'", string(data))
	}
}

func TestEnableFileSync_MultipleWrites(t *testing.T) {
	filename := "test_sync_multiple.dat"
	defer os.Remove(filename)

	buffer := NewEmptySeekBuffer()
	decorator := NewFileSyncDecorator(buffer)
	err := decorator.EnableFileSync(filename)
	if err != nil {
		t.Fatalf("EnableFileSync failed: %v", err)
	}
	defer decorator.Close()

	// Write data in multiple chunks
	decorator.Write([]byte("First "))
	decorator.Write([]byte("Second "))
	decorator.Write([]byte("Third"))

	// Verify file content
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	expected := "First Second Third"
	if string(data) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, string(data))
	}
}

func TestEnableFileSync_WithRewind(t *testing.T) {
	filename := "test_sync_rewind.dat"
	defer os.Remove(filename)

	buffer := NewSeekBuffer([]byte("ABCDEFGHIJ"))
	decorator := NewFileSyncDecorator(buffer)
	err := decorator.EnableFileSync(filename)
	if err != nil {
		t.Fatalf("EnableFileSync failed: %v", err)
	}
	defer decorator.Close()

	// Read some data
	dst := make([]byte, 5)
	n, err := decorator.Read(dst)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if n != 5 || string(dst) != "ABCDE" {
		t.Errorf("First read: expected 'ABCDE', got '%s'", string(dst))
	}

	// Rewind
	decorator.Rewind()

	// Read again - should get same data
	dst2 := make([]byte, 5)
	n2, err := decorator.Read(dst2)
	if err != nil {
		t.Fatalf("Read after rewind failed: %v", err)
	}
	if n2 != 5 || string(dst2) != "ABCDE" {
		t.Errorf("After rewind: expected 'ABCDE', got '%s'", string(dst2))
	}
}

func TestEnableFileSync_WithSeek(t *testing.T) {
	filename := "test_sync_seek.dat"
	defer os.Remove(filename)

	buffer := NewSeekBuffer([]byte("0123456789"))
	decorator := NewFileSyncDecorator(buffer)
	err := decorator.EnableFileSync(filename)
	if err != nil {
		t.Fatalf("EnableFileSync failed: %v", err)
	}
	defer decorator.Close()

	// Seek to position 5
	decorator.Seek(5)

	// Read from position 5
	dst := make([]byte, 3)
	n, err := decorator.Read(dst)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if n != 3 || string(dst) != "567" {
		t.Errorf("After seek: expected '567', got '%s'", string(dst))
	}

	// Seek back to 2
	decorator.Seek(2)

	// Read from position 2
	dst2 := make([]byte, 3)
	n2, err := decorator.Read(dst2)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if n2 != 3 || string(dst2) != "234" {
		t.Errorf("After second seek: expected '234', got '%s'", string(dst2))
	}
}

func TestEnableFileSync_ExistingBuffer(t *testing.T) {
	filename := "test_sync_existing.dat"
	defer os.Remove(filename)

	// Create buffer with existing data
	buffer := NewSeekBuffer([]byte("Existing content"))
	decorator := NewFileSyncDecorator(buffer)

	// Enable sync - should write existing content to file
	err := decorator.EnableFileSync(filename)
	if err != nil {
		t.Fatalf("EnableFileSync failed: %v", err)
	}
	defer decorator.Close()

	// Verify file has existing content
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(data) != "Existing content" {
		t.Errorf("Expected 'Existing content', got '%s'", string(data))
	}

	// Add more data
	decorator.Write([]byte(" + More"))

	// Verify file has all content
	data, err = os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(data) != "Existing content + More" {
		t.Errorf("Expected 'Existing content + More', got '%s'", string(data))
	}
}

func TestDisableFileSync(t *testing.T) {
	filename := "test_sync_disable.dat"
	defer os.Remove(filename)

	buffer := NewEmptySeekBuffer()
	decorator := NewFileSyncDecorator(buffer)
	err := decorator.EnableFileSync(filename)
	if err != nil {
		t.Fatalf("EnableFileSync failed: %v", err)
	}

	// Write some data
	decorator.Write([]byte("Synced data"))

	// Verify it's synced
	data, _ := os.ReadFile(filename)
	if string(data) != "Synced data" {
		t.Errorf("Before disable: expected 'Synced data', got '%s'", string(data))
	}

	// Disable sync
	err = decorator.DisableFileSync()
	if err != nil {
		t.Fatalf("DisableFileSync failed: %v", err)
	}

	// Write more data - should NOT sync to file
	decorator.Write([]byte(" + Not synced"))

	// Verify buffer has all data
	if string(decorator.Bytes()) != "Synced data + Not synced" {
		t.Errorf("SeekableBuffer should have all data: got '%s'", string(decorator.Bytes()))
	}

	// Verify file still has old data only
	data, _ = os.ReadFile(filename)
	if string(data) != "Synced data" {
		t.Errorf("After disable: file should still have 'Synced data', got '%s'", string(data))
	}
}

func TestIsSyncEnabled(t *testing.T) {
	buffer := NewEmptySeekBuffer()
	decorator := NewFileSyncDecorator(buffer)

	// Initially not synced
	if decorator.IsSyncEnabled() {
		t.Error("SeekableBuffer should not be synced initially")
	}

	// Enable sync
	filename := "test_is_sync.dat"
	defer os.Remove(filename)
	decorator.EnableFileSync(filename)
	defer decorator.Close()

	if !decorator.IsSyncEnabled() {
		t.Error("SeekableBuffer should be synced after EnableFileSync")
	}

	// Disable sync
	decorator.DisableFileSync()

	if decorator.IsSyncEnabled() {
		t.Error("SeekableBuffer should not be synced after DisableFileSync")
	}
}

func TestGetSyncFilename(t *testing.T) {
	buffer := NewEmptySeekBuffer()
	decorator := NewFileSyncDecorator(buffer)

	// Initially no filename
	if decorator.GetSyncFilename() != "" {
		t.Errorf("Expected empty filename, got '%s'", decorator.GetSyncFilename())
	}

	// Enable sync
	filename := "test_filename.dat"
	defer os.Remove(filename)
	decorator.EnableFileSync(filename)
	defer decorator.Close()

	if decorator.GetSyncFilename() != filename {
		t.Errorf("Expected '%s', got '%s'", filename, decorator.GetSyncFilename())
	}

	// Disable sync
	decorator.DisableFileSync()

	if decorator.GetSyncFilename() != "" {
		t.Errorf("After disable, expected empty filename, got '%s'", decorator.GetSyncFilename())
	}
}

func TestEnableFileSync_SwitchFiles(t *testing.T) {
	file1 := "test_sync_file1.dat"
	file2 := "test_sync_file2.dat"
	defer os.Remove(file1)
	defer os.Remove(file2)

	buffer := NewEmptySeekBuffer()
	decorator := NewFileSyncDecorator(buffer)

	// Sync to first file
	decorator.EnableFileSync(file1)
	decorator.Write([]byte("File 1 content"))

	// Verify file1
	data1, _ := os.ReadFile(file1)
	if string(data1) != "File 1 content" {
		t.Errorf("File1: expected 'File 1 content', got '%s'", string(data1))
	}

	// Switch to second file - should write current buffer to new file
	decorator.EnableFileSync(file2)

	// Verify file2 has the buffer content
	data2, _ := os.ReadFile(file2)
	if string(data2) != "File 1 content" {
		t.Errorf("File2: expected 'File 1 content', got '%s'", string(data2))
	}

	// Add more data
	decorator.Write([]byte(" + More"))

	// Verify file2 has updated content
	data2, _ = os.ReadFile(file2)
	if string(data2) != "File 1 content + More" {
		t.Errorf("File2 after write: expected 'File 1 content + More', got '%s'", string(data2))
	}

	decorator.Close()
}

func TestEnableFileSync_Append(t *testing.T) {
	filename := "test_sync_append.dat"
	defer os.Remove(filename)

	buffer := NewEmptySeekBuffer()
	decorator := NewFileSyncDecorator(buffer)
	decorator.EnableFileSync(filename)
	defer decorator.Close()

	// Use Append instead of Write
	decorator.Append([]byte("First"))
	decorator.Append([]byte(" Second"))
	decorator.Append([]byte(" Third"))

	// Verify file content
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	expected := "First Second Third"
	if string(data) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, string(data))
	}
}

func TestEnableFileSync_ReadWriteMix(t *testing.T) {
	filename := "test_sync_readwrite.dat"
	defer os.Remove(filename)

	buffer := NewSeekBuffer([]byte("Initial content"))
	decorator := NewFileSyncDecorator(buffer)
	decorator.EnableFileSync(filename)
	defer decorator.Close()

	// Read some data
	dst := make([]byte, 7)
	decorator.Read(dst)
	if string(dst) != "Initial" {
		t.Errorf("Read: expected 'Initial', got '%s'", string(dst))
	}

	// Write more data
	decorator.Write([]byte(" + Added"))

	// Verify buffer
	expected := "Initial content + Added"
	if string(decorator.Bytes()) != expected {
		t.Errorf("SeekableBuffer: expected '%s', got '%s'", expected, string(decorator.Bytes()))
	}

	// Verify file
	data, _ := os.ReadFile(filename)
	if string(data) != expected {
		t.Errorf("File: expected '%s', got '%s'", expected, string(data))
	}

	// Rewind and read all
	decorator.Rewind()
	allData := make([]byte, len(decorator.Bytes()))
	n, _ := decorator.Read(allData)
	if n != len(expected) || string(allData) != expected {
		t.Errorf("After rewind: expected '%s', got '%s'", expected, string(allData))
	}
}
