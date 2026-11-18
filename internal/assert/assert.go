// Package assert checks the preconditions and invariants of lexis code (rule
// R5 of the NASA Power of 10). A failed assertion is a programmer error, so
// it panics. The CLI recovers the panic at its boundary and exits with
// status 2.
package assert

import (
	"fmt"
	"io"
)

// ExitAssertion is the exit status of a command after a failed assertion.
const ExitAssertion = 2

// Failure is the panic value of a failed assertion.
type Failure struct {
	Message string
}

// Error returns the message of the failed assertion.
func (f Failure) Error() string {
	return "assertion failed: " + f.Message
}

// True panics with a Failure when cond is false.
func True(cond bool, message string) {
	if !cond {
		panic(Failure{Message: message})
	}
}

// Recover is the boundary of a command. Defer it directly:
//
//	defer assert.Recover(stderr, "lexis", &code)
//
// After a failed assertion, it writes the message to w and sets *code to
// ExitAssertion. It panics again with each other value, so a runtime error
// keeps its stack trace.
func Recover(w io.Writer, command string, code *int) {
	r := recover()
	if r == nil {
		return
	}
	f, ok := r.(Failure)
	if !ok {
		panic(r)
	}
	fmt.Fprintf(w, "%s: internal error: %v\n", command, f)
	*code = ExitAssertion
}
