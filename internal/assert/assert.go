// Package assert checks the preconditions and invariants of lexiscope code (rule
// R5 of the NASA Power of 10). A failed assertion is a programmer error, so
// it panics. The CLI recovers the panic at its boundary and exits with
// status 2.
package assert

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

// Recover is the boundary of a command. Defer it directly in a function
// with named results, so that it can call recover:
//
//	defer assert.Recover(&code, &message)
//
// After a failed assertion, it sets *code to ExitAssertion and *message to
// "internal error: assertion failed: ...". The caller writes the message.
// Recover panics again with each other value, so a runtime error keeps its
// stack trace.
func Recover(code *int, message *string) {
	r := recover()
	if r == nil {
		return
	}
	f, ok := r.(Failure)
	if !ok {
		panic(r)
	}
	*code = ExitAssertion
	*message = "internal error: " + f.Error()
}
