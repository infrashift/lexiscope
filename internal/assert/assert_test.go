package assert

import (
	"bytes"
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

func runWithRecover(w *bytes.Buffer, fail bool) (code int) {
	defer Recover(w, "cmd", &code)
	True(!fail, "x is set")
	return 0
}

func TestRecoverSetsTheExitCode(t *testing.T) {
	var buf bytes.Buffer
	if code := runWithRecover(&buf, true); code != ExitAssertion {
		t.Fatalf("code = %d, want %d", code, ExitAssertion)
	}
	if got, want := buf.String(), "cmd: internal error: assertion failed: x is set\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
	buf.Reset()
	if code := runWithRecover(&buf, false); code != 0 || buf.Len() != 0 {
		t.Fatalf("code = %d, output = %q; want 0 and no output", code, buf.String())
	}
}

func TestRecoverPanicsAgainOnOtherValues(t *testing.T) {
	defer func() {
		if r := recover(); r != "boom" {
			t.Fatalf("recover() = %#v, want boom", r)
		}
	}()
	var code int
	func() {
		defer Recover(&bytes.Buffer{}, "cmd", &code)
		panic("boom")
	}()
	t.Fatal("Recover swallowed a panic that is not a Failure")
}
