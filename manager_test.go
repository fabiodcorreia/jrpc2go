package jrpc2go_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	jrpc "github.com/fabiodcorreia/jrpc2go"
)

type addMethod struct{}

type addMethodParams struct {
	V1 int64 `json:"v1"`
	V2 int64 `json:"v2"`
}

func (m *addMethod) Execute(req *jrpc.Request, resp *jrpc.Response) {
	var p addMethodParams
	if err := req.ParseParams(&p); err != nil {
		resp.Error = err
		return
	}
	r := p.V1 + p.V2
	if r == 20 {
		time.Sleep(10 * time.Second) // Simulate timeout
	}
	if r == 1 {
		//To cover the case where the response returns Error and Result
		resp.Error = &jrpc.Error{
			Code:    1,
			Message: "Fake error for test",
		}
	}
	resp.Result = r
}

func TestManagerBuilder_Add(t *testing.T) {
	type args struct {
		methodName string
		method     jrpc.Method
		wantErr    bool
	}

	tests := []struct {
		name string
		args args
	}{
		{
			name: "Valid Method",
			args: args{
				methodName: "add",
				method:     &addMethod{},
				wantErr:    false,
			},
		},
		{
			name: "Invalid Method Name",
			args: args{
				methodName: "",
				method:     &addMethod{},
				wantErr:    true,
			},
		},
		{
			name: "Invalid Method",
			args: args{
				methodName: "methodX",
				method:     nil,
				wantErr:    true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) && !tt.args.wantErr {
					t.Errorf("Manager.Add() recover = %v, wantPanic = %v", r, tt.args.wantErr)
				}
			}()
			_ = jrpc.NewManagerBuilder().Add(tt.args.methodName, tt.args.method).Build()
		})
	}
}

func TestManager(t *testing.T) {
	m := jrpc.NewManagerBuilder().
		SetTimeout(1*time.Second).
		Add("add", &addMethod{}).
		Add("sum", &addMethod{}).
		Build()

	type args struct {
		ctx context.Context
		r   io.Reader
		w   io.Writer
	}
	tests := []struct {
		name    string
		args    args
		wantW   string
		wantErr bool
	}{
		{
			name: "No Reader",
			args: args{
				r: nil,
				w: &bytes.Buffer{},
			},
			wantErr: true,
		},
		{
			name: "No Writer",
			args: args{
				r: bytes.NewReader([]byte(`{"jsonrpc": "2.0","method": "add","id": "1","params": {"v1": 10,"v2": 120}}`)),
				w: nil,
			},
			wantErr: true,
		},
		{
			name: "Valid Request",
			args: args{
				r:   bytes.NewReader([]byte(`{"jsonrpc": "2.0","method": "add","id": "1","params": {"v1": 10,"v2": 120}}`)),
				ctx: context.Background(),
				w:   &bytes.Buffer{},
			},
			wantW: `{"jsonrpc":"2.0","id":"1","result":130}`,
		},
		{
			name: "Valid Request no Context",
			args: args{
				r: bytes.NewReader([]byte(`{"jsonrpc": "2.0","method": "add","id": "1","params": {"v1": 10,"v2": 120}}`)),
				w: &bytes.Buffer{},
			},
			wantW: `{"jsonrpc":"2.0","id":"1","result":130}`,
		},
		{
			name: "Valid Batch Request",
			args: args{
				r:   bytes.NewReader([]byte(`[{"jsonrpc":"2.0","method":"add","id":"1","params":{"v1":10,"v2":120}},{"jsonrpc":"2.0","method":"sum","id":"2","params":{"v1":10,"v2":20}}]`)),
				ctx: context.Background(),
				w:   &bytes.Buffer{},
			},
			wantW: `[{"jsonrpc":"2.0","id":"1","result":130},{"jsonrpc":"2.0","id":"2","result":30}]`,
		},
		{
			name: "Empty Request",
			args: args{
				r:   bytes.NewReader([]byte(``)),
				ctx: context.Background(),
				w:   &bytes.Buffer{},
			},
			wantErr: true,
		},
		{
			name: "Empty Request Object",
			args: args{
				r:   bytes.NewReader([]byte(`{}`)),
				ctx: context.Background(),
				w:   &bytes.Buffer{},
			},
			wantW: `{"jsonrpc":"","id":null,"error":{"code":-32001,"message":"JSON RPC Version must be 2.0","data":""}}`,
		},
		{
			name: "Empty Batch Request Object",
			args: args{
				r:   bytes.NewReader([]byte(`[]`)),
				ctx: context.Background(),
				w:   &bytes.Buffer{},
			},
			wantErr: true,
		},
		{
			name: "Method Not Found",
			args: args{
				r:   bytes.NewReader([]byte(`{"jsonrpc": "2.0","method": "no-method","id": "1","params": {"v1": 10,"v2": 1}}`)),
				ctx: context.Background(),
				w:   &bytes.Buffer{},
			},
			wantW: `{"jsonrpc":"2.0","id":"1","error":{"code":-32601,"message":"Method not found","data":"no-method"}}`,
		},
		{
			name: "No Method Found",
			args: args{
				r:   bytes.NewReader([]byte(`{"jsonrpc": "2.0","id": "1","params": {"v1": 10,"v2": 1}}`)),
				ctx: context.Background(),
				w:   &bytes.Buffer{},
			},
			wantW: `{"jsonrpc":"2.0","id":"1","error":{"code":-32601,"message":"Method not found","data":"Method not specified or empty"}}`,
		},
		{
			name: "Request Timeout",
			args: args{
				r:   bytes.NewReader([]byte(`{"jsonrpc": "2.0","method": "sum","id": "1","params": {"v1": 10,"v2": 10}}`)),
				ctx: context.Background(),
				w:   &bytes.Buffer{},
			},
			wantW: `{"jsonrpc":"2.0","id":"1","error":{"code":-32002,"message":"Method execution timeout"}}`,
		},
		{
			name: "Response Error with Result",
			args: args{
				r:   bytes.NewReader([]byte(`{"jsonrpc": "2.0","method": "sum","id": "1","params": {"v1": 0,"v2": 1}}`)),
				ctx: context.Background(),
				w:   &bytes.Buffer{},
			},
			wantW: `{"jsonrpc":"2.0","id":"1","error":{"code":1,"message":"Fake error for test"}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := m.Handle(tt.args.ctx, tt.args.r, tt.args.w)

			if err != nil && !tt.wantErr {
				t.Errorf("Manager.Handle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				//TODO check the error
				return //Skip for now
			}

			wt := tt.args.w.(*bytes.Buffer)
			if gotW := wt.String(); strings.Compare(strings.TrimSpace(gotW), tt.wantW) != 0 {
				t.Errorf("Manager.Handle() result = '%v', want %v", gotW, tt.wantW)
			}
		})
	}
}
