package httprange

import "fmt"

func ExampleFormatRequest_multi() {
	r, err := FormatRequest(Bytes{Start: 0, End: 1}, Bytes{Start: 4095, End: 4096})
	if err != nil {
		panic(err)
	}
	fmt.Println(r)
	// Output: bytes=0-1,4095-4096
}

func ExampleFormatRequest_end() {
	r, err := FormatRequest(Bytes{Start: -4096, End: -1})
	if err != nil {
		panic(err)
	}
	fmt.Println(r)
	// Output: bytes=-4096
}

func ExampleFormatRequest_open() {
	r, err := FormatRequest(Bytes{Start: 0, End: -1})
	if err != nil {
		panic(err)
	}
	fmt.Println(r)
	// Output: bytes=0-
}

func ExampleFormatResponse_noLength() {
	r, err := FormatResponse(Bytes{Start: 0, End: 4095, Length: -1, Satisfied: true})
	if err != nil {
		panic(err)
	}
	fmt.Println(r)
	// Output: bytes 0-4095/*
}

func ExampleFormatResponse_length() {
	r, err := FormatResponse(Bytes{Start: 0, End: 4095, Length: 4096, Satisfied: true})
	if err != nil {
		panic(err)
	}
	fmt.Println(r)
	// Output: bytes 0-4095/4096
}
func ExampleFormatResponse_unsatisfied() {
	r, err := FormatResponse(Bytes{Length: 4096, Satisfied: false})
	if err != nil {
		panic(err)
	}
	fmt.Println(r)
	// Output: bytes */4096
}
