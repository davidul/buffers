# Seekable buffer
Allows to read many times.

New empty seek buffer is created
```go
b := NewEmptySeekBuffer()
```

Append data to empty buffer
```go
b.Append([]byte("hello"))
```