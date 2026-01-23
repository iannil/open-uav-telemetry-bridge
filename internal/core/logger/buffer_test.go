package logger

import (
	"bytes"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		capacity int
		wantCap  int
	}{
		{"positive capacity", 100, 100},
		{"zero capacity defaults to 1000", 0, 1000},
		{"negative capacity defaults to 1000", -5, 1000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := New(tt.capacity)
			if buf.cap != tt.wantCap {
				t.Errorf("New(%d).cap = %d, want %d", tt.capacity, buf.cap, tt.wantCap)
			}
			if buf.size != 0 {
				t.Errorf("New buffer should have size 0, got %d", buf.size)
			}
			if buf.nextID != 1 {
				t.Errorf("New buffer should have nextID 1, got %d", buf.nextID)
			}
		})
	}
}

func TestBuffer_Add(t *testing.T) {
	buf := New(10)

	entry := buf.Add(LevelInfo, "test", "hello world")

	if entry == nil {
		t.Fatal("Add should return entry")
	}
	if entry.ID != 1 {
		t.Errorf("First entry ID should be 1, got %d", entry.ID)
	}
	if entry.Level != LevelInfo {
		t.Errorf("Entry level should be info, got %s", entry.Level)
	}
	if entry.Source != "test" {
		t.Errorf("Entry source should be 'test', got %s", entry.Source)
	}
	if entry.Message != "hello world" {
		t.Errorf("Entry message should be 'hello world', got %s", entry.Message)
	}
	if entry.Timestamp == 0 {
		t.Error("Entry timestamp should not be 0")
	}
	if buf.Size() != 1 {
		t.Errorf("Buffer size should be 1, got %d", buf.Size())
	}
}

func TestBuffer_RingBuffer(t *testing.T) {
	buf := New(3)

	buf.Add(LevelInfo, "test", "msg1")
	buf.Add(LevelInfo, "test", "msg2")
	buf.Add(LevelInfo, "test", "msg3")
	buf.Add(LevelInfo, "test", "msg4") // Should overwrite msg1

	if buf.Size() != 3 {
		t.Errorf("Buffer size should be 3, got %d", buf.Size())
	}

	entries := buf.GetLast(3)
	if len(entries) != 3 {
		t.Fatalf("GetLast(3) should return 3 entries, got %d", len(entries))
	}

	// msg1 should be gone, entries should be msg2, msg3, msg4
	if entries[0].Message != "msg2" {
		t.Errorf("First entry should be 'msg2', got '%s'", entries[0].Message)
	}
	if entries[2].Message != "msg4" {
		t.Errorf("Last entry should be 'msg4', got '%s'", entries[2].Message)
	}
}

func TestBuffer_GetLast(t *testing.T) {
	buf := New(10)

	// Empty buffer
	entries := buf.GetLast(5)
	if len(entries) != 0 {
		t.Errorf("GetLast on empty buffer should return empty slice, got %d", len(entries))
	}

	buf.Add(LevelInfo, "test", "msg1")
	buf.Add(LevelInfo, "test", "msg2")

	// Request more than available
	entries = buf.GetLast(10)
	if len(entries) != 2 {
		t.Errorf("GetLast(10) with 2 entries should return 2, got %d", len(entries))
	}

	// Request exact amount
	entries = buf.GetLast(2)
	if len(entries) != 2 {
		t.Errorf("GetLast(2) should return 2, got %d", len(entries))
	}

	// Request less than available
	entries = buf.GetLast(1)
	if len(entries) != 1 {
		t.Errorf("GetLast(1) should return 1, got %d", len(entries))
	}
	if entries[0].Message != "msg2" {
		t.Errorf("GetLast(1) should return last entry, got '%s'", entries[0].Message)
	}
}

func TestBuffer_GetSince(t *testing.T) {
	buf := New(10)

	buf.Add(LevelInfo, "test", "msg1") // ID 1
	buf.Add(LevelInfo, "test", "msg2") // ID 2
	buf.Add(LevelInfo, "test", "msg3") // ID 3

	entries := buf.GetSince(1) // Get entries after ID 1
	if len(entries) != 2 {
		t.Errorf("GetSince(1) should return 2 entries, got %d", len(entries))
	}

	entries = buf.GetSince(0) // Get all entries
	if len(entries) != 3 {
		t.Errorf("GetSince(0) should return 3 entries, got %d", len(entries))
	}

	entries = buf.GetSince(10) // Get none
	if len(entries) != 0 {
		t.Errorf("GetSince(10) should return 0 entries, got %d", len(entries))
	}
}

