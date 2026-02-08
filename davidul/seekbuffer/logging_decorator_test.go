package seekbuffer

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func TestLoggingDecoratorWrite(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	buffer := NewEmptySeekBuffer()
	logged := NewLoggingDecorator(buffer, logger, "TestBuffer")

	logged.Write([]byte("Hello"))

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "Write: 5 bytes") {
		t.Errorf("Expected write log, got: %s", logOutput)
	}
}

func TestLoggingDecoratorRead(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	buffer := NewSeekBuffer([]byte("Hello World"))
	logged := NewLoggingDecorator(buffer, logger, "TestBuffer")

	data := make([]byte, 5)
	logged.Read(data)

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "Read: 5 bytes") {
		t.Errorf("Expected read log, got: %s", logOutput)
	}
}

func TestLoggingDecoratorAppend(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	buffer := NewSeekBuffer([]byte("Hello"))
	logged := NewLoggingDecorator(buffer, logger, "TestBuffer")

	logged.Append([]byte(" World"))

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "Append: 6 bytes") {
		t.Errorf("Expected append log, got: %s", logOutput)
	}
}

func TestLoggingDecoratorSeek(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	buffer := NewSeekBuffer([]byte("Hello World"))
	logged := NewLoggingDecorator(buffer, logger, "TestBuffer")

	logged.Seek(5)

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "Seek: moved to offset 5") {
		t.Errorf("Expected seek log, got: %s", logOutput)
	}
}

func TestLoggingDecoratorRewind(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	buffer := NewSeekBuffer([]byte("Hello"))
	logged := NewLoggingDecorator(buffer, logger, "TestBuffer")

	logged.Rewind()

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "Rewind: offset reset to 0") {
		t.Errorf("Expected rewind log, got: %s", logOutput)
	}
}

func TestLoggingDecoratorReadBytes(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	buffer := NewSeekBuffer([]byte("Hello\nWorld"))
	logged := NewLoggingDecorator(buffer, logger, "TestBuffer")

	logged.ReadBytes('\n')

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "ReadBytes") && !strings.Contains(logOutput, "read 6 bytes") {
		t.Errorf("Expected readbytes log, got: %s", logOutput)
	}
}

func TestLoggingDecoratorBytes(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	buffer := NewSeekBuffer([]byte("Hello"))
	logged := NewLoggingDecorator(buffer, logger, "TestBuffer")

	logged.Bytes()

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "Bytes: retrieved 5 bytes") {
		t.Errorf("Expected bytes log, got: %s", logOutput)
	}
}

func TestLoggingDecoratorLen(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	buffer := NewSeekBuffer([]byte("Hello"))
	logged := NewLoggingDecorator(buffer, logger, "TestBuffer")

	logged.Len()

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "Len: 5 unread bytes") {
		t.Errorf("Expected len log, got: %s", logOutput)
	}
}

func TestLoggingDecoratorClose(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	buffer := NewSeekBuffer([]byte("Hello"))
	logged := NewLoggingDecorator(buffer, logger, "TestBuffer")

	logged.Close()

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "Close: success") {
		t.Errorf("Expected close log, got: %s", logOutput)
	}
}

func TestLoggingDecoratorCustomName(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	buffer := NewEmptySeekBuffer()
	logged := NewLoggingDecorator(buffer, logger, "MyCustomBuffer")

	logged.Write([]byte("test"))

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "[MyCustomBuffer]") {
		t.Errorf("Expected custom name in log, got: %s", logOutput)
	}
}

func TestLoggingDecoratorSetLogger(t *testing.T) {
	var logBuf1 bytes.Buffer
	var logBuf2 bytes.Buffer
	logger1 := log.New(&logBuf1, "", 0)
	logger2 := log.New(&logBuf2, "PREFIX: ", 0)

	buffer := NewEmptySeekBuffer()
	logged := NewLoggingDecorator(buffer, logger1, "TestBuffer")

	logged.Write([]byte("first"))

	// Change logger
	logged.SetLogger(logger2)
	logged.Write([]byte("second"))

	if !strings.Contains(logBuf1.String(), "first") {
		t.Error("First logger should contain 'first' write")
	}

	if !strings.Contains(logBuf2.String(), "PREFIX:") {
		t.Error("Second logger should have prefix")
	}
}

