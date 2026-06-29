package services

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"SwiftRay/app/models"
)

type LogStream struct {
	mu      sync.Mutex
	entries []models.LogEntry
	maxSize int
	onLog   func(models.LogEntry)
}

func NewLogStream(maxSize int) *LogStream {
	return &LogStream{
		entries: make([]models.LogEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

func (ls *LogStream) OnLog(fn func(models.LogEntry)) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	ls.onLog = fn
}

func (ls *LogStream) Add(level models.LogLevel, message string) {
	entry := models.LogEntry{
		ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
		Level:     level,
		Message:   message,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	ls.mu.Lock()
	ls.entries = append([]models.LogEntry{entry}, ls.entries...)
	if len(ls.entries) > ls.maxSize {
		ls.entries = ls.entries[:ls.maxSize]
	}
	fn := ls.onLog
	ls.mu.Unlock()

	if fn != nil {
		fn(entry)
	}
}

func (ls *LogStream) AddFromXray(level string, message string) {
	var logLevel models.LogLevel
	switch level {
	case "error":
		logLevel = models.LevelError
	case "warning":
		logLevel = models.LevelWarning
	case "debug":
		logLevel = models.LevelDebug
	default:
		logLevel = models.LevelInfo
	}
	ls.Add(logLevel, message)
}

func (ls *LogStream) GetEntries() []models.LogEntry {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	result := make([]models.LogEntry, len(ls.entries))
	copy(result, ls.entries)
	return result
}

func (ls *LogStream) Clear() {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	ls.entries = make([]models.LogEntry, 0, ls.maxSize)
}

func (ls *LogStream) MarshalJSON() ([]byte, error) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	return json.Marshal(ls.entries)
}
