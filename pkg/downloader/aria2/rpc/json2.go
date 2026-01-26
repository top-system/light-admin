package rpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"sync/atomic"
	"time"
)

// clientRequest represents a JSON-RPC request sent by a client
type clientRequest struct {
	Version string      `json:"jsonrpc"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params"`
	Id      uint64      `json:"id"`
}

// ClientResponse represents a JSON-RPC response returned to a client
type ClientResponse struct {
	Version string           `json:"jsonrpc"`
	Result  *json.RawMessage `json:"result"`
	Error   *json.RawMessage `json:"error"`
	Id      *uint64          `json:"id"`
}

// EncodeClientRequest encodes parameters for a JSON-RPC client request
func EncodeClientRequest(method string, args interface{}) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	c := &clientRequest{
		Version: "2.0",
		Method:  method,
		Params:  args,
		Id:      reqid(),
	}
	if err := json.NewEncoder(&buf).Encode(c); err != nil {
		return nil, err
	}
	return &buf, nil
}

func (c ClientResponse) decode(reply interface{}) error {
	if c.Error != nil {
		jsonErr := &Error{}
		if err := json.Unmarshal(*c.Error, jsonErr); err != nil {
			return &Error{
				Code:    E_SERVER,
				Message: string(*c.Error),
			}
		}
		return jsonErr
	}

	if c.Result == nil {
		return ErrNullResult
	}

	return json.Unmarshal(*c.Result, reply)
}

// DecodeClientResponse decodes the response body of a client request
func DecodeClientResponse(r io.Reader, reply interface{}) error {
	var c ClientResponse
	if err := json.NewDecoder(r).Decode(&c); err != nil {
		return err
	}
	return c.decode(reply)
}

// ErrorCode represents JSON-RPC error codes
type ErrorCode int

const (
	E_PARSE       ErrorCode = -32700
	E_INVALID_REQ ErrorCode = -32600
	E_NO_METHOD   ErrorCode = -32601
	E_BAD_PARAMS  ErrorCode = -32602
	E_INTERNAL    ErrorCode = -32603
	E_SERVER      ErrorCode = -32000
)

// ErrNullResult is returned when result is null
var ErrNullResult = errors.New("result is null")

// Error represents a JSON-RPC error
type Error struct {
	Code    ErrorCode   `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func (e *Error) Error() string {
	return e.Message
}

// reqid generates unique request IDs
var reqid = func() func() uint64 {
	var id = uint64(time.Now().UnixNano())
	return func() uint64 {
		return atomic.AddUint64(&id, 1)
	}
}()
