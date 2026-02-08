package main

import (
	"fmt"
	"log"

	"github.com/davidul/buffers/davidul/seekbuffer"
)

// This example demonstrates the different ways to append buffer data to files
func main() {
	fmt.Println("=== SeekBuffer Append Functionality Examples ===\n")

	// Example 1: Basic AppendToFile - appending entire buffer
	fmt.Println("Example 1: Basic AppendToFile")
	fmt.Println("-------------------------------")

	logFile := "application.log"

	// First log entry
	log1 := seekbuffer.NewSeekBuffer([]byte("[INFO] Application started\n"))
	err := log1.SaveToFile(logFile)
	if err != nil {
		log.Fatalf("Failed to save: %v", err)
	}

	// Second log entry - append to existing file
	log2 := seekbuffer.NewSeekBuffer([]byte("[INFO] Processing data...\n"))
	err = log2.AppendToFile(logFile)
	if err != nil {
		log.Fatalf("Failed to append: %v", err)
	}

	// Third log entry - append more
	log3 := seekbuffer.NewSeekBuffer([]byte("[INFO] Process completed\n"))
	err = log3.AppendToFile(logFile)
	if err != nil {
		log.Fatalf("Failed to append: %v", err)
	}

	fmt.Println("Created log file with multiple entries")
	fmt.Println("File content:")
	displayFile(logFile)
	fmt.Println()

	// Example 2: AppendUnreadToFile - appending only unread portion
	fmt.Println("Example 2: AppendUnreadToFile - Selective Append")
	fmt.Println("------------------------------------------------")

	// Create a buffer with processed and unprocessed data
	buffer := seekbuffer.NewSeekBuffer([]byte("PROCESSED_DATA|UNPROCESSED_DATA"))

	// Simulate reading the processed part
	processed := make([]byte, 15)
	buffer.Read(processed)
	fmt.Printf("Read (processed): %s\n", string(processed))
	fmt.Printf("Remaining (unprocessed): %s\n", string(buffer.Bytes()[15:]))

	// Only append the unread/unprocessed portion
	unprocessedFile := "unprocessed.dat"
	err = buffer.AppendUnreadToFile(unprocessedFile)
	if err != nil {
		log.Fatalf("Failed to append unread: %v", err)
	}

	fmt.Println("\nAppended only unread portion to file:")
	displayFile(unprocessedFile)
	fmt.Println()

	// Example 3: Incremental logging pattern
	fmt.Println("Example 3: Incremental Logging Pattern")
	fmt.Println("--------------------------------------")

	eventLog := "events.log"

	// Simulate multiple events being logged incrementally
	events := []string{
		"[2026-02-08 10:00:00] User login: alice",
		"[2026-02-08 10:05:23] Database query executed",
		"[2026-02-08 10:10:45] File uploaded: document.pdf",
		"[2026-02-08 10:15:12] Email sent to customer",
		"[2026-02-08 10:20:33] User logout: alice",
	}

	for i, event := range events {
		buffer := seekbuffer.NewSeekBuffer([]byte(event + "\n"))
		err := buffer.AppendToFile(eventLog)
		if err != nil {
			log.Fatalf("Failed to append event %d: %v", i, err)
		}
		fmt.Printf("✓ Appended: %s\n", event)
	}

	fmt.Println("\nFinal event log:")
	displayFile(eventLog)
	fmt.Println()

	// Example 4: Difference between SaveToFile and AppendToFile
	fmt.Println("Example 4: SaveToFile vs AppendToFile")
	fmt.Println("-------------------------------------")

	testFile := "comparison.txt"

	buf1 := seekbuffer.NewSeekBuffer([]byte("First content\n"))
	buf1.SaveToFile(testFile)
	fmt.Println("After SaveToFile('First content'):")
	displayFile(testFile)

	buf2 := seekbuffer.NewSeekBuffer([]byte("Second content\n"))
	buf2.AppendToFile(testFile)
	fmt.Println("After AppendToFile('Second content'):")
	displayFile(testFile)

	buf3 := seekbuffer.NewSeekBuffer([]byte("Third content (overwrite)\n"))
	buf3.SaveToFile(testFile)
	fmt.Println("After SaveToFile('Third content (overwrite)'):")
	displayFile(testFile)

	buf4 := seekbuffer.NewSeekBuffer([]byte("Fourth content\n"))
	buf4.AppendToFile(testFile)
	fmt.Println("After AppendToFile('Fourth content'):")
	displayFile(testFile)
	fmt.Println()

	// Example 5: Building a file incrementally with data batches
	fmt.Println("Example 5: Batch Data Processing")
	fmt.Println("---------------------------------")

	outputFile := "data_batches.txt"

	// Process and append data in batches
	for batchNum := 1; batchNum <= 3; batchNum++ {
		buffer := seekbuffer.NewEmptySeekBuffer()

		// Simulate processing a batch of data
		buffer.Write([]byte(fmt.Sprintf("=== Batch %d ===\n", batchNum)))
		for i := 1; i <= 5; i++ {
			buffer.Write([]byte(fmt.Sprintf("  Item %d.%d: Data value\n", batchNum, i)))
		}
		buffer.Write([]byte("\n"))

		// Append the entire batch to file
		err := buffer.AppendToFile(outputFile)
		if err != nil {
			log.Fatalf("Failed to append batch %d: %v", batchNum, err)
		}

		fmt.Printf("✓ Processed and appended batch %d\n", batchNum)
	}

	fmt.Println("\nFinal output file:")
	displayFile(outputFile)

	// Example 6: Using AppendUnreadToFile with streaming data
	fmt.Println("\nExample 6: Streaming with Partial Reads")
	fmt.Println("---------------------------------------")

	streamFile := "stream.dat"

	// Simulate receiving a stream of data
	streamBuffer := seekbuffer.NewSeekBuffer([]byte("HEADER:12345|DATA:ABCDEFGHIJKLMNOP|FOOTER:END"))

	// Read and process header
	header, _ := streamBuffer.ReadBytes('|')
	fmt.Printf("Processed header: %s\n", string(header))

	// Read and process data section
	data, _ := streamBuffer.ReadBytes('|')
	fmt.Printf("Processed data: %s\n", string(data))

	// Append only the remaining unread part (footer)
	fmt.Println("Appending only unread portion (footer) to file...")
	streamBuffer.AppendUnreadToFile(streamFile)

	fmt.Println("Saved to file:")
	displayFile(streamFile)

	fmt.Println("\n=== Cleanup ===")
	fmt.Println("Note: In production, remember to manage file cleanup appropriately.")
}

func displayFile(filename string) {
	buffer, err := seekbuffer.NewSeekBufferFromFile(filename)
	if err != nil {
		fmt.Printf("  (Could not read file: %v)\n", err)
		return
	}
	fmt.Printf("  %s\n", string(buffer.Bytes()))
}
