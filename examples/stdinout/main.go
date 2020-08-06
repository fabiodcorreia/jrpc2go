package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	jrpc "github.com/fabiodcorreia/jrpc2go"
)

type addMethod struct {
	// database
	// other resourses needed for this method
}

type addMethodParams struct {
	V1 int64 `json:"value1"`
	V2 int64 `json:"value2"`
}

//func (m *addMethod) Execute(ctx context.Context, req *jrpc.Request, resp *jrpc.Response) {
func (m *addMethod) Execute(req *jrpc.Request, resp *jrpc.Response) {
	var p addMethodParams

	if err := req.ParseParams(&p); err != nil {
		resp.Error = err
		return
	}

	re := p.V1 + p.V2
	if re == 20 {
		time.Sleep(10 * time.Second) // Simulate timeout
		fmt.Println("Finish Long Run")
	}

	resp.Result = re
}

func main() {
	manager := jrpc.NewManagerBuilder().
		SetTimeout(2*time.Second).
		Add("add", &addMethod{}).
		Build()

	stdout := os.Stdout
	stdin := os.Stdin

	ctx := context.Background()

	for {
		if err := manager.Handle(ctx, stdin, stdout); err != nil {
			if _, err := stdout.WriteString(err.Error()); err != nil {
				log.Fatal(err)
			}
		}
	}
}
