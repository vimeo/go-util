// Package httprange provides RFC7233 byte range parsing.
//
// This RFC provides for requesting sections of a resource via the "Range" header.
// The spec defines a generic way format to request ranges and respond to requests,
// and also specfies how to request bytes.
//
// This package provides functions for parsing and formatting byte ranges.
package httprange

import (
	"bytes"
	"fmt"
	"strconv"
)

// Bytes is the bytes range type specified in the RFC.
type Bytes struct {
	Start     int64
	End       int64
	Length    int64
	Satisfied bool
}

func (b Bytes) fmtRequest() (string, error) {
	switch {
	case b.Start < 0 && b.End == -1:
		return strconv.FormatInt(b.Start, 10), nil
	case b.Start >= 0 && b.End == -1:
		return strconv.FormatInt(b.Start, 10) + "-", nil
	case b.Start >= 0 && b.End >= 0:
		if b.End < b.Start {
			break
		}
		s, e := strconv.FormatInt(b.Start, 10), strconv.FormatInt(b.End, 10)
		return s + "-" + e, nil
	default:
	}
	return "", fmt.Errorf("invalid request range")
}

// ErrNotByteUnit is returned from "Parse" functions if the returned unit type is not "bytes".
var ErrNotByteUnit = fmt.Errorf(`httprange: expected "bytes" unit type`)

// ParseResponse parses a Content-Range header, expecting it to be a bytes range.
//
// The returned Bytes has Satisfied set if the response returned an actual
// range. If Satisfied is true and Length is -1, a Length was not provided in
// the response.
func ParseResponse(h string) (Bytes, error) {
	l := lexResponse(h)
	var r Bytes
	for {
		switch t := l.step(); t.kind {
		case itemUnit:
			if t.tok != "bytes" {
				return Bytes{}, ErrNotByteUnit
			}
		case itemStart:
			if t.tok == "*" {
				r.Start = -1
				r.End = -1
				break
			}
			r.Satisfied = true
			i, err := strconv.ParseInt(t.tok, 10, 64)
			if err != nil {
				return Bytes{}, err
			}
			r.Start = i
		case itemEnd:
			i, err := strconv.ParseInt(t.tok, 10, 64)
			if err != nil {
				return Bytes{}, err
			}
			r.End = i
		case itemLength:
			if t.tok == "*" {
				if !r.Satisfied {
					return Bytes{}, fmt.Errorf("nonsense header: %q", h)
				}
				r.Length = -1
				break
			}
			i, err := strconv.ParseInt(t.tok, 10, 64)
			if err != nil {
				return Bytes{}, err
			}
			r.Length = i
		case itemEOF:
			return r, nil
		case itemError:
			return Bytes{}, fmt.Errorf("%s", t.tok)
		default:
			return Bytes{}, fmt.Errorf("lexer error: what's a %q?", t)
		}
	}
}

// ParseRequest parses an incoming Range header.
//
// The returned ranges are not coalesced.
// The Satisfied member is unset and should be ignored in the returned Bytes.
// Length is always set to 0.
//
// An "open" request has a non-negative Start and -1 as End. A "from end"
// request has a negative Start and -1 as End.
func ParseRequest(h string) ([]Bytes, error) {
	l := lexRequest(h)
	var r []Bytes
	var cur *Bytes
	for {
		switch t := l.step(); t.kind {
		case itemUnit:
			if t.tok != "bytes" {
				return nil, ErrNotByteUnit
			}
		case itemStart:
			r = append(r, Bytes{})
			cur = &r[len(r)-1]
			i, err := strconv.ParseInt(t.tok, 10, 64)
			if err != nil {
				return nil, err
			}
			cur.Start = i
		case itemEnd:
			if t.tok == "" {
				cur.End = -1
				break
			}
			i, err := strconv.ParseInt(t.tok, 10, 64)
			if err != nil {
				return nil, err
			}
			cur.End = i
			if cur.End < cur.Start {
				return nil, fmt.Errorf("invalid range: %d-%d", cur.Start, cur.End)
			}
		case itemEOF:
			return r, nil
		case itemError:
			return nil, fmt.Errorf("%s", t.tok)
		default:
			return nil, fmt.Errorf("lexer error: what's a %q?", t)
		}
	}
}

// FormatRequest constructs a string suitable for using as a Range header.
//
// The Length and Satisfied members are ignored.
func FormatRequest(r ...Bytes) (string, error) {
	if len(r) == 0 {
		return "", fmt.Errorf("no ranges provided")
	}
	b := bytes.NewBuffer([]byte("bytes="))
	for i, br := range r {
		if i != 0 {
			b.WriteString(",")
		}
		re, err := br.fmtRequest()
		if err != nil {
			return "", err
		}
		b.WriteString(re)
	}
	return b.String(), nil
}

// FormatResponse constructs a string suitable for using as a Content-Range header.
//
// A caller must indicate if the range was satisfied by setting Satisfied. If
// Satisfied is not set, Length must be non-negative and Start and End are
// ignored.
//
// When a range has been satisfied, Length can be set to -1 to be omitted from
// the header.
func FormatResponse(b Bytes) (string, error) {
	if !b.Satisfied {
		if b.Length < 0 {
			return "", fmt.Errorf("invalid response length")
		}
		return "bytes */" + strconv.FormatInt(b.Length, 10), nil
	}
	if (b.Start < 0 || b.End < 0) ||
		(b.End < b.Start) {
		return "", fmt.Errorf("invalid response range")
	}
	s, e := strconv.FormatInt(b.Start, 10), strconv.FormatInt(b.End, 10)
	if b.Length < 0 {
		return "bytes " + s + "-" + e + "/*", nil
	}
	if (b.Length < b.Start) ||
		(b.Length < b.End) {
		return "", fmt.Errorf("invalid response range")
	}
	return "bytes " + s + "-" + e + "/" + strconv.FormatInt(b.Length, 10), nil
}
