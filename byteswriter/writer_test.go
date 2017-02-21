package byteswriter

import (
	"bytes"
	"fmt"
	"io"
	"testing"
)

func TestBytesWriter(t *testing.T) {
	testdata1 := []byte{1, 2, 3, 4, 5, 6, 7, 8}
	testdata2 := []byte{11, 12, 13, 14}
	testdata3 := []byte{21, 22, 23, 24}

	w := NewPreallocated(4)
	n, err := w.Write(testdata1)
	if err != nil {
		t.Error(err)
	} else if n != len(testdata1) {
		t.Errorf("mismatched sizes")
	} else if !bytes.Equal(w.Bytes(), testdata1) {
		t.Errorf("mismatched data")
	}
	i, err := w.Seek(-6, io.SeekCurrent)
	if err != nil {
		t.Error(err)
	} else if i != 2 {
		t.Errorf("invalid seek return value")
	}
	n, err = w.Write(testdata2)
	if err != nil {
		t.Error(err)
	} else if n != len(testdata2) {
		t.Errorf("mismatched sizes")
	} else if !bytes.Equal(w.Bytes(), []byte{1, 2, 11, 12, 13, 14, 7, 8}) {
		t.Errorf("mismatched data")
	}
	i, err = w.Seek(2, io.SeekEnd)
	if err != nil {
		t.Error(err)
	} else if i != 6 {
		t.Errorf("Invalid seek return value")
	}
	n, err = w.Write(testdata3)
	if err != nil {
		t.Error(err)
	} else if n != len(testdata3) {
		t.Errorf("mismatched sizes")
	} else if !bytes.Equal(w.Bytes(), []byte{1, 2, 11, 12, 13, 14, 21, 22, 23, 24}) {
		t.Errorf("mismatched data")
	}
	size := w.Size()
	if size != 10 {
		t.Errorf("mismatched size")
	}
	i, err = w.Seek(int64(size)-3, io.SeekStart)
	if err != nil {
		t.Error(err)
	} else if i != 7 {
		t.Errorf("Invalid seek return value")
	}
	n, err = w.Write(testdata3)
	if err != nil {
		t.Error(err)
	} else if n != len(testdata3) {
		t.Errorf("mismatched sizes")
	} else if !bytes.Equal(w.Bytes(), []byte{1, 2, 11, 12, 13, 14, 21, 21, 22, 23, 24}) {
		fmt.Println(w.Bytes())
		t.Errorf("mismatched data")
	}

	_, err = w.Seek(0, 255)
	if err == nil {
		t.Errorf("whence not checked properly")
	}
	_, err = w.Seek(-1, io.SeekStart)
	if err == nil {
		t.Errorf("offset start not checked properly")
	}
	_, err = w.Seek(-1, io.SeekEnd)
	if err != nil {
		t.Errorf("offset end not checked properly")
	}
	n, err = w.Write(testdata1)
	if err == nil {
		t.Errorf("writing while seeked past end of buffer should not work.")
	}
}
