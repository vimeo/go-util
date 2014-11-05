package util

import (
    "math"
)

// Calculate a Gaussian function at 'x', with center of 0.0, max peak of 1.0,
// and standard deviation of 'sigma'.
func Gaussian(x, sigma float64) float64 {
    return math.Exp(-(x * x) / (2.0 * sigma * sigma))
}

// Round a float64 to the nearest int64
func Lrint(x float64) int64 {
    if math.IsNaN(x) || math.IsInf(x, 0) {
        return 0
    }

    sign := float64(1.0)
    if x < 0 {
        sign = -1
        x *= -1
    }

    _, frac := math.Modf(x)
    if frac >= 0.5 {
        x = math.Ceil(x)
    } else {
        x = math.Floor(x)
    }

    return int64(x * sign)
}

// Clip an int to a given range.
func ClipInt(v, min, max int) int {
    if v < min {
        return min
    } else if v > max {
        return max
    } else {
        return v
    }
}

// Get the lowest of multiple int values.
func MinInt(a ...int) int {
    switch len(a) {
    case 0:
        return 0
    case 1:
        return a[0]
    case 2:
        if a[0] > a[1] {
            return a[1]
        } else {
            return a[0]
        }
    default:
        min := math.MaxInt32
        for _, i := range a {
            if i < min {
                min = i
            }
        }
        return min
    }
}

// Get the highest of multiple int values.
func MaxInt(a ...int) int {
    switch len(a) {
    case 0:
        return 0
    case 1:
        return a[0]
    case 2:
        if a[0] > a[1] {
            return a[0]
        } else {
            return a[1]
        }
    default:
        max := math.MinInt32
        for _, i := range a {
            if i > max {
                max = i
            }
        }
        return max
    }
}
