package crc32combine

import (
	"hash/crc32"
	"testing"
)

func testCRC32Combine(correct uint32, parts []uint32, lens []int64) bool {
	if len(parts) == 0 || len(lens) == 0 {
		return false
	}

	if len(parts) != len(lens) {
		return false
	}

	test := parts[0]
	for i := 1; i < len(parts); i++ {
		test = CRC32Combine(crc32.Castagnoli, test, parts[i], lens[i])
	}

	if test != correct {
		return false
	}

	return true
}

func TestCRC32Combine(t *testing.T) {
	var crc1 uint32 = 3835734695
	test1Crcs := []uint32{683702737, 3632182834, 3228133190}
	test1CrcsLen := []int64{23801619, 23801619, 23801619}

	if !testCRC32Combine(crc1, test1Crcs, test1CrcsLen) {
		t.Error("Test 1 CRC not equal")
	}

	var crc2 uint32 = 4030990169
	test2Crcs := []uint32{252150019, 2884243012, 3630597959, 3126674565, 1131679063, 790522791, 318556634, 2167552018, 1181691535, 2299024249}
	test2CrcsLen := []int64{100000000, 100000000, 100000000, 100000000, 100000000, 100000000, 100000000, 100000000, 100000000, 28670754}
	if !testCRC32Combine(crc2, test2Crcs, test2CrcsLen) {
		t.Error("Test 2 CRC not equal")
	}
}
