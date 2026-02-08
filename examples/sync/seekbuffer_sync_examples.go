package main

import (
	"fmt"
	"log"
	"os"

	"github.com/davidul/buffers/davidul/seekbuffer"
)

// This example demonstrates how to use file synchronization with SeekBuffer
func main() {
	fmt.Println("=== SeekBuffer File Synchronization Examples ===\n")

	// Example 1: Basic file sync
	fmt.Println("Example 1: Basic File Synchronization")
	fmt.Println("-------------------------------------")
	example1()

	// Example 2: Rewind with file sync
	fmt.Println("\nExample 2: Rewind with File Sync")
	fmt.Println("---------------------------------")
	example2()

	// Example 3: Seek with file sync
	fmt.Println("\nExample 3: Seek with File Sync")
	fmt.Println("-------------------------------")
	example3()

	// Example 4: Enable sync on existing buffer
	fmt.Println("\nExample 4: Enable Sync on Existing SeekableBuffer")
	fmt.Println("-----------------------------------------")
	example4()

	// Example 5: Disable and re-enable sync
	fmt.Println("\nExample 5: Disable and Re-enable Sync")
	fmt.Println("-------------------------------------")
	example5()

	// Example 6: Mixed read/write operations
	fmt.Println("\nExample 6: Mixed Read/Write Operations")
	fmt.Println("--------------------------------------")
	example6()

	fmt.Println("\n=== Cleanup ===")
	fmt.Println("Note: Temporary files created by examples have been cleaned up.")
}

func example1() {
	filename := "example1_sync.dat"
	defer os.Remove(filename)

	// Create buffer and wrap with file sync decorator
	buffer := seekbuffer.NewEmptySeekBuffer()
	syncedBuffer := seekbuffer.NewFileSyncDecorator(buffer)
	err := syncedBuffer.EnableFileSync(filename)
	if err != nil {
		log.Fatalf("Failed to enable file sync: %v", err)
	}
	defer syncedBuffer.Close()

	// Write data - automatically synced to file
	syncedBuffer.Write([]byte("Hello, "))
	syncedBuffer.Write([]byte("World!"))

	fmt.Printf("SeekableBuffer content: %s\n", string(syncedBuffer.Bytes()))
	fmt.Printf("Sync enabled: %v\n", syncedBuffer.IsSyncEnabled())
	fmt.Printf("Sync filename: %s\n", syncedBuffer.GetSyncFilename())

	// Verify file content
	data, _ := os.ReadFile(filename)
	fmt.Printf("File content: %s\n", string(data))
}

func example2() {
	filename := "example2_rewind.dat"
	defer os.Remove(filename)

	// Create buffer with initial content and wrap with sync decorator
	buffer := seekbuffer.NewSeekBuffer([]byte("ABCDEFGHIJ"))
	syncedBuffer := seekbuffer.NewFileSyncDecorator(buffer)
	syncedBuffer.EnableFileSync(filename)
	defer syncedBuffer.Close()

	// Read first 5 bytes
	data := make([]byte, 5)
	syncedBuffer.Read(data)
	fmt.Printf("First read: %s (offset now at %d)\n", string(data), 5)

	// Rewind - resets both buffer and file position
	syncedBuffer.Rewind()
	fmt.Println("Called Rewind() - offset reset to 0")

	// Read again - should get same data
	data2 := make([]byte, 5)
	syncedBuffer.Read(data2)
	fmt.Printf("After rewind: %s\n", string(data2))

	// Verify file still has full content
	fileData, _ := os.ReadFile(filename)
	fmt.Printf("File content: %s\n", string(fileData))
}

func example3() {
	filename := "example3_seek.dat"
	defer os.Remove(filename)

	// Create buffer with data and wrap with sync decorator
	buffer := seekbuffer.NewSeekBuffer([]byte("0123456789"))
	syncedBuffer := seekbuffer.NewFileSyncDecorator(buffer)
	syncedBuffer.EnableFileSync(filename)
	defer syncedBuffer.Close()

	// Seek to position 5
	syncedBuffer.Seek(5)
	fmt.Println("Seeked to position 5")

	// Read 3 bytes from position 5
	data := make([]byte, 3)
	syncedBuffer.Read(data)
	fmt.Printf("Read from position 5: %s\n", string(data))

	// Seek to position 2
	syncedBuffer.Seek(2)
	fmt.Println("Seeked to position 2")

	// Read 4 bytes from position 2
	data2 := make([]byte, 4)
	syncedBuffer.Read(data2)
	fmt.Printf("Read from position 2: %s\n", string(data2))

	// File content remains unchanged
	fileData, _ := os.ReadFile(filename)
	fmt.Printf("File content: %s\n", string(fileData))
}

