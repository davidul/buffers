package seekbuffer

import (
	"os"
	"testing"
)

func TestEnableFileSync_BasicWrite(t *testing.T) {
	filename := "test_sync_write.dat"
	defer os.Remove(filename)

	buffer := NewEmptySeekBuffer()
	err := buffer.EnableFileSync(filename)
	if err != nil {
		t.Fatalf("EnableFileSync failed: %v", err)
	}
	defer buffer.Close()

	// Write some data
	buffer.Write([]byte("Hello, World!"))

	// Verify data is in buffer
	if string(buffer.Bytes()) != "Hello, World!" {
		t.Errorf("Buffer content mismatch: got '%s'", string(buffer.Bytes()))
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
	err := buffer.EnableFileSync(filename)
	if err != nil {
		t.Fatalf("EnableFileSync failed: %v", err)
	}
	defer buffer.Close()

	// Write data in multiple chunks
	buffer.Write([]byte("First "))
	buffer.Write([]byte("Second "))
	buffer.Write([]byte("Third"))

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
	err := buffer.EnableFileSync(filename)
	if err != nil {
		t.Fatalf("EnableFileSync failed: %v", err)
	}
	defer buffer.Close()

	// Read some data
	dst := make([]byte, 5)
	n, err := buffer.Read(dst)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if n != 5 || string(dst) != "ABCDE" {
		t.Errorf("First read: expected 'ABCDE', got '%s'", string(dst))
	}

	// Rewind
	buffer.Rewind()

	// Read again - should get same data
	dst2 := make([]byte, 5)
	n2, err := buffer.Read(dst2)
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
	err := buffer.EnableFileSync(filename)
	if err != nil {
		t.Fatalf("EnableFileSync failed: %v", err)
	}
	defer buffer.Close()

	// Seek to position 5
	buffer.Seek(5)

	// Read from position 5
	dst := make([]byte, 3)
	n, err := buffer.Read(dst)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}
	if n != 3 || string(dst) != "567" {
		t.Errorf("After seek: expected '567', got '%s'", string(dst))
	}

	// Seek back to 2
	buffer.Seek(2)

	// Read from position 2
	dst2 := make([]byte, 3)
	n2, err := buffer.Read(dst2)
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

	// Enable sync - should write existing content to file
	err := buffer.EnableFileSync(filename)
	if err != nil {
		t.Fatalf("EnableFileSync failed: %v", err)
	}
	defer buffer.Close()

	// Verify file has existing content
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(data) != "Existing content" {
		t.Errorf("Expected 'Existing content', got '%s'", string(data))
	}

	// Add more data
	buffer.Write([]byte(" + More"))

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
	err := buffer.EnableFileSync(filename)
	if err != nil {
		t.Fatalf("EnableFileSync failed: %v", err)
	}

	// Write some data
	buffer.Write([]byte("Synced data"))

	// Verify it's synced
	data, _ := os.ReadFile(filename)
	if string(data) != "Synced data" {
		t.Errorf("Before disable: expected 'Synced data', got '%s'", string(data))
	}

	// Disable sync
	err = buffer.DisableFileSync()
	if err != nil {
		t.Fatalf("DisableFileSync failed: %v", err)
	}

	// Write more data - should NOT sync to file
	buffer.Write([]byte(" + Not synced"))

	// Verify buffer has all data
	if string(buffer.Bytes()) != "Synced data + Not synced" {
		t.Errorf("Buffer should have all data: got '%s'", string(buffer.Bytes()))
	}

	// Verify file still has old data only
	data, _ = os.ReadFile(filename)
	if string(data) != "Synced data" {
		t.Errorf("After disable: file should still have 'Synced data', got '%s'", string(data))
	}
}

