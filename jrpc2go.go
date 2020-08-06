package jrpc2go

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// JSON RPC Specification: https://www.jsonrpc.org/specification#notification

// version represents the JSON RPC version supported.
const version = "2.0"

// jsonArrayChar is the char used on JSON to identifiy the start of an array.
const jsonArrayCharCode = 91

// Request represents a JSON-RPC call to the server and contains the following.
//
// Version - A String specifying the version of the JSON-RPC protocol. MUST be exactly "2.0".
//
// Method - A String containing the name of the method to be invoked.
//
// ID - An identifier established by the Client that MUST contain a String, Number, or NULL value if included.
// If it is not included it is assumed to be a notification. The value SHOULD normally not be Null and
// Numbers SHOULD NOT contain fractional parts.
//
// Params - A Structured value that holds the parameter values to be used during the invocation of the method.
type Request struct {
	Version string           `json:"jsonrpc"`
	Method  string           `json:"method"`
	ID      *json.RawMessage `json:"id,omitempty"`
	Params  *json.RawMessage `json:"params,omitempty"`
	ctx     context.Context
}

// ParseParams will get the params from the request and and stores the result in the value pointed to by v.
//
// Request.Params is optional but if we are calling the function they need to be there otherwise returns
// ErrInvalidParams.
func (r *Request) ParseParams(v interface{}) *Error {
	if v == nil {
		return newError(errCodeInvalidParams, "v can't be nil to parse request parameters")
	}
	if r.Params == nil {
		return newError(errCodeInvalidParams, "request doesn't have params")
	}
	if err := json.Unmarshal(*r.Params, &v); err != nil {
		return newError(errCodeInvalidParams, err)
	}
	return nil
}

// Context returns the request's context. To change the context, use WithContext.
//
// The returned context is always non-nil; it defaults to the background context.
func (r *Request) Context() context.Context {
	if r.ctx != nil {
		return r.ctx
	}
	return context.Background()
}

// WithContext returns a shallow copy of r with its context changed to ctx.
// The provided ctx must be non-nil.
func (r *Request) WithContext(ctx context.Context) *Request {
	if ctx == nil {
		panic("nil context")
	}
	r2 := new(Request)
	*r2 = *r
	r2.ctx = ctx
	return r2
}

// Response represents a JSON-RPC response from the server and containers the following.
//
// Version - A String specifying the version of the JSON-RPC protocol. MUST be exactly "2.0".
//
// Result -
//
// Error -
type Response struct {
	Version string           `json:"jsonrpc"`
	ID      *json.RawMessage `json:"id"`
	Result  interface{}      `json:"result,omitempty"`
	Error   *Error           `json:"error,omitempty"`
}

// newResponse create a Response value from a Request value.
func newResponse(r *Request) *Response {
	return &Response{
		Version: r.Version,
		ID:      r.ID,
	}
}

// A Method responds to an JSON RPC request.
//
// Execute should write reply result to the Response and then return.
// Returning signals that the request is finished
type Method interface {
	Execute(req *Request, resp *Response)
}

// parseMethodRequest will receive data from a Reader and convert into a Request value.
//
// It will return a slice of Requests (even if the reader only have one method request)
// or an Error for the following cases:
//
// - ErrCodeParseError if fail to read from the Reader or to fetch the frist char of the content.
//
// - ErrCodeInvalidRequest if the JSON RPC request is not valid.
func parseMethodRequest(r io.Reader) ([]*Request, *Error) {
	br := bufio.NewReader(r)

	if br.Size() == 0 {
		return nil, newError(errCodeParseError, "fail to read the request text: empty")
	}

	f, _, err := br.ReadRune()
	if err != nil {
		return nil, newError(errCodeParseError, fmt.Sprintf("fail to read the request text: %v", err))
	}

	if err := br.UnreadRune(); err != nil {
		return nil, newError(errCodeParseError, fmt.Sprintf("fail to read the request text: %v", err))
	}

	var rs []*Request
	if f != '[' {
		var req *Request
		if err := json.NewDecoder(br).Decode(&req); err != nil {
			return nil, newError(errCodeInvalidRequest, err)
		}
		return append(rs, req), nil
	}

	if err := json.NewDecoder(br).Decode(&rs); err != nil {
		return nil, newError(errCodeInvalidRequest, err)
	}

	return rs, nil
}