func example4() {
	filename := "example4_existing.dat"
	defer os.Remove(filename)

	// Create buffer with existing data
	buffer := seekbuffer.NewSeekBuffer([]byte("Pre-existing content"))
	fmt.Printf("SeekableBuffer before sync: %s\n", string(buffer.Bytes()))

	// Wrap with decorator and enable file sync - writes existing content to file
	syncedBuffer := seekbuffer.NewFileSyncDecorator(buffer)
	syncedBuffer.EnableFileSync(filename)
	defer syncedBuffer.Close()
	fmt.Println("File sync enabled - existing content written to file")

	// Verify file has the content
	fileData, _ := os.ReadFile(filename)
	fmt.Printf("File content: %s\n", string(fileData))

	// Add more data - also synced
	syncedBuffer.Write([]byte(" + New data"))
	fmt.Printf("After write: %s\n", string(syncedBuffer.Bytes()))

	// Verify file updated
	fileData, _ = os.ReadFile(filename)
	fmt.Printf("Updated file: %s\n", string(fileData))
}

func example5() {
	filename := "example5_toggle.dat"
	defer os.Remove(filename)

	buffer := seekbuffer.NewEmptySeekBuffer()
	syncedBuffer := seekbuffer.NewFileSyncDecorator(buffer)

	// Enable sync
	syncedBuffer.EnableFileSync(filename)
	fmt.Println("Sync enabled")
	syncedBuffer.Write([]byte("Synced data"))
	fmt.Printf("SeekableBuffer: %s\n", string(syncedBuffer.Bytes()))

	fileData1, _ := os.ReadFile(filename)
	fmt.Printf("File after sync write: %s\n", string(fileData1))

	// Disable sync
	syncedBuffer.DisableFileSync()
	fmt.Println("\nSync disabled")
	syncedBuffer.Write([]byte(" + Not synced"))
	fmt.Printf("SeekableBuffer: %s\n", string(syncedBuffer.Bytes()))

	fileData2, _ := os.ReadFile(filename)
	fmt.Printf("File (unchanged): %s\n", string(fileData2))

	// Re-enable sync to a new file
	filename2 := "example5_toggle2.dat"
	defer os.Remove(filename2)
	syncedBuffer.EnableFileSync(filename2)
	fmt.Println("\nSync re-enabled to new file")
	fmt.Printf("New file content: %s\n", string(syncedBuffer.Bytes()))

	// Verify new file has all buffer content
	fileData3, _ := os.ReadFile(filename2)
	fmt.Printf("New file: %s\n", string(fileData3))

	syncedBuffer.Close()
}

func example6() {
	filename := "example6_readwrite.dat"
	defer os.Remove(filename)

	// Create and wrap with sync decorator
	buffer := seekbuffer.NewSeekBuffer([]byte("Line 1\nLine 2\nLine 3\n"))
	syncedBuffer := seekbuffer.NewFileSyncDecorator(buffer)
	syncedBuffer.EnableFileSync(filename)
	defer syncedBuffer.Close()

	fmt.Println("Initial content synced to file")

	// Read first line
	line1, _ := syncedBuffer.ReadBytes('\n')
	fmt.Printf("Read line 1: %s", string(line1))

	// Read second line
	line2, _ := syncedBuffer.ReadBytes('\n')
	fmt.Printf("Read line 2: %s", string(line2))

	// Add a new line
	syncedBuffer.Write([]byte("Line 4\n"))
	fmt.Println("Added Line 4")

	// Rewind and read all
	syncedBuffer.Rewind()
	fmt.Println("\nRewound to beginning, reading all:")

	for i := 1; i <= 4; i++ {
		line, err := syncedBuffer.ReadBytes('\n')
		if err != nil {
			break
		}
		fmt.Printf("  %s", string(line))
	}

	// Verify file has all content
	fileData, _ := os.ReadFile(filename)
	fmt.Printf("\nFinal file content:\n%s", string(fileData))
}
