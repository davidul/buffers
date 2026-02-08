package seekbuffer

import (
	"testing"
)

func TestTransactionBasicCommit(t *testing.T) {
	buffer := NewSeekBuffer([]byte("Initial content"))
	tx := NewTransactionDecorator(buffer)

	// Start transaction
	tx.Begin()

	// Modify buffer
	tx.Write([]byte(" + Added"))

	// Check transaction buffer
	if string(tx.Bytes()) != "Initial content + Added" {
		t.Errorf("Expected 'Initial content + Added', got '%s'", string(tx.Bytes()))
	}

	// Original buffer unchanged
	if string(buffer.Bytes()) != "Initial content" {
		t.Errorf("Original buffer should be unchanged, got '%s'", string(buffer.Bytes()))
	}

	// Commit
	tx.Commit()

	// Now original buffer should be updated
	if string(buffer.Bytes()) != "Initial content + Added" {
		t.Errorf("Expected 'Initial content + Added' after commit, got '%s'", string(buffer.Bytes()))
	}
}

func TestTransactionRollback(t *testing.T) {
	buffer := NewSeekBuffer([]byte("Original"))
	tx := NewTransactionDecorator(buffer)

	tx.Begin()
	tx.Write([]byte(" + Changes"))

	// Check transaction buffer has changes
	if string(tx.Bytes()) != "Original + Changes" {
		t.Errorf("Transaction buffer should have changes, got '%s'", string(tx.Bytes()))
	}

	// Rollback
	tx.Rollback()

	// Original buffer unchanged
	if string(buffer.Bytes()) != "Original" {
		t.Errorf("Expected 'Original' after rollback, got '%s'", string(buffer.Bytes()))
	}

	// Transaction buffer restored
	if string(tx.Bytes()) != "Original" {
		t.Errorf("Transaction buffer should be restored to 'Original', got '%s'", string(tx.Bytes()))
	}
}

func TestNestedTransactions(t *testing.T) {
	buffer := NewSeekBuffer([]byte("Start"))
	tx := NewTransactionDecorator(buffer)

	// Level 1
	tx.Begin()
	tx.Write([]byte(" L1"))

	if tx.GetTransactionLevel() != 1 {
		t.Errorf("Expected level 1, got %d", tx.GetTransactionLevel())
	}

	// Level 2
	tx.Begin()
	tx.Write([]byte(" L2"))

	if tx.GetTransactionLevel() != 2 {
		t.Errorf("Expected level 2, got %d", tx.GetTransactionLevel())
	}

	// Level 3
	tx.Begin()
	tx.Write([]byte(" L3"))

	if tx.GetTransactionLevel() != 3 {
		t.Errorf("Expected level 3, got %d", tx.GetTransactionLevel())
	}

	// Rollback level 3
	tx.Rollback()
	if string(tx.Bytes()) != "Start L1 L2" {
		t.Errorf("After L3 rollback: expected 'Start L1 L2', got '%s'", string(tx.Bytes()))
	}

	// Commit level 2
	tx.Commit()
	if string(tx.Bytes()) != "Start L1 L2" {
		t.Errorf("After L2 commit: expected 'Start L1 L2', got '%s'", string(tx.Bytes()))
	}

	// Commit level 1
	tx.Commit()
	if string(buffer.Bytes()) != "Start L1 L2" {
		t.Errorf("After L1 commit: expected 'Start L1 L2' in buffer, got '%s'", string(buffer.Bytes()))
	}
}

func TestTransactionReadWrite(t *testing.T) {
	buffer := NewSeekBuffer([]byte("ABCDEFGH"))
	tx := NewTransactionDecorator(buffer)

	tx.Begin()

	// Read in transaction
	data := make([]byte, 4)
	n, _ := tx.Read(data)
	if n != 4 || string(data) != "ABCD" {
		t.Errorf("Expected to read 'ABCD', got '%s'", string(data))
	}

	// Write in transaction
	tx.Write([]byte("IJKL"))

	// Rewind and read all
	tx.Rewind()
	all := make([]byte, 12)
	n, _ = tx.Read(all)
	if string(all[:n]) != "ABCDEFGHIJKL" {
		t.Errorf("Expected 'ABCDEFGHIJKL', got '%s'", string(all[:n]))
	}

	// Commit
	tx.Commit()

	// Verify underlying buffer
	if string(buffer.Bytes()) != "ABCDEFGHIJKL" {
		t.Errorf("Expected 'ABCDEFGHIJKL' in buffer, got '%s'", string(buffer.Bytes()))
	}
}

func TestTransactionSeek(t *testing.T) {
	buffer := NewSeekBuffer([]byte("0123456789"))
	tx := NewTransactionDecorator(buffer)

	tx.Begin()

	// Seek to position 5
	tx.Seek(5)

	data := make([]byte, 3)
	tx.Read(data)
	if string(data) != "567" {
		t.Errorf("After seek(5): expected '567', got '%s'", string(data))
	}

	// Rollback
	tx.Rollback()

	// Can still use the buffer
	data2 := make([]byte, 3)
	buffer.Seek(2)
	buffer.Read(data2)
	if string(data2) != "234" {
		t.Errorf("After rollback: expected '234', got '%s'", string(data2))
	}
}

func TestAutoRollbackOnClose(t *testing.T) {
	buffer := NewSeekBuffer([]byte("Original"))
	tx := NewTransactionDecorator(buffer)

	tx.Begin()
	tx.Write([]byte(" + Changes"))

	// Close without commit - should auto-rollback
	tx.Close()

	// Original buffer should be unchanged
	if string(buffer.Bytes()) != "Original" {
		t.Errorf("Expected 'Original' after close, got '%s'", string(buffer.Bytes()))
	}
}

