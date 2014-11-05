package util

import (
    "strings"
)

// Create a map from a single string with separators. This is essentially a
// two-dimensional version of strings.Split().
func TwoDimSplit(s, sep1, sep2 string) map[string]string {
    out := make(map[string]string)
    main_pieces := strings.Split(s, sep1)
    for _, piece := range main_pieces {
        minor_pieces := strings.Split(piece, sep2)
        out[minor_pieces[0]] = minor_pieces[1]
    }
    return out
}
