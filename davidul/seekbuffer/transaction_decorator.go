package seekbuffer

import (
	"fmt"
	"io"
)

// TransactionDecorator wraps any SeekableBuffer implementation and adds transaction support.
// It allows you to group operations and commit or rollback them atomically.
//
// Features:
//   - Begin/Commit/Rollback transactions
//   - Nested transactions with savepoints
//   - Auto-rollback on close if transaction active
//   - All operations work on transaction buffer until commit
type TransactionDecorator struct {
	buffer            SeekableBuffer
	inTransaction     bool
	transactionBuffer []byte
	transactionOffset int
	originalBuffer    []byte
	originalOffset    int
	nestedLevel       int
	savepoints        []transactionSavepoint
}

type transactionSavepoint struct {
	buffer []byte
	offset int
	level  int
}

// NewTransactionDecorator creates a decorator that wraps the given buffer
// with transaction support
func NewTransactionDecorator(buffer SeekableBuffer) *TransactionDecorator {
	return &TransactionDecorator{
		buffer:        buffer,
		inTransaction: false,
		nestedLevel:   0,
		savepoints:    make([]transactionSavepoint, 0),
	}
}

// Begin starts a new transaction. All subsequent operations will be buffered
// until Commit() or Rollback() is called.
//
// Supports nested transactions - each Begin() must be matched with either
// Commit() or Rollback().
func (d *TransactionDecorator) Begin() error {
	if !d.inTransaction {
		// First transaction - save original state
		d.originalBuffer = make([]byte, len(d.buffer.Bytes()))
		copy(d.originalBuffer, d.buffer.Bytes())
		d.originalOffset = d.getCurrentOffset()

		d.transactionBuffer = make([]byte, len(d.buffer.Bytes()))
		copy(d.transactionBuffer, d.buffer.Bytes())
		d.transactionOffset = d.originalOffset

		d.inTransaction = true
		d.nestedLevel = 1
	} else {
		// Nested transaction - create savepoint
		savepoint := transactionSavepoint{
			buffer: make([]byte, len(d.transactionBuffer)),
			offset: d.transactionOffset,
			level:  d.nestedLevel,
		}
		copy(savepoint.buffer, d.transactionBuffer)
		d.savepoints = append(d.savepoints, savepoint)
		d.nestedLevel++
	}

	return nil
}

// Commit commits the current transaction, applying all buffered changes.
// For nested transactions, it commits to the parent transaction.
// Only the outermost Commit() applies changes to the underlying buffer.
func (d *TransactionDecorator) Commit() error {
	if !d.inTransaction {
		return fmt.Errorf("no transaction in progress")
	}

	if d.nestedLevel > 1 {
		// Nested transaction - just remove the savepoint and continue
		d.nestedLevel--
		if len(d.savepoints) > 0 {
			d.savepoints = d.savepoints[:len(d.savepoints)-1]
		}
		return nil
	}

	// Top-level transaction - apply changes to underlying buffer
	d.nestedLevel = 0
	d.inTransaction = false
	d.savepoints = d.savepoints[:0]

	// Apply transaction buffer to underlying buffer
	d.buffer.Close()
	if len(d.transactionBuffer) > 0 {
		d.buffer.Write(d.transactionBuffer)
	}
	d.buffer.Seek(d.transactionOffset)

	return nil
}

// Rollback rolls back the current transaction, discarding all changes.
// For nested transactions, it rolls back to the parent transaction's savepoint.
func (d *TransactionDecorator) Rollback() error {
	if !d.inTransaction {
		return fmt.Errorf("no transaction in progress")
	}

	if d.nestedLevel > 1 {
		// Nested transaction - restore from savepoint
		d.nestedLevel--
		if len(d.savepoints) > 0 {
			savepoint := d.savepoints[len(d.savepoints)-1]
			d.transactionBuffer = make([]byte, len(savepoint.buffer))
			copy(d.transactionBuffer, savepoint.buffer)
			d.transactionOffset = savepoint.offset
			d.savepoints = d.savepoints[:len(d.savepoints)-1]
		}
		return nil
	}

	// Top-level transaction - restore original state
	d.transactionBuffer = make([]byte, len(d.originalBuffer))
	copy(d.transactionBuffer, d.originalBuffer)
	d.transactionOffset = d.originalOffset

	d.nestedLevel = 0
	d.inTransaction = false
	d.savepoints = d.savepoints[:0]

	return nil
}