func TestLoggingDecoratorSetName(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	buffer := NewEmptySeekBuffer()
	logged := NewLoggingDecorator(buffer, logger, "OldName")

	logged.SetName("NewName")
	logged.Write([]byte("test"))

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "[NewName]") {
		t.Errorf("Expected new name in log, got: %s", logOutput)
	}
}

func TestLoggingDecoratorSummary(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	buffer := NewSeekBuffer([]byte("Hello World"))
	logged := NewLoggingDecorator(buffer, logger, "TestBuffer")

	// Read some bytes
	data := make([]byte, 5)
	logged.Read(data)

	// Log summary
	logBuf.Reset()
	logged.LogSummary()

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "Summary") {
		t.Errorf("Expected summary log, got: %s", logOutput)
	}
	if !strings.Contains(logOutput, "total=11") {
		t.Errorf("Expected total=11 in summary, got: %s", logOutput)
	}
}

func TestLoggingDecoratorCustomMessage(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	buffer := NewEmptySeekBuffer()
	logged := NewLoggingDecorator(buffer, logger, "TestBuffer")

	logged.LogWithMessage("Custom message: %s", "test data")

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "Custom message: test data") {
		t.Errorf("Expected custom message, got: %s", logOutput)
	}
}

func TestLoggingDecoratorMultipleOperations(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	buffer := NewEmptySeekBuffer()
	logged := NewLoggingDecorator(buffer, logger, "TestBuffer")

	logged.Write([]byte("Line 1\n"))
	logged.Write([]byte("Line 2\n"))
	logged.Rewind()
	logged.ReadBytes('\n')
	logged.Seek(7)
	logged.ReadBytes('\n')

	logOutput := logBuf.String()

	// Check all operations are logged
	operations := []string{"Write", "Rewind", "ReadBytes", "Seek"}
	for _, op := range operations {
		if !strings.Contains(logOutput, op) {
			t.Errorf("Expected %s in log output", op)
		}
	}
}

func TestLoggingDecoratorDefaultLogger(t *testing.T) {
	buffer := NewEmptySeekBuffer()
	logged := NewLoggingDecorator(buffer, nil, "TestBuffer")

	// Should not panic with nil logger (uses default)
	logged.Write([]byte("test"))

	if logged.GetLogger() == nil {
		t.Error("Expected default logger to be set")
	}
}

func TestLoggingDecoratorDefaultName(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	buffer := NewEmptySeekBuffer()
	logged := NewLoggingDecorator(buffer, logger, "")

	logged.Write([]byte("test"))

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "[SeekBuffer]") {
		t.Errorf("Expected default name [SeekBuffer] in log, got: %s", logOutput)
	}
}

func TestLoggingDecoratorWithTransactions(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	buffer := NewSeekBuffer([]byte("Initial"))
	logged := NewLoggingDecorator(buffer, logger, "TxBuffer")
	tx := NewTransactionDecorator(logged)

	tx.Begin()
	tx.Write([]byte(" + Changes"))
	tx.Commit()

	logOutput := logBuf.String()

	// Should see logs for the write operation
	if !strings.Contains(logOutput, "Write") {
		t.Errorf("Expected write logs in transaction, got: %s", logOutput)
	}
}

func TestLoggingDecoratorWithFileSync(t *testing.T) {
	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)

	buffer := NewEmptySeekBuffer()
	logged := NewLoggingDecorator(buffer, logger, "SyncBuffer")
	synced := NewFileSyncDecorator(logged)

	// Note: Not actually enabling file sync in test to avoid file I/O
	synced.Write([]byte("test"))

	logOutput := logBuf.String()
	if !strings.Contains(logOutput, "Write") {
		t.Errorf("Expected write log with file sync, got: %s", logOutput)
	}
}
