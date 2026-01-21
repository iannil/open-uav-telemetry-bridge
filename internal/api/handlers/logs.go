// Package handlers provides HTTP API handlers
package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/open-uav/telemetry-bridge/internal/core/logger"
)

// LogsHandler handles log-related API endpoints
type LogsHandler struct {
	buffer *logger.Buffer
}

// NewLogsHandler creates a new logs handler
func NewLogsHandler(buffer *logger.Buffer) *LogsHandler {
	return &LogsHandler{
		buffer: buffer,
	}
}

// GetLogs returns historical log entries with optional filtering
// GET /api/v1/logs?level=info&source=Engine&limit=100&since_id=123
func (h *LogsHandler) GetLogs(w http.ResponseWriter, r *http.Request) {
	if h.buffer == nil {
		http.Error(w, "Log buffer not available", http.StatusServiceUnavailable)
		return
	}

	// Parse query parameters
	levelStr := r.URL.Query().Get("level")
	source := r.URL.Query().Get("source")
	limitStr := r.URL.Query().Get("limit")
	sinceIDStr := r.URL.Query().Get("since_id")

	// Default values
	level := logger.LevelDebug
	limit := 100

	if levelStr != "" {
		level = logger.Level(levelStr)
	}

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	var entries []logger.Entry

	if sinceIDStr != "" {
		// Get entries since a specific ID
		if sinceID, err := strconv.ParseInt(sinceIDStr, 10, 64); err == nil {
			entries = h.buffer.GetSince(sinceID)
			// Apply level and source filter
			filtered := make([]logger.Entry, 0)
			for _, e := range entries {
				if shouldSendLog(e.Level, level) && (source == "" || e.Source == source) {
					filtered = append(filtered, e)
				}
			}
			entries = filtered
		}
	} else {
		// Get filtered entries
		entries = h.buffer.GetFiltered(level, source, limit)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"logs":  entries,
		"count": len(entries),
		"total": h.buffer.Size(),
	})
}

// StreamLogs provides Server-Sent Events for real-time log streaming
// GET /api/v1/logs/stream?level=info
func (h *LogsHandler) StreamLogs(w http.ResponseWriter, r *http.Request) {
	if h.buffer == nil {
		http.Error(w, "Log buffer not available", http.StatusServiceUnavailable)
		return
	}

	// Check if SSE is supported
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Parse level filter
	levelStr := r.URL.Query().Get("level")
	level := logger.LevelInfo
	if levelStr != "" {
		level = logger.Level(levelStr)
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable nginx buffering

	// Create subscriber
	subID := uuid.New().String()
	sub := h.buffer.Subscribe(subID, level)
	defer h.buffer.Unsubscribe(subID)

	// Send initial connection event
	fmt.Fprintf(w, "event: connected\ndata: {\"subscriber_id\":\"%s\"}\n\n", subID)
	flusher.Flush()

	// Set up ping ticker to keep connection alive
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	ctx := r.Context()

	for {
		select {
		case <-ctx.Done():
			// Client disconnected
			return

		case entry, ok := <-sub.Ch:
			if !ok {
				// Channel closed
				return
			}

			// Marshal entry to JSON
			data, err := json.Marshal(entry)
			if err != nil {
				continue
			}

			// Send SSE event
			fmt.Fprintf(w, "event: log\ndata: %s\n\n", data)
			flusher.Flush()

		case <-ticker.C:
			// Send ping to keep connection alive
			fmt.Fprintf(w, "event: ping\ndata: {\"time\":%d}\n\n", time.Now().UnixMilli())
			flusher.Flush()
		}
	}
}

// ClearLogs clears all log entries
// DELETE /api/v1/logs
func (h *LogsHandler) ClearLogs(w http.ResponseWriter, r *http.Request) {
	if h.buffer == nil {
		http.Error(w, "Log buffer not available", http.StatusServiceUnavailable)
		return
	}

	h.buffer.Clear()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logs cleared",
	})
}

// shouldSendLog checks if an entry with given level should be included based on filter
func shouldSendLog(entryLevel, filterLevel logger.Level) bool {
	levels := map[logger.Level]int{
		logger.LevelDebug: 0,
		logger.LevelInfo:  1,
		logger.LevelWarn:  2,
		logger.LevelError: 3,
	}

	entryPriority, ok1 := levels[entryLevel]
	filterPriority, ok2 := levels[filterLevel]

	if !ok1 || !ok2 {
		return true
	}

	return entryPriority >= filterPriority
}
