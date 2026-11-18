package output

import (
	"bytes"
	"errors"
	"testing"
)

type failingWriter struct{ writes int }

func (f *failingWriter) Write(p []byte) (int, error) {
	f.writes++
	return 0, errors.New("disk full")
}

func TestPrintf(t *testing.T) {
	var buf bytes.Buffer
	o := New(&buf)
	o.Printf("%s=%d\n", "a", 1)
	o.Printf("b\n")
	if buf.String() != "a=1\nb\n" || o.Err() != nil {
		t.Fatalf("output = %q, err = %v", buf.String(), o.Err())
	}
}

func TestFirstErrorStops(t *testing.T) {
	f := &failingWriter{}
	o := New(f)
	o.Printf("a\n")
	o.Printf("b\n")
	if o.Err() == nil || o.Err().Error() != "disk full" {
		t.Fatalf("Err() = %v, want disk full", o.Err())
	}
	if f.writes != 1 {
		t.Fatalf("writes = %d, want 1: no write after the first error", f.writes)
	}
}