func TestIsSyncEnabled(t *testing.T) {
	buffer := NewEmptySeekBuffer()

	// Initially not synced
	if buffer.IsSyncEnabled() {
		t.Error("Buffer should not be synced initially")
	}

	// Enable sync
	filename := "test_is_sync.dat"
	defer os.Remove(filename)
	buffer.EnableFileSync(filename)
	defer buffer.Close()

	if !buffer.IsSyncEnabled() {
		t.Error("Buffer should be synced after EnableFileSync")
	}

	// Disable sync
	buffer.DisableFileSync()

	if buffer.IsSyncEnabled() {
		t.Error("Buffer should not be synced after DisableFileSync")
	}
}

func TestGetSyncFilename(t *testing.T) {
	buffer := NewEmptySeekBuffer()

	// Initially no filename
	if buffer.GetSyncFilename() != "" {
		t.Errorf("Expected empty filename, got '%s'", buffer.GetSyncFilename())
	}

	// Enable sync
	filename := "test_filename.dat"
	defer os.Remove(filename)
	buffer.EnableFileSync(filename)
	defer buffer.Close()

	if buffer.GetSyncFilename() != filename {
		t.Errorf("Expected '%s', got '%s'", filename, buffer.GetSyncFilename())
	}

	// Disable sync
	buffer.DisableFileSync()

	if buffer.GetSyncFilename() != "" {
		t.Errorf("After disable, expected empty filename, got '%s'", buffer.GetSyncFilename())
	}
}

func TestEnableFileSync_SwitchFiles(t *testing.T) {
	file1 := "test_sync_file1.dat"
	file2 := "test_sync_file2.dat"
	defer os.Remove(file1)
	defer os.Remove(file2)

	buffer := NewEmptySeekBuffer()

	// Sync to first file
	buffer.EnableFileSync(file1)
	buffer.Write([]byte("File 1 content"))

	// Verify file1
	data1, _ := os.ReadFile(file1)
	if string(data1) != "File 1 content" {
		t.Errorf("File1: expected 'File 1 content', got '%s'", string(data1))
	}

	// Switch to second file - should write current buffer to new file
	buffer.EnableFileSync(file2)

	// Verify file2 has the buffer content
	data2, _ := os.ReadFile(file2)
	if string(data2) != "File 1 content" {
		t.Errorf("File2: expected 'File 1 content', got '%s'", string(data2))
	}

	// Add more data
	buffer.Write([]byte(" + More"))

	// Verify file2 has updated content
	data2, _ = os.ReadFile(file2)
	if string(data2) != "File 1 content + More" {
		t.Errorf("File2 after write: expected 'File 1 content + More', got '%s'", string(data2))
	}

	buffer.Close()
}

func TestEnableFileSync_Append(t *testing.T) {
	filename := "test_sync_append.dat"
	defer os.Remove(filename)

	buffer := NewEmptySeekBuffer()
	buffer.EnableFileSync(filename)
	defer buffer.Close()

	// Use Append instead of Write
	buffer.Append([]byte("First"))
	buffer.Append([]byte(" Second"))
	buffer.Append([]byte(" Third"))

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
	buffer.EnableFileSync(filename)
	defer buffer.Close()

	// Read some data
	dst := make([]byte, 7)
	buffer.Read(dst)
	if string(dst) != "Initial" {
		t.Errorf("Read: expected 'Initial', got '%s'", string(dst))
	}

	// Write more data
	buffer.Write([]byte(" + Added"))

	// Verify buffer
	expected := "Initial content + Added"
	if string(buffer.Bytes()) != expected {
		t.Errorf("Buffer: expected '%s', got '%s'", expected, string(buffer.Bytes()))
	}

	// Verify file
	data, _ := os.ReadFile(filename)
	if string(data) != expected {
		t.Errorf("File: expected '%s', got '%s'", expected, string(data))
	}

	// Rewind and read all
	buffer.Rewind()
	allData := make([]byte, len(buffer.Bytes()))
	n, _ := buffer.Read(allData)
	if n != len(expected) || string(allData) != expected {
		t.Errorf("After rewind: expected '%s', got '%s'", expected, string(allData))
	}
}
