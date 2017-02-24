package httprange

import (
	"testing"
)

type brfmt struct {
	Name string
	In   []Bytes
	Out  string
}

func TestFormatBytesResponse(t *testing.T) {
	tbl := []brfmt{
		{"Satisfied", []Bytes{{0, 1, 1, true}}, "bytes 0-1/1"},
		{"SatisfiedNoLength", []Bytes{{0, 1, -1, true}}, "bytes 0-1/*"},
		{"NotSatisfied", []Bytes{{-1, -1, 1, false}}, "bytes */1"},
	}
	for _, f := range tbl {
		f := f
		t.Run(f.Name, func(t *testing.T) {
			t.Parallel()
			got, err := FormatResponse(f.In[0])
			if err != nil {
				t.Error(err)
			}
			if want, got := f.Out, got; want != got {
				t.Errorf("want: %q, got: %q", want, got)
			}
		})
	}
}

func TestFormatBytesRequest(t *testing.T) {
	tbl := []brfmt{
		{"Basic", []Bytes{{0, 1, 0, false}}, "bytes=0-1"},
		{"Negative", []Bytes{{-42, -1, 0, false}}, "bytes=-42"},
		{"Open", []Bytes{{0, -1, 0, false}}, "bytes=0-"},
		{"Multiple", []Bytes{
			{0, 1, 0, false},
			{-42, -1, 0, false},
			{0, -1, 0, false},
		}, "bytes=0-1,-42,0-"},
	}
	for _, f := range tbl {
		f := f
		t.Run(f.Name, func(t *testing.T) {
			t.Parallel()
			got, err := FormatRequest(f.In...)
			if err != nil {
				t.Error(err)
			}
			if want, got := f.Out, got; want != got {
				t.Errorf("want: %q, got: %q", want, got)
			}
		})
	}
}

func TestFormatBytesBadRequest(t *testing.T) {
	tbl := []brfmt{
		{"BadRange", []Bytes{{2, 1, 0, false}}, ""},
		{"BadRange", []Bytes{{-1, 1, 0, false}}, ""},
		{"BadRange", []Bytes{{-11, -10, 0, false}}, ""},
		{"NoRange", []Bytes{}, ""},
	}
	for _, f := range tbl {
		f := f
		t.Run(f.Name, func(t *testing.T) {
			t.Parallel()
			t.Log(f.In)
			got, err := FormatRequest(f.In...)
			if err == nil {
				t.Errorf("want: error, got: %q, nil", got)
			}
		})
	}
}

func TestFormatBytesBadResponse(t *testing.T) {
	tbl := []brfmt{
		{"BadLength", []Bytes{{0, 0, -1, false}}, ""},
		{"BadLength", []Bytes{{0, 1200, 0, true}}, ""},
		{"BadRange", []Bytes{{-1, -1, 0, true}}, ""},
		{"BadRange", []Bytes{{1200, 0, 0, true}}, ""},
	}
	for _, f := range tbl {
		f := f
		t.Run(f.Name, func(t *testing.T) {
			t.Parallel()
			_, err := FormatResponse(f.In[0])
			if err == nil {
				t.Error("want: error, got: nil")
			}
		})
	}
}

type endtoend struct {
	Name string
	In   []Bytes
	Out  []Bytes
}

func TestRequest(t *testing.T) {
	tbl := []endtoend{
		{"Single",
			[]Bytes{
				{Start: 0, End: 1200}},
			[]Bytes{
				{Start: 0, End: 1200}},
		},
		{"Last",
			[]Bytes{
				{Start: -1, End: -1}},
			[]Bytes{
				{Start: -1, End: -1}},
		},
		{"Open",
			[]Bytes{
				{Start: 0, End: -1}},
			[]Bytes{
				{Start: 0, End: -1}},
		},
		{"First-Last",
			[]Bytes{
				{Start: 0, End: 1},
				{Start: -1, End: -1}},
			[]Bytes{
				{Start: 0, End: 1},
				{Start: -1, End: -1}},
		},
	}
	for _, h := range tbl {
		h := h
		t.Run(h.Name, func(t *testing.T) {
			t.Parallel()
			o, err := FormatRequest(h.In...)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("header: %q", o)
			got, err := ParseRequest(o)
			if err != nil {
				t.Fatal(err)
			}
			for i, want := range h.Out {
				// Equals method comes from lex_test.go
				want.Equals(t, got[i])
			}
		})
	}
}

func TestResponse(t *testing.T) {
	tbl := []endtoend{
		{"Single",
			[]Bytes{
				{Start: 0, End: 1200, Length: 1200, Satisfied: true}},
			[]Bytes{
				{Start: 0, End: 1200, Length: 1200, Satisfied: true}},
		},
	}
	for _, h := range tbl {
		h := h
		t.Run(h.Name, func(t *testing.T) {
			t.Parallel()
			for i, in := range h.In {
				want := h.Out[i]
				o, err := FormatResponse(in)
				if err != nil {
					t.Fatal(err)
				}
				t.Logf("header: %q", o)
				got, err := ParseResponse(o)
				if err != nil {
					t.Fatal(err)
				}
				// Equals method comes from lex_test.go
				want.Equals(t, got)
			}
		})
	}
}
