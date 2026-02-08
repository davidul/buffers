package main

import (
	"log"
	"os"

	"github.com/davidul/buffers/davidul/seekbuffer"
)

func main() {

	buffer := seekbuffer.NewSeekBuffer([]byte("Data"))
	logger := log.New(os.Stdout, "[BUFFER] ", log.LstdFlags)
	logged := seekbuffer.NewLoggingDecorator(buffer, logger, "MyBuffer")

	// All operations automatically logged
	logged.Write([]byte(" + More"))
	// Output: [BUFFER] 2026/02/08 12:34:56 [MyBuffer] Write: 6 bytes, duration: 1.2µs

	logged.Seek(5)
	// Output: [BUFFER] 2026/02/08 12:34:57 [MyBuffer] Seek: moved to offset 5, duration: 0.8µs
}
