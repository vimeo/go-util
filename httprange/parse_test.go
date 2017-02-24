package httprange

import (
	"fmt"
	"testing"
)

type pout struct {
	Name, In string
	Out      []Bytes
}

// Add an Equals for testing only
func (want Bytes) Equals(t *testing.T, got Bytes) {
	if want.Start != got.Start {
		t.Errorf("want: %v, got: %v", want.Start, got.Start)
	}
	if want.End != got.End {
		t.Errorf("want: %v, got: %v", want.End, got.End)
	}
	if want.Length != got.Length {
		t.Errorf("want: %v, got: %v", want.Length, got.Length)
	}
	if want.Satisfied != got.Satisfied {
		t.Errorf("want: %v, got: %v", want.Satisfied, got.Satisfied)
	}
}

func TestParseResponseWellFormed(t *testing.T) {
	tbl := []pout{
		{"Full", "bytes 0-1200/2400", []Bytes{{0, 1200, 2400, true}}},
		{"NoLength", "bytes 0-1200/*", []Bytes{{0, 1200, -1, true}}},
		{"Unsatisfiable", "bytes */2400", []Bytes{{-1, -1, 2400, false}}},
	}
	for _, c := range tbl {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			t.Logf("input: %q", c.In)
			got, err := ParseResponse(c.In)
			if err != nil {
				t.Fatal(err)
			}
			c.Out[0].Equals(t, got)
		})
	}
}

func TestParseRequestWellFormed(t *testing.T) {
	tbl := []pout{
		{"Single", "bytes=0-1200", []Bytes{{0, 1200, 0, false}}},
		{"Open", "bytes=0-", []Bytes{{0, -1, 0, false}}},
		{"Backwards", "bytes=-1200", []Bytes{{-1200, -1, 0, false}}},
		{"Multiple", "bytes=0-1200,4096-5296", []Bytes{
			{0, 1200, 0, false},
			{4096, 5296, 0, false}}},
		{"MultipleOpen", "bytes=0-1200,4096-", []Bytes{
			{0, 1200, 0, false},
			{4096, -1, 0, false}}},
		{"MultipleOpen", "bytes=4096-,0-1200", []Bytes{
			{4096, -1, 0, false},
			{0, 1200, 0, false}}},
		{"MultipleBackwards", "bytes=0-1200,-1200", []Bytes{
			{0, 1200, 0, false},
			{-1200, -1, 0, false}}},
		{"MultipleBackwards", "bytes=-1200,0-1200", []Bytes{
			{-1200, -1, 0, false},
			{0, 1200, 0, false}}},
	}
	for _, c := range tbl {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			t.Logf("input: %q", c.In)
			out, err := ParseRequest(c.In)
			if err != nil {
				t.Fatal(err)
			}
			for i := range out {
				c.Out[i].Equals(t, out[i])
			}
		})
	}

}

type perr struct {
	Name, In string
	Out      error
}

func TestParseResponseMalformed(t *testing.T) {
	tbl := []perr{
		{"NotBytes", "notbytes 0-1200/2400", ErrNotByteUnit},
		{"BadStart", "bytes f-10/*", fmt.Errorf(`wanted an int-like (got "f")`)},
		{"BadEnd", "bytes 0-af/*", fmt.Errorf(`wanted an int-like (got "a")`)},
		{"BadRange", "bytes 0/*", fmt.Errorf(`wanted an int-like until '-' (got "0/")`)},
		{"BadLen", "bytes */0xff", fmt.Errorf(`wanted int-like until end of string (got "0x")`)},
		{"NoEnd", "bytes 0-/42", fmt.Errorf(`wanted an int-like (got "")`)},
		{"NoLenSep", "bytes 0-42", fmt.Errorf(`wanted an int-like until '/', hit eof (got "42")`)},
		{"NoLen", "bytes 0-42/", fmt.Errorf(`wanted int-like until eof (got "")`)},
		{"Nonsense", "bytes */*", fmt.Errorf(`nonsense header: "bytes */*"`)},
	}
	for _, c := range tbl {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			t.Logf("input: %q", c.In)
			_, got := ParseResponse(c.In)
			if got == nil {
				t.Fatal("want: error, got: nil")
			}
			if want := c.Out; got != want && got.Error() != want.Error() {
				t.Fatalf("want: %v, got: %v", want, got)
			}
		})
	}
}

func TestParseRequestMalformed(t *testing.T) {
	tbl := []perr{
		{"NotBytes", "notbytes=0-1200", ErrNotByteUnit},
		{"BadStart", "bytes=f", fmt.Errorf(`wanted int-like until '-' (got "f")`)},
		{"BadStart", "bytes=0xff", fmt.Errorf(`wanted int-like until '-' (got "0x")`)},
		{"BadEnd", "bytes=0-af", fmt.Errorf(`wanted int-like until ',' or eof (got "a")`)},
		{"NoRange", "bytes=0", fmt.Errorf(`wanted int-like until '-' (got "0")`)},
		{"BadRange", "bytes=12-0", fmt.Errorf(`invalid range: 12-0`)},
		{"NoEnd", "bytes=0,", fmt.Errorf(`wanted int-like until '-' (got "0,")`)},
	}
	for _, c := range tbl {
		c := c
		t.Run(c.Name, func(t *testing.T) {
			t.Parallel()
			t.Logf("input: %q", c.In)
			_, got := ParseRequest(c.In)
			if got == nil {
				t.Fatal("want: error, got: nil")
			}
			if want := c.Out; got != want && got.Error() != want.Error() {
				t.Fatalf("want: %v, got: %v", want, got)
			}
		})
	}
}
