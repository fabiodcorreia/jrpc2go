package jrpc2go_test

import (
	"encoding/json"
	"testing"

	jrpc "github.com/fabiodcorreia/jrpc2go"
)

type TestValue struct {
	V1 int64
	V2 int64
}

func TestRequest_ParseParams(t *testing.T) {
	invalidParams := []byte("{ V1:10, V2:10 }")
	validParams := []byte("{ \"V1\":10, \"V2\":10 }")
	var validValue TestValue

	type args struct {
		v interface{}
	}
	tests := []struct {
		name        string
		r           *jrpc.Request
		args        args
		wantErr     bool
		wantErrCode jrpc.ErrorCode
		want        TestValue
	}{
		{
			name: "v is nil",
			args: args{
				v: nil,
			},
			wantErr:     true,
			wantErrCode: -32602,
		},
		{
			name: "params is nil",
			args: args{
				v: &TestValue{},
			},
			r: &jrpc.Request{
				Params: nil,
			},
			wantErr:     true,
			wantErrCode: -32602,
		},
		{
			name: "params is valid",
			args: args{
				v: &validValue,
			},
			r: &jrpc.Request{
				Params: (*json.RawMessage)(&validParams),
			},
			wantErr: false,
			want:    TestValue{V1: 10, V2: 10},
		},
		{
			name: "params is invalid",
			args: args{
				v: &TestValue{},
			},
			r: &jrpc.Request{
				Params: (*json.RawMessage)(&invalidParams),
			},
			wantErr:     true,
			wantErrCode: -32602,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.r.ParseParams(tt.args.v)
			if err != nil && !tt.wantErr {
				t.Errorf("Request.ParseParams() error = %v, wantErr %v", err, tt.wantErr)
			} else if tt.wantErr && err.Code != tt.wantErrCode {
				t.Errorf("Request.ParseParams() error code = %d, wantErr %d", err.Code, tt.wantErrCode)
			} else if !tt.wantErr && tt.want != *(tt.args.v).(*TestValue) {
				t.Errorf("Request.ParseParams() value = %#v, want %#v", tt.args.v, tt.want)
			}
		})
	}
}
