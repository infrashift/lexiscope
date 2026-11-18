// Package output writes the text of a command and keeps the first write
// error. A command checks Err once at its end and exits with status 2 when
// the output failed. This checks the result of each write (rule R7) without
// a test after each line.
package output

import (
	"fmt"
	"io"

	"github.com/infrashift/lexis/internal/assert"
)

// Writer writes formatted text to an io.Writer.
type Writer struct {
	w   io.Writer
	err error
}

// New returns a Writer for w.
func New(w io.Writer) *Writer {
	assert.True(w != nil, "the destination exists")
	o := &Writer{w: w}
	assert.True(o.err == nil, "a new writer has no error")
	return o
}

// Printf writes formatted text. After the first error it writes nothing.
func (o *Writer) Printf(format string, args ...any) {
	assert.True(o != nil && o.w != nil, "the writer exists")
	assert.True(format != "", "the format is not empty")
	if o.err != nil {
		return
	}
	n, err := fmt.Fprintf(o.w, format, args...)
	assert.True(n >= 0, "the byte count is not negative")
	o.err = err
}

// Err returns the first write error.
func (o *Writer) Err() error {
	assert.True(o != nil, "the writer exists")
	assert.True(o.w != nil, "the writer has a destination")
	return o.err
}
