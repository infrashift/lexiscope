package assert

import (
	"errors"
	"testing"
)

func TestTrueDoesNotPanic(t *testing.T) {
	True(true, "unused")
}

func TestFalsePanicsWithFailure(t *testing.T) {
	defer func() {
		r := recover()
		err, ok := r.(error)
		var f Failure
		if !ok || !errors.As(err, &f) || f.Message != "count > 0" {
			t.Fatalf("recover() = %#v, want Failure{count > 0}", r)
		}
		if got := f.Error(); got != "assertion failed: count > 0" {
			t.Fatalf("Error() = %q", got)
		}
	}()
	True(false, "count > 0")
	t.Fatal("True(false) did not panic")
}
