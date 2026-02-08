package main

import (
	"fmt"
	"os"

	"github.com/davidul/buffers/davidul/seekbuffer"
)

func main() {
	fmt.Println("=== SeekBuffer Transaction Examples ===\n")

	// Example 1: Basic Transaction - Commit
	example1()

	// Example 2: Transaction Rollback
	example2()

	// Example 3: Nested Transactions
	example3()

	// Example 4: Transaction with File Sync
	example4()

	// Example 5: Error Handling
	example5()

	// Example 6: Stacking Decorators (Transaction + FileSync)
	example6()
}

func example1() {
	fmt.Println("Example 1: Basic Transaction - Commit")
	fmt.Println("--------------------------------------")

	buffer := seekbuffer.NewSeekBuffer([]byte("Account Balance: $1000"))
	tx := seekbuffer.NewTransactionDecorator(buffer)

	fmt.Printf("Before transaction: %s\n", string(buffer.Bytes()))

	// Begin transaction
	err := tx.Begin()
	if err != nil {
		return
	}

	// Make changes
	tx.Write([]byte(" -> $1500"))
	fmt.Printf("In transaction: %s\n", string(tx.Bytes()))
	fmt.Printf("Original buffer (unchanged): %s\n", string(buffer.Bytes()))

	// Commit
	tx.Commit()
	fmt.Printf("After commit: %s\n\n", string(buffer.Bytes()))
}

func example2() {
	fmt.Println("Example 2: Transaction Rollback")
	fmt.Println("--------------------------------")

	buffer := seekbuffer.NewSeekBuffer([]byte("Valid Data"))
	tx := seekbuffer.NewTransactionDecorator(buffer)

	fmt.Printf("Before transaction: %s\n", string(buffer.Bytes()))

	tx.Begin()
	tx.Write([]byte(" + Invalid Changes"))
	fmt.Printf("In transaction: %s\n", string(tx.Bytes()))

	// Rollback - discard changes
	tx.Rollback()
	fmt.Printf("After rollback: %s\n", string(buffer.Bytes()))
	fmt.Printf("Transaction buffer restored: %s\n\n", string(tx.Bytes()))
}

func example3() {
	fmt.Println("Example 3: Nested Transactions")
	fmt.Println("-------------------------------")

	buffer := seekbuffer.NewSeekBuffer([]byte("Start"))
	tx := seekbuffer.NewTransactionDecorator(buffer)

	// Outer transaction
	tx.Begin()
	tx.Write([]byte(" -> Outer"))
	fmt.Printf("Level 1 (nesting=%d): %s\n", tx.GetTransactionLevel(), string(tx.Bytes()))

	// Inner transaction
	tx.Begin()
	tx.Write([]byte(" -> Inner"))
	fmt.Printf("Level 2 (nesting=%d): %s\n", tx.GetTransactionLevel(), string(tx.Bytes()))

	// Another nested level
	tx.Begin()
	tx.Write([]byte(" -> Deepest"))
	fmt.Printf("Level 3 (nesting=%d): %s\n", tx.GetTransactionLevel(), string(tx.Bytes()))

	// Rollback deepest
	tx.Rollback()
	fmt.Printf("After inner rollback (nesting=%d): %s\n", tx.GetTransactionLevel(), string(tx.Bytes()))

	// Commit middle level
	tx.Commit()
	fmt.Printf("After middle commit (nesting=%d): %s\n", tx.GetTransactionLevel(), string(tx.Bytes()))

	// Commit outer
	tx.Commit()
	fmt.Printf("After outer commit (nesting=%d): %s\n", tx.GetTransactionLevel(), string(buffer.Bytes()))
	fmt.Printf("In transaction: %v\n\n", tx.InTransaction())
}

func example4() {
	fmt.Println("Example 4: Transaction + File Sync")
	fmt.Println("-----------------------------------")

	filename := "account_transaction.dat"
	defer os.Remove(filename)

	// Create buffer
	buffer := seekbuffer.NewEmptySeekBuffer()

	// Add file sync
	synced := seekbuffer.NewFileSyncDecorator(buffer)
	synced.EnableFileSync(filename)
	defer synced.Close()

	// Add transactions on top of file sync
	tx := seekbuffer.NewTransactionDecorator(synced)

	// Initial data
	tx.Begin()
	tx.Write([]byte("Balance: $1000"))
	tx.Commit()
	fmt.Println("✓ Committed to file: Balance: $1000")

	// Transaction that will rollback
	tx.Begin()
	tx.Write([]byte(" + $500 (fraudulent)"))
	fmt.Printf("Pending change: %s\n", string(tx.Bytes()))
	tx.Rollback()
	fmt.Println("✗ Rolled back - file unchanged")

	// Successful transaction
	tx.Begin()
	tx.Write([]byte(" + $200 (deposit)"))
	tx.Commit()
	fmt.Printf("✓ Final balance: %s\n\n", string(synced.Bytes()))

	// Verify file
	data, _ := os.ReadFile(filename)
	fmt.Printf("File content: %s\n\n", string(data))
}

func example5() {
	fmt.Println("Example 5: Error Handling")
	fmt.Println("--------------------------")

	buffer := seekbuffer.NewSeekBuffer([]byte("Test"))
	tx := seekbuffer.NewTransactionDecorator(buffer)

	// Try to commit without begin
	err := tx.Commit()
	if err != nil {
		fmt.Printf("Error (expected): %v\n", err)
	}

	// Try to rollback without begin
	err = tx.Rollback()
	if err != nil {
		fmt.Printf("Error (expected): %v\n\n", err)
	}
}

func example6() {
	fmt.Println("Example 6: Stacking Decorators")
	fmt.Println("-------------------------------")

	filename := "stacked_example.dat"
	defer os.Remove(filename)

	// Base buffer
	buffer := seekbuffer.NewEmptySeekBuffer()
	fmt.Println("Layer 1: SeekBuffer (base)")

	// Add file sync
	synced := seekbuffer.NewFileSyncDecorator(buffer)
	synced.EnableFileSync(filename)
	defer synced.Close()
	fmt.Println("Layer 2: FileSyncDecorator (persistence)")

	// Add transactions
	tx := seekbuffer.NewTransactionDecorator(synced)
	fmt.Println("Layer 3: TransactionDecorator (ACID)")
	fmt.Println()

	// Now we have: Transactions -> File Sync -> Buffer
	fmt.Println("Performing operations through all layers:")

	// Start transaction
	tx.Begin()
	fmt.Println("- Begin transaction")

	// Write data
	tx.Write([]byte("Data Layer 1"))
	fmt.Println("- Write 'Data Layer 1'")

	// This is in transaction buffer only
	fmt.Printf("  Transaction buffer: %s\n", string(tx.Bytes()))
	fmt.Printf("  File sync buffer: %s\n", string(synced.Bytes()))

	// Commit transaction
	tx.Commit()
	fmt.Println("- Commit transaction")

	// Now data flows through: Transaction -> FileSync -> File
	fmt.Printf("  File sync buffer: %s\n", string(synced.Bytes()))

	// Verify file
	data, _ := os.ReadFile(filename)
	fmt.Printf("  File content: %s\n", string(data))

	fmt.Println()
	fmt.Println("✓ All layers working together!")
}
