package mcpstdio

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/Fabien-Halaby/datara/internal/core/ports"
)

// Server implements the MCP stdio transport: it reads newline-delimited
// JSON-RPC 2.0 requests from in, dispatches them, and writes
// newline-delimited JSON-RPC 2.0 responses to out. stdout must never
// receive anything other than these responses, which is why every other
// component (audit logging, error diagnostics) writes to stderr instead.
type Server struct {
	validator  ports.SQLValidator
	datasource ports.DataSource
	auditor    ports.AuditLogger

	in  *bufio.Scanner
	out io.Writer
}

// New builds a Server reading from in and writing responses to out.
func New(in io.Reader, out io.Writer, validator ports.SQLValidator, datasource ports.DataSource, auditor ports.AuditLogger) *Server {
	scanner := bufio.NewScanner(in)
	// MCP messages can be larger than bufio's default 64KB line limit
	// (a query result can be sizeable), so raise the buffer.
	buf := make([]byte, 0, 1024*1024)
	scanner.Buffer(buf, 10*1024*1024)

	return &Server{
		validator:  validator,
		datasource: datasource,
		auditor:    auditor,
		in:         scanner,
		out:        out,
	}
}

// Run blocks, processing requests until ctx is cancelled or stdin is
// closed.
func (s *Server) Run(ctx context.Context) error {
	for s.in.Scan() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line := s.in.Bytes()
		if len(line) == 0 {
			continue
		}

		var req Request
		if err := json.Unmarshal(line, &req); err != nil {
			log.Printf("datara: failed to parse request: %v", err)
			continue
		}

		resp := s.dispatch(ctx, req)
		// Notifications (no "id") never get a response.
		if len(req.ID) == 0 {
			continue
		}
		if err := s.writeResponse(resp); err != nil {
			return fmt.Errorf("writing response: %w", err)
		}
	}
	if err := s.in.Err(); err != nil {
		return fmt.Errorf("reading stdin: %w", err)
	}
	return nil
}

func (s *Server) writeResponse(resp Response) error {
	b, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	b = append(b, '\n')
	_, err = s.out.Write(b)
	return err
}

func (s *Server) dispatch(ctx context.Context, req Request) Response {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "notifications/initialized":
		return Response{} // no-op, no response expected anyway
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(ctx, req)
	default:
		return newError(req.ID, -32601, fmt.Sprintf("method not found: %s", req.Method))
	}
}