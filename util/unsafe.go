package util

import (
	"reflect"
	"unsafe"
)

// Convert an unsafe pointer to a byte slice.
// The input buffer has to remain valid for the whole life-cycle of the output
// slice, and users are responsible for freeing the associated memory.
func UnsafeToBytes(buffer unsafe.Pointer, length int) []byte {
	var output []byte

	bufconv := (*reflect.SliceHeader)(unsafe.Pointer(&output))
	bufconv.Data = uintptr(buffer)
	bufconv.Len = length
	bufconv.Cap = length

	return output
}
