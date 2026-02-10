package linkedlist

import (
	"bytes"
	"encoding/binary"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMap_Add(t *testing.T) {
	log := NewLog()
	addFunc := func(key uint64, value []byte, log *AppendOnly) {
		log.Add(key, value)
		assert.Equal(t, log.Get(key), value)
	}

	t.Run("Add", func(t *testing.T) {
		addFunc(1, []byte("ABCDS"), log)
		addFunc(2, []byte("XYZ"), log)
	})
}

func TestMap_Add_1(t *testing.T) {
	now, _ := time.Now().MarshalBinary()
	log := NewLog()

	log.Add(1, now)
	log.Add(2, []byte("David"))

	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, math.Pi)

	if err != nil {
		assert.Fail(t, err.Error())
	}

	log.Add(3, buf.Bytes())

	assert.Equal(t, now, log.Get(1))
	assert.Equal(t, []byte("David"), log.Get(2))

	buf2 := new(bytes.Buffer)
	err = binary.Write(buf2, binary.BigEndian, math.Pi)
	if err != nil {
		assert.Fail(t, err.Error())
	}
	assert.Equal(t, buf2.Bytes(), log.Get(3))
}

func TestMap_Types(t *testing.T) {
	log := NewLog()
	//bytes
	log.Add(1, []byte{1, 2, 3, 4})
	//ints
	bytes32 := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes32, 124)
	log.Add(2, bytes32)
	bytes64 := make([]byte, 8)
	binary.LittleEndian.PutUint64(bytes64, 123423234)
	log.Add(3, bytes64)
	// string/rune
	log.Add(4, []byte("Hello World"))

	assert.Equal(t, log.Get(1), []byte{1, 2, 3, 4})
	assert.Equal(t, log.Get(2), bytes32)
	assert.Equal(t, log.Get(3), bytes64)
	assert.Equal(t, log.Get(4), []byte("Hello World"))
}

// TestConcurrentWrites verifies that multiple goroutines can safely write to the log
func TestConcurrentWrites(t *testing.T) {
	log := NewLog()
	numGoroutines := 5
	writesPerGoroutine := 3

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Each goroutine writes multiple entries
	// Each entry: 8 (key) + 8 (timestamp) + 4 (length) + 2 (value) = 22 bytes
	// Total: 5 * 3 * 22 = 330 bytes (well under 1024)
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < writesPerGoroutine; j++ {
				key := uint64(goroutineID*100 + j)
				value := []byte{byte(goroutineID), byte(j)}
				log.Add(key, value)
			}
		}(i)
	}

	wg.Wait()

	// Verify all entries were written correctly
	for i := 0; i < numGoroutines; i++ {
		for j := 0; j < writesPerGoroutine; j++ {
			key := uint64(i*100 + j)
			expected := []byte{byte(i), byte(j)}
			actual := log.Get(key)
			assert.Equal(t, expected, actual, "Mismatch for goroutine %d, write %d", i, j)
		}
	}
}

// TestConcurrentReads verifies that multiple goroutines can safely read from the log
func TestConcurrentReads(t *testing.T) {
	log := NewLog()

	// Prepopulate the log with small entries
	// Each entry: 22 bytes, 20 entries = 440 bytes
	numEntries := 20
	for i := uint64(0); i < uint64(numEntries); i++ {
		log.Add(i, []byte{byte(i)})
	}

	numGoroutines := 10
	readsPerGoroutine := 50

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Multiple goroutines reading concurrently
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < readsPerGoroutine; j++ {
				key := uint64(j % numEntries)
				value := log.Get(key)
				expected := []byte{byte(key)}
				assert.Equal(t, expected, value)
			}
		}()
	}

	wg.Wait()
}

