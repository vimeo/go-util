package httprange

import "testing"

type lout struct {
	Name, In string
	Out      []token
}

// lexCmp takes a funciton that returns a lexer and a lout struct, and returns a
// function that tests that the token stream is as predicted.
func lexCmp(f func(string) *lexer, h lout) func(*testing.T) {
	l := f(h.In)
	return func(t *testing.T) {
		t.Parallel()
		t.Logf("input: %q", h.In)
		for got, i := l.step(), 0; got.kind != itemEOF; got, i = l.step(), i+1 {
			if i == len(h.Out) {
				t.Errorf("too many tokens, only wanted %d", len(h.Out))
			}
			want := h.Out[i]
			t.Logf("%v == %v", want, got)
			if want.kind != got.kind {
				t.Errorf("want: %q, got: %q", want.kind, got.kind)
			}
			if want.tok != got.tok {
				t.Errorf("want: %q, got: %q", want.tok, got.tok)
			}
		}
	}
}

func TestResponseWellFormed(t *testing.T) {
	hdrs := []lout{
		{"Full", "bytes 0-1200/2400", []token{
			{itemUnit, "bytes"},
			{itemStart, "0"},
			{itemEnd, "1200"},
			{itemLength, "2400"}}},
		{"416", "bytes */1", []token{
			{itemUnit, "bytes"},
			{itemStart, "*"},
			{itemLength, "1"}}},
		{"NoLength", "bytes 1-1/*", []token{
			{itemUnit, "bytes"},
			{itemStart, "1"},
			{itemEnd, "1"},
			{itemLength, "*"}}},
	}
	for _, h := range hdrs {
		t.Run(h.Name, lexCmp(lexResponse, h))
	}
}

func TestResponseMalformed(t *testing.T) {
	hdrs := []lout{
		{"Gibberish", "aslongasthere'snospaces", []token{
			{itemError, `wanted a space (got "aslongasthere'snospaces")`}}},
		{"BadStart", "bytes a-500/*", []token{
			{itemUnit, "bytes"},
			{itemError, `wanted an int-like (got "a")`}}},
		{"BadRange", "bytes 0to500/*", []token{
			{itemUnit, "bytes"},
			{itemError, `wanted an int-like (got "0t")`}}},
		{"BadEnd", "bytes 0-ff/*", []token{
			{itemUnit, "bytes"},
			{itemStart, "0"},
			{itemError, `wanted an int-like (got "f")`}}},
		{"BadRangeSep", "bytes 0-1Len*", []token{
			{itemUnit, "bytes"},
			{itemStart, "0"},
			{itemError, `wanted an int-like (got "1L")`}}},
		{"BadLength", "bytes 0-1/ff", []token{
			{itemUnit, "bytes"},
			{itemStart, "0"},
			{itemEnd, "1"},
			{itemError, `wanted int-like until end of string (got "f")`}}},
	}
	for _, h := range hdrs {
		t.Run(h.Name, lexCmp(lexResponse, h))
	}
}

func TestRequestWellFormed(t *testing.T) {
	hdrs := []lout{
		{"Single", "bytes=0-1200", []token{
			{itemUnit, "bytes"},
			{itemStart, "0"},
			{itemEnd, "1200"}}},
		{"Backwards", "bytes=-1200", []token{
			{itemUnit, "bytes"},
			{itemStart, "-1200"},
			{itemEnd, ""}}},
		{"Multiple", "bytes=0-1200,4096-5296", []token{
			{itemUnit, "bytes"},
			{itemStart, "0"},
			{itemEnd, "1200"},
			{itemStart, "4096"},
			{itemEnd, "5296"}}},
		{"MultipleBackwards", "bytes=0-1200,-1200", []token{
			{itemUnit, "bytes"},
			{itemStart, "0"},
			{itemEnd, "1200"},
			{itemStart, "-1200"},
			{itemEnd, ""}}},
		{"MultipleBackwards", "bytes=-1200,0-1200", []token{
			{itemUnit, "bytes"},
			{itemStart, "-1200"},
			{itemEnd, ""},
			{itemStart, "0"},
			{itemEnd, "1200"}}},
	}
	for _, h := range hdrs {
		t.Run(h.Name, lexCmp(lexRequest, h))
	}
}

func TestRequestMalformed(t *testing.T) {
	hdrs := []lout{
		{"BadUnit", "bytes 0-1200", []token{
			{itemError, `wanted a '=' (got "bytes 0-1200")`}}},
		{"BadRangeStart", "bytes=1200", []token{
			{itemUnit, "bytes"},
			{itemError, `wanted int-like until '-' (got "1200")`}}},
		{"BadRangeEnd", "bytes=0-ff", []token{
			{itemUnit, "bytes"},
			{itemStart, "0"},
			{itemError, `wanted int-like until ',' or eof (got "f")`}}},
		{"BadBackwardsRange", "bytes=-f", []token{
			{itemUnit, "bytes"},
			{itemError, `wanted int-like until ',' or eof (got "-f")`}}},
	}
	for _, h := range hdrs {
		t.Run(h.Name, lexCmp(lexRequest, h))
	}
}