// InTransaction returns true if a transaction is currently in progress
func (d *TransactionDecorator) InTransaction() bool {
	return d.inTransaction
}

// GetTransactionLevel returns the current nesting level of transactions (0 if not in transaction)
func (d *TransactionDecorator) GetTransactionLevel() int {
	if !d.inTransaction {
		return 0
	}
	return d.nestedLevel
}

// Write delegates to transaction buffer if in transaction, otherwise to underlying buffer
func (d *TransactionDecorator) Write(p []byte) (int, error) {
	if d.inTransaction {
		d.transactionBuffer = append(d.transactionBuffer, p...)
		return len(p), nil
	}
	return d.buffer.Write(p)
}

// Read reads from transaction buffer if in transaction, otherwise from underlying buffer
func (d *TransactionDecorator) Read(p []byte) (int, error) {
	if d.inTransaction {
		if d.transactionOffset >= len(d.transactionBuffer) {
			return 0, io.EOF
		}
		n := copy(p, d.transactionBuffer[d.transactionOffset:])
		d.transactionOffset += n
		return n, nil
	}
	return d.buffer.Read(p)
}

// Append appends to transaction buffer if in transaction, otherwise to underlying buffer
func (d *TransactionDecorator) Append(src []byte) {
	if d.inTransaction {
		d.transactionBuffer = append(d.transactionBuffer, src...)
	} else {
		d.buffer.Append(src)
	}
}

// Bytes returns the transaction buffer if in transaction, otherwise underlying buffer
func (d *TransactionDecorator) Bytes() []byte {
	if d.inTransaction {
		return d.transactionBuffer
	}
	return d.buffer.Bytes()
}

// Rewind resets offset in transaction or underlying buffer
func (d *TransactionDecorator) Rewind() {
	if d.inTransaction {
		d.transactionOffset = 0
	} else {
		d.buffer.Rewind()
	}
}

// Seek seeks in transaction or underlying buffer
func (d *TransactionDecorator) Seek(offset int) {
	if d.inTransaction {
		d.transactionOffset = offset
	} else {
		d.buffer.Seek(offset)
	}
}

// Len returns unread length from transaction or underlying buffer
func (d *TransactionDecorator) Len() int {
	if d.inTransaction {
		return len(d.transactionBuffer) - d.transactionOffset
	}
	return d.buffer.Len()
}

// ReadBytes reads until delimiter from transaction or underlying buffer
func (d *TransactionDecorator) ReadBytes(c byte) ([]byte, error) {
	if d.inTransaction {
		for i := d.transactionOffset; i < len(d.transactionBuffer); i++ {
			if d.transactionBuffer[i] == c {
				result := d.transactionBuffer[d.transactionOffset : i+1]
				d.transactionOffset = i + 1
				return result, nil
			}
		}
		// Not found
		result := d.transactionBuffer[d.transactionOffset:]
		d.transactionOffset = len(d.transactionBuffer)
		return result, io.EOF
	}
	return d.buffer.ReadBytes(c)
}

// Close closes the transaction decorator and underlying buffer
// Automatically rolls back any active transaction
func (d *TransactionDecorator) Close() error {
	if d.inTransaction {
		// Auto-rollback on close if transaction is active
		d.Rollback()
	}
	// Don't close the underlying buffer - just the decorator
	// The underlying buffer should be closed by its owner
	return nil
}

// getCurrentOffset is a helper to get the current offset from the underlying buffer
func (d *TransactionDecorator) getCurrentOffset() int {
	totalLen := len(d.buffer.Bytes())
	unreadLen := d.buffer.Len()
	return totalLen - unreadLen
}
