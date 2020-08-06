package jrpc2go

import "fmt"

// Error represents a JSON-RPC error, the Response MUST contain the error member if the RPC call encounters an error.
//
// Code - A Number that indicates the error type that occurred. This MUST be an integer.
//
// Message - A String providing a short description of the error. SHOULD be limited to a concise single sentence.
//
// Data - A Primitive or Structured value that contains additional information about the error.
type Error struct {
	Code    ErrorCode   `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("jsonrpc: { code: %d, message: %s, data: %+v }", e.Code, e.Message, e.Data)
}

// ErrorCode represents the API error number.
type ErrorCode int

//ErrCodeParseError means Invalid JSON was received by the server. An error occurred on the server while parsing the JSON text.
const errCodeParseError ErrorCode = -32700

// ErrCodeInvalidRequest means the JSON sent is not a valid Request object.
const errCodeInvalidRequest ErrorCode = -32600

// ErrCodeMethodNotFound menas the method does not exist / is not available.
const errCodeMethodNotFound ErrorCode = -32601

// ErrCodeInvalidParams means a invalid method parameter(s).
const errCodeInvalidParams ErrorCode = -32602

// ErrCodeInternal means internal JSON-RPC error.
const ErrCodeInternal ErrorCode = -32603

// ErrCodeInvalidRPCVersion means the requested JSON RPC version is not correct or invalid.
const errCodeInvalidRPCVersion ErrorCode = -32001

// ErrCodeExecutionTimeout means the
const errCodeExecutionTimeout ErrorCode = -32002

// newError it's for internal use, it's used the messsages and codes from JSON RPC spec.
func newError(code ErrorCode, data interface{}) *Error {
	e := &Error{
		Code: code,
		Data: data,
	}
	switch code {
	case errCodeParseError:
		e.Message = "Parse error"
	case errCodeInvalidRequest:
		e.Message = "Invalid Request"
	case errCodeMethodNotFound:
		e.Message = "Method not found"
	case errCodeInvalidParams:
		e.Message = "Invalid method parameter(s)"
	case ErrCodeInternal:
		e.Message = "Internal error"
	case errCodeInvalidRPCVersion:
		e.Message = "JSON RPC Version must be 2.0"
	case errCodeExecutionTimeout:
		e.Message = "Method execution timeout"
	}
	return e
}

// NewError will create a new Error with a custom message
func NewError(code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}
