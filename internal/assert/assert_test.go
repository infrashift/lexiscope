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

func runWithRecover(fail bool) (code int, message string) {
	defer Recover(&code, &message)
	True(!fail, "x is set")
	return 0, ""
}

func TestRecoverSetsTheExitCode(t *testing.T) {
	code, msg := runWithRecover(true)
	if code != ExitAssertion || msg != "internal error: assertion failed: x is set" {
		t.Fatalf("runWithRecover(true) = %d, %q", code, msg)
	}
	if code, msg := runWithRecover(false); code != 0 || msg != "" {
		t.Fatalf("runWithRecover(false) = %d, %q; want 0 and no message", code, msg)
	}
}

func TestRecoverPanicsAgainOnOtherValues(t *testing.T) {
	defer func() {
		if r := recover(); r != "boom" {
			t.Fatalf("recover() = %#v, want boom", r)
		}
	}()
	var code int
	var msg string
	func() {
		defer Recover(&code, &msg)
		panic("boom")
	}()
	t.Fatal("Recover swallowed a panic that is not a Failure")
}
