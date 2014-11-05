package util

import (
    "time"
)

// Clip a time.Duration to a given range.
func ClipDuration(v, min, max time.Duration) time.Duration {
    if v < min {
        return min
    } else if v > max {
        return max
    } else {
        return v
    }
}
