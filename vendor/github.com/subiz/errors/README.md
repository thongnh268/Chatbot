install tools:

	go get golang.org/x/tools/cmd/stringer

run:

	go generate



# Golang error with call stack

# [![GoDoc](https://godoc.org/github.com/subiz/errors?status.svg)](http://godoc.org/github.com/subiz/errors)

This package lets you see which line of code has created an error along with its call stack.

```json
err := readDatabase()
fmt.Println(err.(*errors.Error).Stack)
```
```
account/core/account.go:26
/vendor/git.subiz.net/header/account/account.pb.go:3306
/vendor/git.subiz.net/goutils/grpc/grpc.go:86
/vendor/git.subiz.net/goutils/grpc/grpc.go:87
/vendor/git.subiz.net/header/account/account.pb.go:3308
/vendor/google.golang.org/grpc/server.go:681
```

# Error definition
The package provides Error struct which describe an error

``` go
type Error struct {
	// Give more detail about the error
	Description string
	// HTTP code, could be 400, 500 or whatsoever
	Class       int32
	// Call stack of error (stripped)
	Stack       string
	// Creation time in nanosecond
	Created     int64
	// Should contains the unique code for an error
	Code        string
	// Describe root cause of error after being wrapped
	Root        string
	// ID of the http (rpc) request which causes the error
	RequestId   string
}
```

# Example
### Creating a brand new error
```go
err := errors.New(400, errors.E_account_id_missing, "account id not found", 1234)
```
### Wraping an existing error
`Wrap()` method converts a random error to an `*errors.Error`, information of the old error stored in Root field.
```go
err := callDatabase()
err = errors.Wrap(err, 400, errors.E_account_id_missing, "new error message")
```
### Serialize to and deserialize from string
Sometime you want to serialize the error to string so you can send it in the wire. `Error()` method do just that.
```
err.Error()
```
At the reading endpoint you can deserialize it back
```
err := errors.FromString(errstring)
```

# Caution

Must run `go generate` before run test or commit
(stringer should be install using `go get -u -a golang.org/x/tools/cmd/stringer`)

easyjson errors.go