func TestBuffer_GetFiltered(t *testing.T) {
	buf := New(10)

	buf.Add(LevelDebug, "src1", "debug msg")
	buf.Add(LevelInfo, "src1", "info msg")
	buf.Add(LevelWarn, "src2", "warn msg")
	buf.Add(LevelError, "src2", "error msg")

	// Filter by level
	entries := buf.GetFiltered(LevelWarn, "", 0)
	if len(entries) != 2 {
		t.Errorf("GetFiltered(warn) should return 2 entries, got %d", len(entries))
	}

	entries = buf.GetFiltered(LevelError, "", 0)
	if len(entries) != 1 {
		t.Errorf("GetFiltered(error) should return 1 entry, got %d", len(entries))
	}

	// Filter by source
	entries = buf.GetFiltered(LevelDebug, "src1", 0)
	if len(entries) != 2 {
		t.Errorf("GetFiltered(debug, src1) should return 2 entries, got %d", len(entries))
	}

	// Filter with limit
	entries = buf.GetFiltered(LevelDebug, "", 2)
	if len(entries) != 2 {
		t.Errorf("GetFiltered with limit 2 should return 2, got %d", len(entries))
	}
}

func TestBuffer_Write(t *testing.T) {
	buf := New(10)

	// Test Write implementation
	msg := "[HTTP] Server listening on :8080\n"
	n, err := buf.Write([]byte(msg))
	if err != nil {
		t.Errorf("Write should not return error, got %v", err)
	}
	if n != len(msg) {
		t.Errorf("Write should return number of bytes written, got %d, want %d", n, len(msg))
	}

	entries := buf.GetLast(1)
	if len(entries) != 1 {
		t.Fatal("Write should add one entry")
	}
	if entries[0].Source != "HTTP" {
		t.Errorf("Source should be 'HTTP', got '%s'", entries[0].Source)
	}
}

func TestBuffer_Write_LevelDetection(t *testing.T) {
	buf := New(10)

	buf.Write([]byte("[Test] Error occurred\n"))
	buf.Write([]byte("[Test] Warning message\n"))
	buf.Write([]byte("[Test] Debug info\n"))
	buf.Write([]byte("[Test] Regular message\n"))

	entries := buf.GetLast(4)
	if entries[0].Level != LevelError {
		t.Errorf("Entry with 'Error' should be error level, got %s", entries[0].Level)
	}
	if entries[1].Level != LevelWarn {
		t.Errorf("Entry with 'Warning' should be warn level, got %s", entries[1].Level)
	}
	if entries[2].Level != LevelDebug {
		t.Errorf("Entry with 'Debug' should be debug level, got %s", entries[2].Level)
	}
	if entries[3].Level != LevelInfo {
		t.Errorf("Regular message should be info level, got %s", entries[3].Level)
	}
}

