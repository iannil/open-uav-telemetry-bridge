// Package logger provides a buffered logging system with SSE streaming support
package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"sync"
	"time"
)

// Level represents log severity level
type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

// Entry represents a single log entry
type Entry struct {
	ID        int64  `json:"id"`
	Timestamp int64  `json:"timestamp"`
	Level     Level  `json:"level"`
	Source    string `json:"source"`
	Message   string `json:"message"`
}

// Subscriber receives log entries via a channel
type Subscriber struct {
	ID     string
	Filter Level
	Ch     chan *Entry
}

// Buffer is a ring buffer for log entries with subscriber support
type Buffer struct {
	entries     []Entry
	head        int
	size        int
	cap         int
	nextID      int64
	subscribers map[string]*Subscriber
	mu          sync.RWMutex
}

// New creates a new log buffer with the specified capacity
func New(capacity int) *Buffer {
	if capacity <= 0 {
		capacity = 1000
	}
	return &Buffer{
		entries:     make([]Entry, capacity),
		cap:         capacity,
		nextID:      1,
		subscribers: make(map[string]*Subscriber),
	}
}

// Write implements io.Writer interface for use with log.SetOutput
func (b *Buffer) Write(p []byte) (n int, err error) {
	// Parse the log message and add to buffer
	msg := string(p)
	if len(msg) > 0 && msg[len(msg)-1] == '\n' {
		msg = msg[:len(msg)-1]
	}

	// Try to detect level from message prefix
	level := LevelInfo
	source := "system"

	// Parse source from [Source] prefix
	if len(msg) > 2 && msg[0] == '[' {
		for i := 1; i < len(msg); i++ {
			if msg[i] == ']' {
				source = msg[1:i]
				msg = msg[i+1:]
				if len(msg) > 0 && msg[0] == ' ' {
					msg = msg[1:]
				}
				break
			}
		}
	}

	// Detect level from keywords
	switch {
	case containsIgnoreCase(msg, "error"):
		level = LevelError
	case containsIgnoreCase(msg, "warn"):
		level = LevelWarn
	case containsIgnoreCase(msg, "debug"):
		level = LevelDebug
	}

	b.Add(level, source, msg)
	return len(p), nil
}

