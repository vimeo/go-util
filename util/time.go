package util

import (
    "encoding/json"
    "fmt"
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

// JSON-friendly time.Duration
type Duration struct {
    time.Duration
}

func (d *Duration) UnmarshalJSON(b []byte) (err error) {
    if b[0] == '"' {
        sd := string(b[1 : len(b)-1])
        d.Duration, err = time.ParseDuration(sd)
        return
    }

    var id int64
    id, err = json.Number(string(b)).Int64()
    d.Duration = time.Duration(id)

    return
}

func (d Duration) MarshalJSON() (b []byte, err error) {
    return []byte(fmt.Sprintf(`"%s"`, d.String())), nil
}
