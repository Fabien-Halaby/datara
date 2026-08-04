package mcpstdio

import "encoding/json"

// Request is a JSON-RPC 2.0 request or notification (notifications omit ID).
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response is a JSON-RPC 2.0 response.
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

// RPCError is a JSON-RPC 2.0 error object.
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func newResult(id json.RawMessage, result any) Response {
	return Response{JSONRPC: "2.0", ID: id, Result: result}
}

func newError(id json.RawMessage, code int, message string) Response {
	return Response{JSONRPC: "2.0", ID: id, Error: &RPCError{Code: code, Message: message}}
}