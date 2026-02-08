package main

import (
	"fmt"
	"log"
	"os"

	"github.com/davidul/buffers/davidul/seekbuffer"
)

func main() {
	fmt.Println("=== SeekBuffer: Using All Decorators ===\n")

	// Example 1: Single Decorators
	example1_SingleDecorators()

	// Example 2: Two Decorators - Logging + FileSync
	example2_LoggingAndFileSync()

	// Example 3: Two Decorators - Logging + Transactions
	example3_LoggingAndTransactions()

	// Example 4: Two Decorators - Transactions + FileSync
	example4_TransactionsAndFileSync()

	// Example 5: All Three Decorators
	example5_AllThreeDecorators()

	// Example 6: Different Stacking Orders
	example6_StackingOrders()

	// Example 7: Real-World Use Case
	example7_RealWorldUseCase()

	fmt.Println("\n=== All Examples Complete ===")
}

func example1_SingleDecorators() {
	fmt.Println("Example 1: Single Decorators")
	fmt.Println("=============================\n")

	// 1a. Just SeekBuffer (no decorators)
	fmt.Println("1a. Core SeekBuffer (no decorators):")
	buffer := seekbuffer.NewSeekBuffer([]byte("Hello World"))
	buffer.Write([]byte("!"))
	fmt.Printf("   Result: %s\n\n", string(buffer.Bytes()))

	// 1b. With Logging
	fmt.Println("1b. With Logging Decorator:")
	buffer2 := seekbuffer.NewSeekBuffer([]byte("Hello"))
	logger := log.New(os.Stdout, "   [LOG] ", 0)
	logged := seekbuffer.NewLoggingDecorator(buffer2, logger, "LoggedBuffer")
	logged.Write([]byte(" World"))
	fmt.Println()

	// 1c. With Transactions
	fmt.Println("1c. With Transaction Decorator:")
	buffer3 := seekbuffer.NewSeekBuffer([]byte("Initial"))
	tx := seekbuffer.NewTransactionDecorator(buffer3)
	tx.Begin()
	tx.Write([]byte(" + Transaction"))
	tx.Commit()
	fmt.Printf("   Result: %s\n\n", string(buffer3.Bytes()))

	// 1d. With FileSync (not enabling for demo)
	fmt.Println("1d. With FileSync Decorator:")
	buffer4 := seekbuffer.NewSeekBuffer([]byte("File data"))
	synced := seekbuffer.NewFileSyncDecorator(buffer4)
	fmt.Println("   (FileSync created but not enabled for demo)")
	synced.Write([]byte(" more"))
	fmt.Printf("   Result: %s\n\n", string(buffer4.Bytes()))
}

func example2_LoggingAndFileSync() {
	fmt.Println("Example 2: Logging + FileSync")
	fmt.Println("==============================\n")

	filename := "example2_log_sync.dat"
	defer os.Remove(filename)

	// Stack: Logging -> FileSync -> Buffer
	buffer := seekbuffer.NewEmptySeekBuffer()
	logger := log.New(os.Stdout, "   [LOG] ", 0)
	logged := seekbuffer.NewLoggingDecorator(buffer, logger, "FileSyncBuffer")
	synced := seekbuffer.NewFileSyncDecorator(logged)

	err := synced.EnableFileSync(filename)
	if err != nil {
		fmt.Printf("   Error enabling file sync: %v\n", err)
		return
	}
	defer synced.Close()

	fmt.Println("   Writing data (logged and synced to file):")
	synced.Write([]byte("Persistent logged data"))

	fmt.Printf("\n   File contains: %s\n\n", readFile(filename))
}

func example3_LoggingAndTransactions() {
	fmt.Println("Example 3: Logging + Transactions")
	fmt.Println("==================================\n")

	// Stack: Logging -> Transaction -> Buffer
	buffer := seekbuffer.NewSeekBuffer([]byte("Balance: $1000"))
	logger := log.New(os.Stdout, "   [LOG] ", 0)
	logged := seekbuffer.NewLoggingDecorator(buffer, logger, "TxBuffer")
	tx := seekbuffer.NewTransactionDecorator(logged)

	fmt.Println("   Starting transaction:")
	tx.Begin()
	tx.Write([]byte(" -> $1500"))
	fmt.Println("   Committing:")
	tx.Commit()

	fmt.Printf("\n   Final result: %s\n\n", string(buffer.Bytes()))
}

func example4_TransactionsAndFileSync() {
	fmt.Println("Example 4: Transactions + FileSync")
	fmt.Println("===================================\n")

	filename := "example4_tx_sync.dat"
	defer os.Remove(filename)

	// Stack: Transaction -> FileSync -> Buffer
	buffer := seekbuffer.NewEmptySeekBuffer()
	synced := seekbuffer.NewFileSyncDecorator(buffer)
	synced.EnableFileSync(filename)
	defer synced.Close()
	tx := seekbuffer.NewTransactionDecorator(synced)

	fmt.Println("   Transaction 1 (will commit):")
	tx.Begin()
	tx.Write([]byte("Committed data"))
	tx.Commit()
	fmt.Printf("   File after commit: %s\n", readFile(filename))

	fmt.Println("\n   Transaction 2 (will rollback):")
	tx.Begin()
	tx.Write([]byte(" + Rolled back"))
	fmt.Printf("   In transaction: %s\n", string(tx.Bytes()))
	tx.Rollback()
	fmt.Printf("   File after rollback: %s\n\n", readFile(filename))
}

