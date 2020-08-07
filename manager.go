package jrpc2go

import (
	"context"
	"encoding/json"
	"io"
	"sync"
	"time"
)

// ManagerBuilder will support the Builder pattern for the Manager struct.
type ManagerBuilder struct {
	timeout time.Duration
	methods map[string]Method
}

// NewManagerBuilder will return a new builder for the Manager.
func NewManagerBuilder() *ManagerBuilder {
	return &ManagerBuilder{
		timeout: 10 * time.Second,
		methods: make(map[string]Method),
	}
}

// SetTimeout allows to specify a custom timeout for each method execution.
//
// Default timeout is 10 seconds
func (mb *ManagerBuilder) SetTimeout(timeout time.Duration) *ManagerBuilder {
	mb.timeout = timeout
	return mb
}

// Add will append a new method to the manager to be executed. the name should be unique
// if the name name is used more then one time it will overwrite the handler of that method.
//
// If the name is empty or the h is nil this function will panic.
func (mb *ManagerBuilder) Add(name string, h Method) *ManagerBuilder {
	if name == "" || h == nil {
		// The program should not even start in this cases otherwise it will crash later trying to execute
		// h wich is nil.
		panic("jsonrpc: method name and function should not be empty")
	}
	mb.methods[name] = h
	return mb
}

// Build will use the configuration collected during the build return a manager
// with these configurations.
func (mb *ManagerBuilder) Build() Manager {
	return Manager{
		methods: mb.methods,
		timeout: mb.timeout,
	}
}

// Manager represent the JSON RPC method register manager.
type Manager struct {
	mu      sync.RWMutex
	methods map[string]Method
	timeout time.Duration
}

// Handle will receive a request content and write the result of the excecution to the writer.
//
// It can return an error if the JSON encoding or the writing fails.
func (m *Manager) Handle(ctx context.Context, r io.Reader, w io.Writer) error {
	if r == nil {
		return newError(errCodeInternal, "r io.Reader can't be nil")
	}

	if w == nil {
		return newError(errCodeInternal, "w io.Writer can't be nil")
	}

	rq, err := parseMethodRequest(r)
	if err != nil {
		return err
	}

	if ctx == nil {
		ctx = context.Background()
	}

	if len(rq) == 0 {
		return newError(errCodeInvalidRequest, "no methods specified")
	}

	resp := make([]*Response, 0, len(rq))

	for i := range rq {
		tResp := m.execMethod(ctx, rq[i])
		// If no ID means it's a notification and the server shouldn't reply
		// if we have an error it should return anyway
		if rq[i].ID != nil || tResp.Error != nil {
			resp = append(resp, tResp)
		}
	}

	// If more then one response return a json array
	if len(resp) > 1 {
		return json.NewEncoder(w).Encode(resp)
		//if err := json.NewEncoder(w).Encode(resp); err != nil {
		//	return err
		//}
	}
	// If only one response return a json object
	if len(resp) == 1 {
		return json.NewEncoder(w).Encode(resp[0])
	}
	// If no response don't send anything
	return nil
}

// execMethod will receive a request, execute the method and return the response.
func (m *Manager) execMethod(ctx context.Context, req *Request) *Response {
	res := newResponse(req)
	if req.Version != version {
		res.Error = newError(errCodeInvalidRPCVersion, res.Version)
		return res
	}

	if req.Method == "" {
		res.Error = newError(errCodeMethodNotFound, "Method not specified or empty")
		return res
	}

	m.mu.RLock()
	method, ok := m.methods[req.Method]
	m.mu.RUnlock()

	if !ok {
		res.Error = newError(errCodeMethodNotFound, req.Method)
		return res
	}

	finish := make(chan bool, 1)

	ctxT, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()
	req = req.WithContext(ctxT)

	//! The goroutine will stay there until it finish even after the timeout
	go func() {
		method.Execute(req, res)
		close(finish)
	}()

	select {
	case <-ctxT.Done():
		res.Error = newError(errCodeExecutionTimeout, nil)
	case <-finish:
		if res.Error != nil {
			res.Result = nil
		}
	}
	return res
}