// containsIgnoreCase checks if s contains substr (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	if len(substr) > len(s) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			c1, c2 := s[i+j], substr[j]
			// Simple lowercase comparison for ASCII
			if c1 >= 'A' && c1 <= 'Z' {
				c1 += 32
			}
			if c2 >= 'A' && c2 <= 'Z' {
				c2 += 32
			}
			if c1 != c2 {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// Add adds a new log entry to the buffer
func (b *Buffer) Add(level Level, source, message string) *Entry {
	b.mu.Lock()

	entry := Entry{
		ID:        b.nextID,
		Timestamp: time.Now().UnixMilli(),
		Level:     level,
		Source:    source,
		Message:   message,
	}
	b.nextID++

	b.entries[b.head] = entry
	b.head = (b.head + 1) % b.cap

	if b.size < b.cap {
		b.size++
	}

	// Get subscribers to notify
	subs := make([]*Subscriber, 0, len(b.subscribers))
	for _, sub := range b.subscribers {
		subs = append(subs, sub)
	}

	b.mu.Unlock()

	// Notify subscribers (non-blocking)
	for _, sub := range subs {
		if shouldSend(entry.Level, sub.Filter) {
			select {
			case sub.Ch <- &entry:
			default:
				// Channel full, skip this entry
			}
		}
	}

	return &entry
}

// shouldSend checks if an entry with given level should be sent to a subscriber with the given filter
func shouldSend(entryLevel, filterLevel Level) bool {
	levels := map[Level]int{
		LevelDebug: 0,
		LevelInfo:  1,
		LevelWarn:  2,
		LevelError: 3,
	}

	entryPriority, ok1 := levels[entryLevel]
	filterPriority, ok2 := levels[filterLevel]

	if !ok1 || !ok2 {
		return true // Unknown levels, send anyway
	}

	return entryPriority >= filterPriority
}

// GetLast returns the last N entries in chronological order
func (b *Buffer) GetLast(n int) []Entry {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if n > b.size {
		n = b.size
	}
	if n == 0 {
		return []Entry{}
	}

	result := make([]Entry, n)
	start := (b.head - n + b.cap) % b.cap

	for i := 0; i < n; i++ {
		idx := (start + i) % b.cap
		result[i] = b.entries[idx]
	}

	return result
}

// GetSince returns all entries since the given ID
func (b *Buffer) GetSince(sinceID int64) []Entry {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.size == 0 {
		return []Entry{}
	}

	var result []Entry
	start := 0
	if b.size == b.cap {
		start = b.head
	}

	for i := 0; i < b.size; i++ {
		idx := (start + i) % b.cap
		if b.entries[idx].ID > sinceID {
			result = append(result, b.entries[idx])
		}
	}

	return result
}

// GetFiltered returns entries filtered by level and optionally source
func (b *Buffer) GetFiltered(level Level, source string, limit int) []Entry {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.size == 0 {
		return []Entry{}
	}

	var result []Entry
	start := 0
	if b.size == b.cap {
		start = b.head
	}

	for i := 0; i < b.size; i++ {
		idx := (start + i) % b.cap
		entry := b.entries[idx]

		if !shouldSend(entry.Level, level) {
			continue
		}

		if source != "" && entry.Source != source {
			continue
		}

		result = append(result, entry)

		if limit > 0 && len(result) >= limit {
			break
		}
	}

	return result
}

// Subscribe creates a new subscriber for real-time log streaming
func (b *Buffer) Subscribe(id string, filter Level) *Subscriber {
	b.mu.Lock()
	defer b.mu.Unlock()

	sub := &Subscriber{
		ID:     id,
		Filter: filter,
		Ch:     make(chan *Entry, 100),
	}

	b.subscribers[id] = sub
	return sub
}

// Unsubscribe removes a subscriber
func (b *Buffer) Unsubscribe(id string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if sub, ok := b.subscribers[id]; ok {
		close(sub.Ch)
		delete(b.subscribers, id)
	}
}

// Size returns the current number of entries
func (b *Buffer) Size() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.size
}

// Clear removes all entries
func (b *Buffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.head = 0
	b.size = 0
}

// Logger helpers for common use cases

// Info logs an info level message
func (b *Buffer) Info(source, format string, args ...interface{}) {
	b.Add(LevelInfo, source, fmt.Sprintf(format, args...))
}

// Debug logs a debug level message
func (b *Buffer) Debug(source, format string, args ...interface{}) {
	b.Add(LevelDebug, source, fmt.Sprintf(format, args...))
}

// Warn logs a warning level message
func (b *Buffer) Warn(source, format string, args ...interface{}) {
	b.Add(LevelWarn, source, fmt.Sprintf(format, args...))
}

// Error logs an error level message
func (b *Buffer) Error(source, format string, args ...interface{}) {
	b.Add(LevelError, source, fmt.Sprintf(format, args...))
}

// MultiWriter combines the buffer with stdout for dual output
type MultiWriter struct {
	buffer *Buffer
	stdout io.Writer
}

// NewMultiWriter creates a writer that outputs to both buffer and stdout
func NewMultiWriter(buffer *Buffer, stdout io.Writer) *MultiWriter {
	return &MultiWriter{
		buffer: buffer,
		stdout: stdout,
	}
}

// Write implements io.Writer
func (mw *MultiWriter) Write(p []byte) (n int, err error) {
	// Write to stdout first
	if mw.stdout != nil {
		mw.stdout.Write(p)
	}
	// Then write to buffer
	return mw.buffer.Write(p)
}

// SetupGlobalLogger configures the standard log package to use the buffer
func SetupGlobalLogger(buffer *Buffer, stdout io.Writer) {
	mw := NewMultiWriter(buffer, stdout)
	log.SetOutput(mw)
	log.SetFlags(log.Ldate | log.Ltime)
}

// MarshalJSON returns JSON representation of entries
func (e *Entry) MarshalJSON() ([]byte, error) {
	type alias Entry
	return json.Marshal((*alias)(e))
}