func example5_AllThreeDecorators() {
	fmt.Println("Example 5: All Three Decorators Together")
	fmt.Println("=========================================\n")

	filename := "example5_all_three.dat"
	defer os.Remove(filename)

	// Stack: Transaction -> FileSync -> Logging -> Buffer
	fmt.Println("   Building the stack:")
	fmt.Println("   1. Core SeekBuffer")
	buffer := seekbuffer.NewEmptySeekBuffer()

	fmt.Println("   2. + Logging Decorator")
	logger := log.New(os.Stdout, "      [LOG] ", 0)
	logged := seekbuffer.NewLoggingDecorator(buffer, logger, "Core")

	fmt.Println("   3. + FileSync Decorator")
	synced := seekbuffer.NewFileSyncDecorator(logged)
	synced.EnableFileSync(filename)
	defer synced.Close()

	fmt.Println("   4. + Transaction Decorator")
	tx := seekbuffer.NewTransactionDecorator(synced)

	fmt.Println("\n   Stack: Transaction -> FileSync -> Logging -> Buffer")
	fmt.Println("   Features: ACID + Persistence + Logging\n")

	fmt.Println("   Performing transactional, logged, persistent write:")
	tx.Begin()
	tx.Write([]byte("All three decorators working!"))
	tx.Commit()

	fmt.Printf("\n   File contains: %s\n\n", readFile(filename))
}

func example6_StackingOrders() {
	fmt.Println("Example 6: Different Stacking Orders")
	fmt.Println("=====================================\n")

	// Order matters! Different stacking orders give different behaviors

	// Order 1: Logging outermost
	fmt.Println("   Order 1: Logging -> Transaction -> Buffer")
	fmt.Println("   Effect: Logs all transaction operations\n")
	buffer1 := seekbuffer.NewSeekBuffer([]byte("Data1"))
	tx1 := seekbuffer.NewTransactionDecorator(buffer1)
	logger1 := log.New(os.Stdout, "      [LOG1] ", 0)
	logged1 := seekbuffer.NewLoggingDecorator(tx1, logger1, "Outer")
	logged1.Write([]byte(" + More"))
	fmt.Println()

	// Order 2: Transaction outermost
	fmt.Println("   Order 2: Transaction -> Logging -> Buffer")
	fmt.Println("   Effect: Transactions control what gets logged\n")
	buffer2 := seekbuffer.NewSeekBuffer([]byte("Data2"))
	logger2 := log.New(os.Stdout, "      [LOG2] ", 0)
	logged2 := seekbuffer.NewLoggingDecorator(buffer2, logger2, "Inner")
	tx2 := seekbuffer.NewTransactionDecorator(logged2)
	tx2.Begin()
	tx2.Write([]byte(" + More"))
	tx2.Commit()
	fmt.Println()
}

func example7_RealWorldUseCase() {
	fmt.Println("Example 7: Real-World Use Case - Atomic Config Updates")
	fmt.Println("========================================================\n")

	filename := "app_config.json"
	defer os.Remove(filename)
	logfile, _ := os.Create("config_operations.log")
	defer logfile.Close()
	defer os.Remove("config_operations.log")

	// Full stack for production config management
	fmt.Println("   Scenario: Update application config with:")
	fmt.Println("   - Logging (audit trail)")
	fmt.Println("   - Transactions (rollback on error)")
	fmt.Println("   - File sync (persistence)\n")

	// Build the stack
	buffer := seekbuffer.NewEmptySeekBuffer()

	// Add logging for audit trail
	logger := log.New(logfile, "", log.LstdFlags)
	logged := seekbuffer.NewLoggingDecorator(buffer, logger, "ConfigBuffer")

	// Add file sync for persistence
	synced := seekbuffer.NewFileSyncDecorator(logged)
	synced.EnableFileSync(filename)
	defer synced.Close()

	// Add transactions for atomicity
	tx := seekbuffer.NewTransactionDecorator(synced)

	// Simulate config update
	fmt.Println("   Attempting config update...")
	tx.Begin()

	// Write new config
	newConfig := `{"version": "2.0", "debug": false, "timeout": 30}`
	tx.Write([]byte(newConfig))

	// Validate
	configValid := validateConfig(tx.Bytes())
	if configValid {
		tx.Commit()
		fmt.Println("   ✓ Config updated successfully")
		fmt.Printf("   ✓ Written to file: %s\n", filename)
		fmt.Println("   ✓ All operations logged to: config_operations.log")
	} else {
		tx.Rollback()
		fmt.Println("   ✗ Config validation failed - rolled back")
	}

	fmt.Printf("\n   Final config file: %s\n\n", readFile(filename))
}

// Helper functions

func readFile(filename string) string {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Sprintf("(error reading file: %v)", err)
	}
	if len(data) == 0 {
		return "(empty file)"
	}
	return string(data)
}

func validateConfig(data []byte) bool {
	// Simplified validation
	return len(data) > 0 && string(data) != ""
}