// TestConcurrentReadWrite verifies that reads and writes can happen concurrently
func TestConcurrentReadWrite(t *testing.T) {
	log := NewLog()

	// Prepopulate with some initial data (10 entries = 220 bytes)
	for i := uint64(0); i < 10; i++ {
		log.Add(i, []byte{byte(i)})
	}

	numReaders := 5
	numWriters := 3
	readsPerReader := 30
	writesPerWriter := 3

	var wg sync.WaitGroup
	wg.Add(numReaders + numWriters)

	// Start reader goroutines
	for i := 0; i < numReaders; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < readsPerReader; j++ {
				key := uint64(j % 10)
				log.Get(key) // Just verify no panic occurs
				time.Sleep(time.Microsecond)
			}
		}()
	}

	// Start writer goroutines (9 more entries = 198 bytes, total ~418 bytes)
	for i := 0; i < numWriters; i++ {
		go func(writerID int) {
			defer wg.Done()
			for j := 0; j < writesPerWriter; j++ {
				key := uint64(100 + writerID*10 + j)
				value := []byte{byte(writerID), byte(j)}
				log.Add(key, value)
				time.Sleep(time.Microsecond)
			}
		}(i)
	}

	wg.Wait()

	// Verify the written data
	for i := 0; i < numWriters; i++ {
		for j := 0; j < writesPerWriter; j++ {
			key := uint64(100 + i*10 + j)
			expected := []byte{byte(i), byte(j)}
			actual := log.Get(key)
			assert.Equal(t, expected, actual)
		}
	}
}

// TestConcurrentDuplicateKeys verifies that concurrent writes with duplicate keys are handled safely
func TestConcurrentDuplicateKeys(t *testing.T) {
	log := NewLog()
	numGoroutines := 5

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	keys := make([]uint64, numGoroutines)

	// All goroutines try to add with the same key (1)
	// 5 entries = 110 bytes
	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			value := []byte{byte(goroutineID)}
			// All try to use key 1, but only the first should succeed
			// Others will get auto-generated keys
			key := log.Add(1, value)
			keys[goroutineID] = key
		}(i)
	}

	wg.Wait()

	// Verify that all keys are unique and all values are retrievable
	keySet := make(map[uint64]bool)
	for i := 0; i < numGoroutines; i++ {
		key := keys[i]
		assert.False(t, keySet[key], "Duplicate key generated: %d", key)
		keySet[key] = true

		value := log.Get(key)
		assert.NotNil(t, value)
	}
}

// TestIterator_EmptyLog verifies iterator behavior on an empty log
func TestIterator_EmptyLog(t *testing.T) {
	log := NewLog()
	iter := log.NewIterator()

	// Should not have any records
	assert.False(t, iter.Next())
	assert.Nil(t, iter.Value())
	assert.Nil(t, iter.Error())
}

// TestIterator_SingleRecord verifies iterator on a log with one record
func TestIterator_SingleRecord(t *testing.T) {
	log := NewLog()
	log.Add(1, []byte("Hello"))

	iter := log.NewIterator()

	// Should have exactly one record
	assert.True(t, iter.Next())
	record := iter.Value()
	assert.NotNil(t, record)
	assert.Equal(t, uint64(1), record.Key)
	assert.Equal(t, []byte("Hello"), record.Value)
	assert.Greater(t, record.Timestamp, uint64(0))

	// Should not have more records
	assert.False(t, iter.Next())
	assert.Nil(t, iter.Error())
}

// TestIterator_MultipleRecords verifies iterator on a log with multiple records
func TestIterator_MultipleRecords(t *testing.T) {
	log := NewLog()

	// Add multiple records
	keys := []uint64{1, 2, 3, 4, 5}
	values := [][]byte{
		[]byte("First"),
		[]byte("Second"),
		[]byte("Third"),
		[]byte("Fourth"),
		[]byte("Fifth"),
	}

	for i := 0; i < len(keys); i++ {
		log.Add(keys[i], values[i])
	}

	iter := log.NewIterator()

	// Iterate through all records
	count := 0
	for iter.Next() {
		record := iter.Value()
		assert.NotNil(t, record)
		assert.Equal(t, keys[count], record.Key)
		assert.Equal(t, values[count], record.Value)
		count++
	}

	assert.Equal(t, len(keys), count)
	assert.Nil(t, iter.Error())
}

// TestIterator_AllFieldsDecoded verifies all fields are properly decoded
func TestIterator_AllFieldsDecoded(t *testing.T) {
	log := NewLog()

	beforeAdd := time.Now().UnixNano()
	log.Add(42, []byte{1, 2, 3, 4})
	afterAdd := time.Now().UnixNano()

	iter := log.NewIterator()
	assert.True(t, iter.Next())

	record := iter.Value()
	assert.Equal(t, uint64(42), record.Key)
	assert.Equal(t, []byte{1, 2, 3, 4}, record.Value)
	// Verify timestamp is within reasonable range
	assert.GreaterOrEqual(t, record.Timestamp, uint64(beforeAdd))
	assert.LessOrEqual(t, record.Timestamp, uint64(afterAdd))
}