func TestBuffer_Subscribe(t *testing.T) {
	buf := New(10)

	sub := buf.Subscribe("test-sub", LevelInfo)
	if sub == nil {
		t.Fatal("Subscribe should return subscriber")
	}
	if sub.ID != "test-sub" {
		t.Errorf("Subscriber ID should be 'test-sub', got '%s'", sub.ID)
	}

	// Add entry and check if subscriber receives it
	buf.Add(LevelInfo, "test", "hello")

	select {
	case entry := <-sub.Ch:
		if entry.Message != "hello" {
			t.Errorf("Subscriber should receive 'hello', got '%s'", entry.Message)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Subscriber should receive entry")
	}

	buf.Unsubscribe("test-sub")
}

func TestBuffer_Subscribe_Filter(t *testing.T) {
	buf := New(10)

	sub := buf.Subscribe("test-sub", LevelWarn)

	// Debug and info should not be sent
	buf.Add(LevelDebug, "test", "debug")
	buf.Add(LevelInfo, "test", "info")

	select {
	case <-sub.Ch:
		t.Error("Subscriber with warn filter should not receive debug/info")
	case <-time.After(50 * time.Millisecond):
		// Expected
	}

	// Warn and error should be sent
	buf.Add(LevelWarn, "test", "warn")

	select {
	case entry := <-sub.Ch:
		if entry.Level != LevelWarn {
			t.Errorf("Subscriber should receive warn level, got %s", entry.Level)
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("Subscriber should receive warn entry")
	}

	buf.Unsubscribe("test-sub")
}

func TestBuffer_Unsubscribe(t *testing.T) {
	buf := New(10)

	sub := buf.Subscribe("test-sub", LevelInfo)
	buf.Unsubscribe("test-sub")

	// Channel should be closed
	_, ok := <-sub.Ch
	if ok {
		t.Error("Channel should be closed after unsubscribe")
	}

	// Unsubscribe again should not panic
	buf.Unsubscribe("test-sub")
}

func TestBuffer_Clear(t *testing.T) {
	buf := New(10)

	buf.Add(LevelInfo, "test", "msg1")
	buf.Add(LevelInfo, "test", "msg2")

	if buf.Size() != 2 {
		t.Errorf("Size before clear should be 2, got %d", buf.Size())
	}

	buf.Clear()

	if buf.Size() != 0 {
		t.Errorf("Size after clear should be 0, got %d", buf.Size())
	}

	entries := buf.GetLast(10)
	if len(entries) != 0 {
		t.Errorf("GetLast after clear should return empty, got %d", len(entries))
	}
}

func TestBuffer_LogHelpers(t *testing.T) {
	buf := New(10)

	buf.Info("src", "info %s", "msg")
	buf.Debug("src", "debug %s", "msg")
	buf.Warn("src", "warn %s", "msg")
	buf.Error("src", "error %s", "msg")

	entries := buf.GetLast(4)
	if len(entries) != 4 {
		t.Fatalf("Should have 4 entries, got %d", len(entries))
	}

	if entries[0].Level != LevelInfo || entries[0].Message != "info msg" {
		t.Errorf("Info helper failed: %+v", entries[0])
	}
	if entries[1].Level != LevelDebug || entries[1].Message != "debug msg" {
		t.Errorf("Debug helper failed: %+v", entries[1])
	}
	if entries[2].Level != LevelWarn || entries[2].Message != "warn msg" {
		t.Errorf("Warn helper failed: %+v", entries[2])
	}
	if entries[3].Level != LevelError || entries[3].Message != "error msg" {
		t.Errorf("Error helper failed: %+v", entries[3])
	}
}

func TestMultiWriter(t *testing.T) {
	buf := New(10)
	var stdout bytes.Buffer

	mw := NewMultiWriter(buf, &stdout)

	msg := "[Test] Hello world\n"
	n, err := mw.Write([]byte(msg))

	if err != nil {
		t.Errorf("Write should not error: %v", err)
	}
	if n != len(msg) {
		t.Errorf("Write should return %d, got %d", len(msg), n)
	}

	// Check stdout
	if stdout.String() != msg {
		t.Errorf("Stdout should contain '%s', got '%s'", msg, stdout.String())
	}

	// Check buffer
	entries := buf.GetLast(1)
	if len(entries) != 1 {
		t.Error("Buffer should have 1 entry")
	}
}

func TestContainsIgnoreCase(t *testing.T) {
	tests := []struct {
		s      string
		substr string
		want   bool
	}{
		{"Error occurred", "error", true},
		{"ERROR", "error", true},
		{"No match", "error", false},
		{"", "error", false},
		{"error", "", true},
		{"WARNING: something", "warn", true},
	}

	for _, tt := range tests {
		got := containsIgnoreCase(tt.s, tt.substr)
		if got != tt.want {
			t.Errorf("containsIgnoreCase(%q, %q) = %v, want %v", tt.s, tt.substr, got, tt.want)
		}
	}
}

func TestShouldSend(t *testing.T) {
	tests := []struct {
		entry  Level
		filter Level
		want   bool
	}{
		{LevelError, LevelDebug, true},
		{LevelWarn, LevelDebug, true},
		{LevelInfo, LevelDebug, true},
		{LevelDebug, LevelDebug, true},
		{LevelError, LevelError, true},
		{LevelDebug, LevelError, false},
		{LevelInfo, LevelWarn, false},
		{Level("unknown"), LevelInfo, true}, // Unknown levels pass through
	}

	for _, tt := range tests {
		got := shouldSend(tt.entry, tt.filter)
		if got != tt.want {
			t.Errorf("shouldSend(%s, %s) = %v, want %v", tt.entry, tt.filter, got, tt.want)
		}
	}
}
