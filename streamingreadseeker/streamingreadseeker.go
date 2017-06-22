package streamingreadseeker

import (
	"errors"
	"io"
	"io/ioutil"
)

var ErrSeekWhence = errors.New("Seek: invalid whence")
var ErrSeekOffset = errors.New("Seek: invalid offset")

var errSeekPos = errors.New("Read: read/seek position mismatch")

// A io.ReadSeeker that works for non-backward seeks on an underlying
// io.Reader by discarding data on read if needed to progress the read position
// to the seek position.
type Reader struct {
	r    io.Reader
	rpos int64
	pos  int64
	eof  bool
}

// New returns a Reader that reads from r
func New(r io.Reader) *Reader {
	return &Reader{
		r: r,
	}
}

// Seek sets the seek position to the specified offset. If io.SeekEnd is used
// for whence, Seek will return ErrSeekWhence. If the calculated seek position
// is less than the current read position, Seek will return ErrSeekOffset.
// Calling Seek does not cause any reads on the underlying Reader. If the new
// position requires skipping data, that will be done during the next Read.
func (srs *Reader) Seek(offset int64, whence int) (int64, error) {
	var newPos int64

	switch whence {
	case io.SeekStart:
		newPos = offset
	case io.SeekCurrent:
		newPos = srs.pos + offset
	case io.SeekEnd:
		fallthrough
	default:
		return srs.pos, ErrSeekWhence
	}

	if newPos < srs.rpos {
		return srs.pos, ErrSeekOffset
	}

	srs.pos = newPos

	return newPos, nil
}

// Read reads data from the underlying Reader into p. If there was a Seek call
// that requires skipping data, that will be done prior to reading into p.
// If a previous Read call reached EOF, any subsequent Read calls will
// immediately return io.EOF.
func (srs *Reader) Read(p []byte) (int, error) {
	// check for EOF state
	if srs.eof {
		return 0, io.EOF
	}

	// catch read position up to seek position
	if srs.pos > srs.rpos {
		n, err := io.CopyN(ioutil.Discard, srs.r, srs.pos-srs.rpos)
		srs.rpos += n
		if err != nil {
			switch err {
			case io.EOF, io.ErrUnexpectedEOF:
				srs.eof = true
				return 0, io.EOF
			default:
				return 0, err
			}
		}
	}

	// this shouldn't happen
	if srs.rpos != srs.pos {
		return 0, errSeekPos
	}

	// read data from underlying Reader
	n, err := srs.r.Read(p)
	srs.rpos += int64(n)
	srs.pos = srs.rpos
	switch err {
	case io.EOF, io.ErrUnexpectedEOF:
		srs.eof = true
	}

	return n, err
}