// TestIterator_Snapshot verifies iterator uses snapshot semantics
func TestIterator_Snapshot(t *testing.T) {
	log := NewLog()

	// Add initial records
	log.Add(1, []byte("First"))
	log.Add(2, []byte("Second"))

	// Create iterator (takes snapshot)
	iter := log.NewIterator()

	// Add more records after iterator creation
	log.Add(3, []byte("Third"))
	log.Add(4, []byte("Fourth"))

	// Iterator should only see the first 2 records (snapshot)
	count := 0
	for iter.Next() {
		count++
	}

	assert.Equal(t, 2, count, "Iterator should only see records that existed at creation time")
	assert.Nil(t, iter.Error())
}

// TestIterator_ConcurrentIteration verifies multiple iterators can work concurrently
func TestIterator_ConcurrentIteration(t *testing.T) {
	log := NewLog()

	// Add some records (5 records = 110 bytes)
	for i := uint64(0); i < 5; i++ {
		log.Add(i, []byte{byte(i)})
	}

	numIterators := 5
	var wg sync.WaitGroup
	wg.Add(numIterators)

	results := make([]int, numIterators)

	// Multiple iterators reading concurrently
	for i := 0; i < numIterators; i++ {
		go func(iteratorID int) {
			defer wg.Done()

			iter := log.NewIterator()
			count := 0

			for iter.Next() {
				record := iter.Value()
				assert.NotNil(t, record)
				assert.Equal(t, []byte{byte(count)}, record.Value)
				count++
			}

			assert.Nil(t, iter.Error())
			results[iteratorID] = count
		}(i)
	}

	wg.Wait()

	// All iterators should have seen the same number of records
	for i := 0; i < numIterators; i++ {
		assert.Equal(t, 5, results[i], "Iterator %d should have seen 5 records", i)
	}
}

// TestIterator_WithVariableLengthValues verifies iterator handles different value sizes
func TestIterator_WithVariableLengthValues(t *testing.T) {
	log := NewLog()

	// Add records with different value lengths
	log.Add(1, []byte("A"))
	log.Add(2, []byte("AB"))
	log.Add(3, []byte("ABC"))
	log.Add(4, []byte("ABCD"))

	iter := log.NewIterator()

	expectedValues := [][]byte{
		[]byte("A"),
		[]byte("AB"),
		[]byte("ABC"),
		[]byte("ABCD"),
	}

	count := 0
	for iter.Next() {
		record := iter.Value()
		assert.Equal(t, expectedValues[count], record.Value)
		count++
	}

	assert.Equal(t, 4, count)
	assert.Nil(t, iter.Error())
}

// TestIterator_MultipleIterationsOnSameIterator verifies iterator can't be reused
func TestIterator_MultipleIterationsOnSameIterator(t *testing.T) {
	log := NewLog()
	log.Add(1, []byte("Test"))

	iter := log.NewIterator()

	// First iteration
	assert.True(t, iter.Next())
	assert.False(t, iter.Next())

	// Iterator is exhausted, can't iterate again
	assert.False(t, iter.Next())
}

// TestIterator_ConcurrentIterationAndWrites verifies iteration while writes are happening
func TestIterator_ConcurrentIterationAndWrites(t *testing.T) {
	log := NewLog()

	// Add initial records (3 records = 66 bytes)
	for i := uint64(0); i < 3; i++ {
		log.Add(i, []byte{byte(i)})
	}

	var wg sync.WaitGroup
	wg.Add(2)

	iteratorResults := 0

	// Start iterator
	go func() {
		defer wg.Done()
		iter := log.NewIterator()
		count := 0
		for iter.Next() {
			count++
			time.Sleep(time.Microsecond)
		}
		iteratorResults = count
		assert.Nil(t, iter.Error())
	}()

	// Start writer (adds 2 more records = 44 bytes, total ~110 bytes)
	go func() {
		defer wg.Done()
		time.Sleep(time.Microsecond * 5)
		log.Add(100, []byte{100})
		log.Add(101, []byte{101})
	}()

	wg.Wait()

	// Iterator should see snapshot (3 records), not the newly added ones
	assert.Equal(t, 3, iteratorResults)
}
