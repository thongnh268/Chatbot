// This package lets you see which line of code has created an error along with its call stack.
//
//     err := readDatabase()
//     fmt.Println(err.(*errors.Error).Stack)
//
//
//    account/core/account.go:26
//    /vendor/github.com/subiz/header/account/account.pb.go:3306
//    /vendor/github.com/subiz/goutils/grpc/grpc.go:86
//    /vendor/github.com/subiz/goutils/grpc/grpc.go:87
//    /vendor/github.com/subiz/header/account/account.pb.go:3308
//    /vendor/google.golang.org/grpc/server.go:681
package errors

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"
)

//easyjson:json
// Error describe an error. It implements the standard golang error interface.
type Error struct {
	// Give more detail about the error
	Description string `protobuf:"bytes,2,opt,name=description" json:"description,omitempty"`
	Debug       string `protobuf:"bytes,3,opt,name=debug" json:"debug,omitempty"`
	// HTTP code, could be 400, 500 or whatsoever
	Class int32 `protobuf:"varint,6,opt,name=class" json:"class,omitempty"`
	// Call stack of error (stripped)
	Stack string `protobuf:"bytes,7,opt,name=stack" json:"stack,omitempty"`
	// Creation time in nanosecond
	Created int64 `protobuf:"varint,8,opt,name=created" json:"created,omitempty"`
	// Should contains the unique code for an error
	Code string `protobuf:"bytes,4,opt,name=code" json:"code,omitempty"`
	// Describe root cause of error after being wrapped
	Root string `protobuf:"bytes,10,opt,name=base" json:"root,omitempty"`
	// ID of the http (rpc) request which causes the error
	RequestId string `protobuf:"bytes,12,opt,name=request_id" json:"request_id,omitempty"`
}

// Wrap converts a random error to an `*errors.Error`, information of the
// old error stored in Root field.
func Wrap(err error, class int, code Code, v ...interface{}) error {
	if err == nil {
		err = &Error{}
	}
	mye, ok := err.(*Error)
	if !ok {
		e := New(class, code, append(v, err.Error()))
		e.Root = err.Error()
		return e
	}

	if mye == nil {
		return nil
	}

	if code.String() != "" && (mye.Code == "" || mye.Code == "unknown") {
		mye.Code = code.String()
	}

	if class != 0 && mye.Class == 0 {
		mye.Class = int32(class)
	}

	if len(v) > 0 {
		e := New(class, code, v)
		mye.Description += "\n" + e.Description
	}
	return mye
}

// Errorf creates default *Error with description
// This method does not include stacktrace into returned object
func Errorf(format string, v ...interface{}) *Error {
	desc := fmt.Sprintf(format, v...)
	return &Error{
		Description: desc,
		Class:       int32(500),
		Created:     time.Now().UnixNano(),
		Code:        E_unknown.String(),
	}
}

// New returns an error with the supplied message.
// New also records the stack trace at the point it was called.
func New(class int, code Code, v ...interface{}) *Error {
	var format, message string
	if len(v) == 0 {
		format = ""
	} else {
		var ok bool
		format, ok = v[0].(string)
		if !ok {
			format = strings.Repeat("%v", len(v))
		} else {
			v = v[1:]
		}
	}
	message = fmt.Sprintf(format, v...)

	e := &Error{}
	e.Description = message
	e.Class = int32(class)
	e.Stack = getStack(1)
	e.Created = time.Now().UnixNano()
	e.Code = code.String()
	return e
}

// FromString unmarshal an error string to *Error
func FromString(err string) *Error {
	if !strings.HasPrefix(err, "#ERR ") {
		return New(500, E_unknown, err)
	}
	e := &Error{}
	if er := json.Unmarshal([]byte(err[len("#ERR "):]), e); er != nil {
		return New(500, E_json_marshal_error, "%s, %s", er, err)
	}
	return e
}

// Error returns string representation of an Error
func (e *Error) Error() string {
	if e == nil {
		return ""
	}

	b, _ := json.Marshal(e)
	return "#ERR " + string(b)
}

// getStack returns 10 closest stacktrace, included file paths and line numbers
// it will ignore all system path, path which is vendor is striped to /vendor/
// skip: number of stack ignored
func getStack(skip int) string {
	stack := make([]uintptr, 10)
	var sb strings.Builder
	// skip one system stack, the this current stack line
	length := runtime.Callers(2+skip, stack[:])
	for i := 0; i < length; i++ {
		pc := stack[i]
		// pc - 1 because the program counters we use are usually return addresses,
		// and we want to show the line that corresponds to the function call
		f := runtime.FuncForPC(pc)
		file, line := f.FileLine(pc - 1)
		// dont report system path
		if isSystemPath(file) {
			continue
		}

		file = trimToPrefix(file, "/vendor/")

		// trim out common provider since most of go projects are hosted
		// in single host, there is no need to include them in the call stack
		// remove them help keeping the call stack smaller, navigatiing easier
		if !strings.HasPrefix(file, "/vendor") {
			file = trimOutPrefix(file, "/git.subiz.net/")
			file = trimOutPrefix(file, "/github.com/")
			file = trimOutPrefix(file, "/gitlab.com/")
			file = trimOutPrefix(file, "/bitbucket.org/")
			file = trimOutPrefix(file, "/gopkg.in/")
		}

		sb.WriteString(file)
		sb.WriteString(":")
		sb.WriteString(strconv.Itoa(line))
		sb.WriteString("\n")
	}
	return sb.String()
}

// isSystemPath tells whether a file is in system golang packages
func isSystemPath(path string) bool {
	if strings.Contains(path, "/github.com/subiz/errors/") {
		return true
	}
	return strings.HasPrefix(path, "/usr/local/go/src")
}

// trimToPrefix removes all the characters before the prefix
// its return the original string if not found prefix in str
func trimToPrefix(str, prefix string) string {
	i := strings.Index(str, prefix)
	if i < 0 {
		return str
	}
	return str[i:]
}

// trimOutPrefix removes all the characters before AND the prefix
// its return the original string if not found prefix in str
func trimOutPrefix(str, prefix string) string {
	i := strings.Index(str, prefix)
	if i < 0 {
		return str
	}
	return str[i+len(prefix):]
}

// CompareCode tells whether the code of the given error equals the given
// code
func CompareCode(err error, code Code) bool {
	mye, ok := err.(*Error)
	if !ok {
		return false
	}

	if mye == nil {
		return false
	}

	return mye.Code == code.String()
}
