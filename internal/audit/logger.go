package audit

import (
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/Fabien-Halaby/datara/internal/core/domain"
)

// JSONLineLogger writes one JSON object per line for every blocked or
// executed query. It defaults to stderr because, in stdio transport mode,
// stdout is reserved exclusively for MCP protocol messages — anything
// else written there would corrupt the JSON-RPC stream.
type JSONLineLogger struct {
	out io.Writer
}

// NewStderrLogger returns a JSONLineLogger writing to os.Stderr.
func NewStderrLogger() *JSONLineLogger {
	return &JSONLineLogger{out: os.Stderr}
}

type logLine struct {
	Type       string    `json:"type"`
	Query      string    `json:"query"`
	Reason     string    `json:"reason,omitempty"`
	RowCount   int       `json:"row_count,omitempty"`
	DurationMs int64     `json:"duration_ms,omitempty"`
	Timestamp  time.Time `json:"timestamp"`
}

func (l *JSONLineLogger) LogBlocked(event domain.QueryBlockedEvent) {
	l.write(logLine{
		Type:      "blocked",
		Query:     event.Query,
		Reason:    event.Reason,
		Timestamp: event.Timestamp,
	})
}

func (l *JSONLineLogger) LogExecuted(event domain.QueryExecutedEvent) {
	l.write(logLine{
		Type:       "executed",
		Query:      event.Query,
		RowCount:   event.RowCount,
		DurationMs: event.Duration.Milliseconds(),
		Timestamp:  event.Timestamp,
	})
}

func (l *JSONLineLogger) write(line logLine) {
	b, err := json.Marshal(line)
	if err != nil {
		return
	}
	b = append(b, '\n')
	_, _ = l.out.Write(b)
}