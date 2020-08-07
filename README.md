# JRPC2GO: Zero dependencies JSON RPC 2.0 Library for Golang

JRPC2GO is a minimal API to handle JSON RPC 2.0 requests that works with transport layer that implements io.Reader and io.Writer.

## Concepts

JRPC2Go works around two main concepts, the `Manager` and the `Method`.

### Manager

The Manager is responsible for handling the requests and reply with responses, it also keeps all the methods supported
and calls the right method for each request.

Once a request is received the Manager will validate the request, invoke the proper method and send back the response.

To create a new Manager we use the `NewManagerBuilder` that allows to create a new manager with the specified configuration ready to start handling requests.

```go
manager := jrpc.NewManagerBuilder().
	SetTimeout(2*time.Second).
	Add("add", &addMethod{}).
	Add("sum", &addMethod{}).
	Build()
```

With the Manager ready we just need to call the `Handle` method from the manager and provide the input source (io.Reader), the output source (io.Writer) and the current context.

```go
ctx := context.Background()
for {
	if err := manager.Handle(ctx, os.Stdin, os.Stdout); err != nil {
		if _, err := stdout.WriteString(err.Error()); err != nil {
			log.Fatal(err)
		}
	}
}
```

The example above connects the Manager to Stdin and Stdout, so we can send JSON-RPC requests and get the response from the terminal.

### Method

The Method it's an interface with only one method called `Execute`

```go
Execute(req *jrpc.Request, resp *jrpc.Response)
```

Inside this method we will put the operation "business-logic", from the Request we can get all the data sent from the client and also the request context and we put the result or error on the response.

```go
type addMethod struct {
	// database and other resourses needed for this method
}

type addMethodParams struct {
	V1 int64 `json:"value1"`
	V2 int64 `json:"value2"`
}

func (m *addMethod) Execute(req *jrpc.Request, resp *jrpc.Response) {
	var p addMethodParams
	if err := req.ParseParams(&p); err != nil {
		resp.Error = err
		return
    }
    
	resp.Result = p.V1 + p.V2
}
```

The `addMethod` struct can contain the resources needed by the `Execute` and the `addMethodParams` it's the struct that represent the parameters sent on the request.

After that we jus need to add the method to the manager and it will take care of the rest.

```go
manager := jrpc.NewManagerBuilder().
	Add("add", &addMethod{}).
	Build()
```

## Installing

```
go get -u github.com/fabiodcorreia/jrpc2go
```

Next, include the package in your application:

```go
import "github.com/fabiodcorreia/jrpc2go"
```

To make the package name less verbose it's recommended to use an alias:

```go
import jrpc "github.com/fabiodcorreia/jrpc2go"
```

## Examples

Examples full examples with HTTP and Stdin/out implementations can be found at [_examples](_examples).

## Contributing
1. Fork it
2. Clone your fork to your local machine (git clone https://github.com/your_username/jrpc2go && cd jsonrpc2go)
3. Create your feature branch (git checkout -b my-new-feature)
4. Make changes and add them (git add .)
5. Commit your changes (git commit -m 'Add some feature')
6. Push to the branch (git push origin my-new-feature)
7. Create new pull request

## License
JRPC2GO is released under the Apache 2.0 license. See [LICENSE.txt](https://github.com/fabiodcorreia/jrpc2go/blob/master/LICENSE.txt)
