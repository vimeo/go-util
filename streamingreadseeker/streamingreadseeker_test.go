package streamingreadseeker

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"fmt"
	"hash"
	"io"
	"testing"
)

type testDef struct {
	desc   string
	tFunc  func(*Reader, hash.Hash) error
	md5sum string
	err    error
}

const emptyMD5 = "d41d8cd98f00b204e9800998ecf8427e"

// raw data length must remain divisible by 4
const testDataBase64 = "YYHWqJWrHbj+H9ei+AbLHO79sSrbMElQQ+iwZ0j+0JPrbRJnqaxoWeFXEqg2ONub8YjExM0R1L1pDr9Oj5YdHg=="
var fSize int64
var hSize int64
var qSize int64
var testData []byte

var errReadSizeMismatch = errors.New("read size mismatch")
var errSeekPosMismatch = errors.New("seek position mismatch")

func genData() error {
	var err error

	testData, err = base64.StdEncoding.DecodeString(testDataBase64)
	if err != nil {
		return err
	}

	fSize = int64(len(testData))
	hSize = fSize / 2
	qSize = fSize / 4

	return err
}

func getNewReader() *Reader {
	b := make([]byte, fSize)
	copy(b, testData)
	return New(bytes.NewReader(b))
}

func read(r *Reader, size int64, h hash.Hash) error {
	b := make([]byte, size)
	n, err := r.Read(b)
	if err != nil {
		return err
	}
	if int64(n) != size {
		return errReadSizeMismatch
	}
	if h != nil {
		h.Write(b)
	}
	return nil
}

func seek(r *Reader, offset int64, whence int, newPos int64) error {
	pos, err := r.Seek(offset, whence)
	if err != nil {
		return err
	}
	if pos != newPos {
		return errSeekPosMismatch
	}
	return nil
}

func testOneRead(r *Reader, h hash.Hash) error {
	return read(r, fSize, h)
}

func testTwoReads(r *Reader, h hash.Hash) error {
	err := read(r, hSize, h)
	if err != nil {
		return err
	}
	return read(r, hSize, h)
}

func testCopy(r *Reader, h hash.Hash) error {
	n, err := io.Copy(h, r)
	if err != nil {
		return err
	}
	if n != fSize {
		return errReadSizeMismatch
	}

	return nil
}

func testCopyBuffer(r *Reader, h hash.Hash) error {
	buf := make([]byte, 8)
	n, err := io.CopyBuffer(h, r, buf)
	if err != nil {
		return err
	}
	if n != fSize {
		return errReadSizeMismatch
	}

	return nil
}

func testSeekFwdStart(r *Reader, h hash.Hash) error {
	var err error

	err = seek(r, hSize, io.SeekStart, hSize)
	if err != nil {
		return err
	}

	return read(r, hSize, h)
}

func testSeekFwdCurrent(r *Reader, h hash.Hash) error {
	var err error

	err = seek(r, qSize, io.SeekCurrent, qSize)
	if err != nil {
		return err
	}

	err = seek(r, qSize, io.SeekCurrent, hSize)
	if err != nil {
		return err
	}

	return read(r, hSize, h)
}

func testReadSeekRead(r *Reader, h hash.Hash) error {
	var err error

	err = read(r, qSize, h)
	if err != nil {
		return err
	}

	err = seek(r, qSize, io.SeekCurrent, hSize)
	if err != nil {
		return err
	}

	return read(r, hSize, h)
}

func testSeekBackNoRead(r *Reader, h hash.Hash) error {
	var err error

	err = seek(r, fSize, io.SeekStart, fSize)
	if err != nil {
		return err
	}

	err = seek(r, 0, io.SeekStart, 0)
	if err != nil {
		return err
	}

	return read(r, fSize, h)
}

func testReadAfterEOF(r *Reader, h hash.Hash) error {
	err := seek(r, fSize, io.SeekStart, fSize)
	if err != nil {
		return err
	}

	return read(r, 1, nil)
}

func testSeekBackStart(r *Reader, h hash.Hash) error {
	err := read(r, hSize, nil)
	if err != nil {
		return err
	}

	return seek(r, 0, io.SeekStart, 0)
}

func testSeekBackCurrent(r *Reader, h hash.Hash) error {
	err := read(r, hSize, nil)
	if err != nil {
		return err
	}

	return seek(r, -hSize, io.SeekCurrent, 0)
}

func testSeekEnd(r *Reader, h hash.Hash) error {
	return seek(r, 0, io.SeekEnd, 0)
}

func TestReader(t *testing.T) {
	tests := []testDef{
		{
			desc:   "One Read",
			tFunc:  testOneRead,
			md5sum: "94e369260530319dd1e0e87cc43caf40",
			err:    nil,
		},
		{
			desc:   "Two Reads",
			tFunc:  testTwoReads,
			md5sum: "94e369260530319dd1e0e87cc43caf40",
			err:    nil,
		},
		{
			desc:   "Copy",
			tFunc:  testCopy,
			md5sum: "94e369260530319dd1e0e87cc43caf40",
			err:    nil,
		},
		{
			desc:   "CopyBuffer",
			tFunc:  testCopyBuffer,
			md5sum: "94e369260530319dd1e0e87cc43caf40",
			err:    nil,
		},
		{
			desc:   "Seek Forward Start",
			tFunc:  testSeekFwdStart,
			md5sum: "6add1b6b87692d8701a1c744aef1dae2",
			err:    nil,
		},
		{
			desc:   "Seek Forward Current",
			tFunc:  testSeekFwdCurrent,
			md5sum: "6add1b6b87692d8701a1c744aef1dae2",
			err:    nil,
		},
		{
			desc:   "Read Seek Read",
			tFunc:  testReadSeekRead,
			md5sum: "783aa7685be6a1c7e91dffc755f1ba6b",
			err:    nil,
		},
		{
			desc:   "Seek Backwards Without Read",
			tFunc:  testSeekBackNoRead,
			md5sum: "94e369260530319dd1e0e87cc43caf40",
			err:    nil,
		},
		{
			desc:   "Read After EOF",
			tFunc:  testReadAfterEOF,
			md5sum: emptyMD5,
			err:    io.EOF,
		},
		{
			desc:   "Seek Backwards Start",
			tFunc:  testSeekBackStart,
			md5sum: emptyMD5,
			err:    ErrSeekOffset,
		},
		{
			desc:   "Seek Backwards Current",
			tFunc:  testSeekBackCurrent,
			md5sum: emptyMD5,
			err:    ErrSeekOffset,
		},
		{
			desc:   "Seek End",
			tFunc:  testSeekEnd,
			md5sum: emptyMD5,
			err:    ErrSeekWhence,
		},
	}

	err := genData()
	if err != nil {
		t.Fatalf("error generating test data: %v", err)
	}

	for _, def := range tests {
		t.Run(def.desc, fromDef(def))
	}
}

func fromDef(def testDef) func(*testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		h := md5.New()
		err := def.tFunc(getNewReader(), h)
		md5sum := fmt.Sprintf("%x", h.Sum(nil))

		if err != def.err {
			t.Fatalf("expected error \"%v\", got \"%v\"", def.err, err)
		}
		if md5sum != def.md5sum {
			t.Fatalf("expected MD5 %s, got %s", def.md5sum, md5sum)
		}
	}
}