func TestTransactionWithoutBegin(t *testing.T) {
	buffer := NewSeekBuffer([]byte("Test"))
	tx := NewTransactionDecorator(buffer)

	// Try to commit without begin
	err := tx.Commit()
	if err == nil {
		t.Error("Expected error when committing without begin")
	}

	// Try to rollback without begin
	err = tx.Rollback()
	if err == nil {
		t.Error("Expected error when rolling back without begin")
	}
}

func TestTransactionInTransactionCheck(t *testing.T) {
	buffer := NewSeekBuffer([]byte("Test"))
	tx := NewTransactionDecorator(buffer)

	if tx.InTransaction() {
		t.Error("Should not be in transaction initially")
	}

	tx.Begin()
	if !tx.InTransaction() {
		t.Error("Should be in transaction after Begin()")
	}

	tx.Commit()
	if tx.InTransaction() {
		t.Error("Should not be in transaction after Commit()")
	}

	tx.Begin()
	tx.Rollback()
	if tx.InTransaction() {
		t.Error("Should not be in transaction after Rollback()")
	}
}

func TestTransactionReadBytes(t *testing.T) {
	buffer := NewSeekBuffer([]byte("Line1\nLine2\nLine3\n"))
	tx := NewTransactionDecorator(buffer)

	tx.Begin()

	// Read first line
	line1, err := tx.ReadBytes('\n')
	if err != nil {
		t.Errorf("ReadBytes failed: %v", err)
	}
	if string(line1) != "Line1\n" {
		t.Errorf("Expected 'Line1\\n', got '%s'", string(line1))
	}

	// Add a line
	tx.Write([]byte("Line4\n"))

	// Commit
	tx.Commit()

	// Verify buffer
	expected := "Line1\nLine2\nLine3\nLine4\n"
	if string(buffer.Bytes()) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, string(buffer.Bytes()))
	}
}

func TestTransactionAppend(t *testing.T) {
	buffer := NewSeekBuffer([]byte("Start"))
	tx := NewTransactionDecorator(buffer)

	tx.Begin()

	// Use Append instead of Write
	tx.Append([]byte(" Middle"))
	tx.Append([]byte(" End"))

	if string(tx.Bytes()) != "Start Middle End" {
		t.Errorf("Expected 'Start Middle End', got '%s'", string(tx.Bytes()))
	}

	tx.Commit()

	if string(buffer.Bytes()) != "Start Middle End" {
		t.Errorf("Expected 'Start Middle End' in buffer, got '%s'", string(buffer.Bytes()))
	}
}

func TestTransactionRewind(t *testing.T) {
	buffer := NewSeekBuffer([]byte("ABCDEFGH"))
	tx := NewTransactionDecorator(buffer)

	tx.Begin()

	// Read some data
	data := make([]byte, 4)
	tx.Read(data)

	// Rewind
	tx.Rewind()

	// Read again
	data2 := make([]byte, 4)
	tx.Read(data2)

	if string(data2) != "ABCD" {
		t.Errorf("After rewind: expected 'ABCD', got '%s'", string(data2))
	}

	tx.Commit()
}

func TestNestedTransactionRollbackAll(t *testing.T) {
	buffer := NewSeekBuffer([]byte("Base"))
	tx := NewTransactionDecorator(buffer)

	tx.Begin()
	tx.Write([]byte(" L1"))

	tx.Begin()
	tx.Write([]byte(" L2"))

	tx.Begin()
	tx.Write([]byte(" L3"))

	// Rollback all the way
	tx.Rollback() // L3
	tx.Rollback() // L2
	tx.Rollback() // L1

	// Should be back to original
	if string(buffer.Bytes()) != "Base" {
		t.Errorf("Expected 'Base' after rolling back all, got '%s'", string(buffer.Bytes()))
	}
}

func TestTransactionLen(t *testing.T) {
	buffer := NewSeekBuffer([]byte("Test Data"))
	tx := NewTransactionDecorator(buffer)

	tx.Begin()

	// Read some bytes
	data := make([]byte, 4)
	tx.Read(data)

	// Check length of unread data
	if tx.Len() != 5 { // " Data" = 5 bytes
		t.Errorf("Expected Len() = 5, got %d", tx.Len())
	}

	tx.Commit()
}

func TestTransactionWithEmptyBuffer(t *testing.T) {
	buffer := NewEmptySeekBuffer()
	tx := NewTransactionDecorator(buffer)

	tx.Begin()
	tx.Write([]byte("New Data"))
	tx.Commit()

	if string(buffer.Bytes()) != "New Data" {
		t.Errorf("Expected 'New Data', got '%s'", string(buffer.Bytes()))
	}
}

func TestMultipleSequentialTransactions(t *testing.T) {
	buffer := NewSeekBuffer([]byte("Start"))
	tx := NewTransactionDecorator(buffer)

	// Transaction 1
	tx.Begin()
	tx.Write([]byte(" T1"))
	tx.Commit()

	// Transaction 2
	tx.Begin()
	tx.Write([]byte(" T2"))
	tx.Commit()

	// Transaction 3
	tx.Begin()
	tx.Write([]byte(" T3"))
	tx.Commit()

	expected := "Start T1 T2 T3"
	if string(buffer.Bytes()) != expected {
		t.Errorf("Expected '%s', got '%s'", expected, string(buffer.Bytes()))
	}
}
