// Package assert checks the preconditions and invariants of lexis code (rule
// R5 of the NASA Power of 10). A failed assertion is a programmer error, so
// it panics. The CLI recovers the panic at its boundary and exits with
// status 2.
package assert

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
