package mcpstdio

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Fabien-Halaby/datara/internal/core/domain"
)

func (s *Server) handleInitialize(req Request) Response {
	result := map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]any{
			"tools": map[string]any{},
		},
		"serverInfo": map[string]any{
			"name":    "datara",
			"version": "0.1.0",
		},
	}
	return newResult(req.ID, result)
}

func (s *Server) handleToolsList(req Request) Response {
	tools := []map[string]any{
		{
			"name":        "query_database",
			"description": "Run a read-only SQL SELECT query against the connected Postgres database. Any statement that is not a single SELECT is rejected.",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"sql": map[string]any{
						"type":        "string",
						"description": "A single SQL SELECT statement.",
					},
				},
				"required": []string{"sql"},
			},
		},
	}
	return newResult(req.ID, map[string]any{"tools": tools})
}

type toolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type queryArguments struct {
	SQL string `json:"sql"`
}

func (s *Server) handleToolsCall(ctx context.Context, req Request) Response {
	var params toolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return newError(req.ID, -32602, "invalid params")
	}

	if params.Name != "query_database" {
		return newError(req.ID, -32602, fmt.Sprintf("unknown tool: %s", params.Name))
	}

	var args queryArguments
	if err := json.Unmarshal(params.Arguments, &args); err != nil {
		return newError(req.ID, -32602, `invalid arguments: expected {"sql": "..."}`)
	}

	query, err := s.validator.Validate(args.SQL)
	if err != nil {
		s.auditor.LogBlocked(domain.QueryBlockedEvent{
			Query:     args.SQL,
			Reason:    err.Error(),
			Timestamp: time.Now(),
		})
		return newResult(req.ID, toolErrorResult(fmt.Sprintf("Query rejected: %s", err.Error())))
	}

	start := time.Now()
	result, err := s.datasource.Execute(ctx, query)
	if err != nil {
		return newResult(req.ID, toolErrorResult(fmt.Sprintf("Execution failed: %s", err.Error())))
	}

	s.auditor.LogExecuted(domain.QueryExecutedEvent{
		Query:     query.Raw(),
		RowCount:  result.RowCount,
		Duration:  time.Since(start),
		Timestamp: time.Now(),
	})

	return newResult(req.ID, toolSuccessResult(result))
}

func toolErrorResult(message string) map[string]any {
	return map[string]any{
		"isError": true,
		"content": []map[string]any{
			{"type": "text", "text": message},
		},
	}
}

func toolSuccessResult(result domain.QueryResult) map[string]any {
	return map[string]any{
		"isError": false,
		"content": []map[string]any{
			{"type": "text", "text": formatResultAsText(result)},
		},
	}
}

func formatResultAsText(result domain.QueryResult) string {
	if result.RowCount == 0 {
		return "Query executed successfully. No rows returned."
	}

	out := fmt.Sprintf("%d row(s)", result.RowCount)
	if result.Truncated {
		out += " (truncated by max-rows policy)"
	}
	out += "\n\n"

	// Simple pipe-separated table: good enough for an MVP and easy for
	// Claude to parse and present to the user.
	out += joinStrings(result.Columns, " | ") + "\n"
	for _, row := range result.Rows {
		cells := make([]string, len(row))
		for i, v := range row {
			cells[i] = fmt.Sprintf("%v", v)
		}
		out += joinStrings(cells, " | ") + "\n"
	}
	return out
}

func joinStrings(items []string, sep string) string {
	out := ""
	for i, item := range items {
		if i > 0 {
			out += sep
		}
		out += item
	}
	return out
}