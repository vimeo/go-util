// Package byteswriter implements a WriteSeeker backed by a
// dynamically expanding buffer. Concurrent writes and seeks
// the same Writer are not safe, and the user is responsible
// for ensuring this.
package byteswriter

import (
	"fmt"
	"io"
)

// A Writer implements a WriteSeeker interface backed by a
// dynamically expanding buffer.
type Writer struct {
	buf []byte
	pos int
}

// New returns a new writer with an initial allocation of 4KB.
func New() *Writer {
	ret := new(Writer)
	ret.buf = make([]byte, 0, 4*1024)
	return ret
}

// NewPreallocated returns a new writer with an initial allocation of n.
func NewPreallocated(n int) *Writer {
	ret := new(Writer)
	ret.buf = make([]byte, 0, n)
	return ret
}

// Size returns the current size of the buffer.
func (w *Writer) Size() int64 {
	return int64(len(w.buf))
}

// Seek seeks to a given offset in a buffer.
func (w *Writer) Seek(offset int64, whence int) (int64, error) {
	var off int64

	switch whence {
	case io.SeekCurrent:
		off = offset + int64(w.pos)
	case io.SeekStart:
		off = offset
	case io.SeekEnd:
		off = int64(len(w.buf)) - offset
	default:
		return 0, fmt.Errorf("invalid whence")
	}

	if off < 0 {
		return 0, fmt.Errorf("cannot seek before start of buffer")
	}

	w.pos = int(off)

	return off, nil
}

// Write writes to the underlying buffer and increases size as necessary.
func (w *Writer) Write(buf []byte) (int, error) {
	if w.pos > len(w.buf) {
		return 0, fmt.Errorf("Cannot write while past end of buffer.")
	} else if w.pos == len(w.buf) {
		w.buf = append(w.buf, buf...)
	} else if w.pos+len(buf) <= len(w.buf) {
		copy(w.buf[w.pos:], buf)
	} else if w.pos+len(buf) > len(w.buf) {
		overlap := copy(w.buf[w.pos:], buf)
		w.buf = append(w.buf, buf[overlap:]...)
	}
	w.pos += len(buf)
	return len(buf), nil
}

// Bytes returns the underlying byte buffer. The slice is valid for use only
// until the next write. The slice aliases the buffer content, so changes to
// the slice will affect the content of the writer itself.
func (w *Writer) Bytes() []byte {
	return w.buf
}
