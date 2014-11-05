package util

import (
    "fmt"
    "math"
    "strings"
)

// Allows using named bitmask values. Useful for commandline flag options
// that need to be treated as bitmask values. In all functions, names are
// case-insensitive.
type BitMask struct {
    fields map[string]uint64
    values map[uint64]string
    nextVal uint64
}

func (this *BitMask) addName(name string) error {
    if this.nextVal == 0 {
        this.fields = make(map[string]uint64, 64)
        this.values = make(map[uint64]string, 64)
        this.nextVal = 1
    }
    name = strings.ToLower(name)
    if name == "all" || name == "none" {
        return fmt.Errorf("reserved field name: %s", name)
    }
    _, ok := this.fields[name]
    if ok {
        return fmt.Errorf("field already exists: %s", name)
    }
    if this.nextVal == math.MaxUint64 {
        return fmt.Errorf("too many fields")
    }
    this.fields[name] = this.nextVal
    this.values[this.nextVal] = name
    if this.nextVal == uint64(0x8000000000000000) {
        this.nextVal = math.MaxUint64
    } else {
        this.nextVal <<= 1
    }
    return nil
}

// Add list of names to the bitmask. Takes a single string with names
// separated by the '|' character.
func (this *BitMask) Add(names string) error {
    ss := strings.Split(names, "|")
    if len(ss) == 0 {
        return fmt.Errorf("invalid bitmask string")
    }
    for _, sn := range ss {
        err := this.addName(sn)
        if err != nil {
            return err
        }
    }
    return nil
}

// Get the bitmask value for a list of names. Takes a single string with names
// separated by the '|' character.
func (this *BitMask) Parse(s string) (uint64, error) {
    s = strings.ToLower(s)
    if s == "all" {
        v := uint64(0)
        for vn := range this.values {
            v |= vn
        }
        return v, nil
    } else if s == "none" {
        return 0, nil
    }
    ss := strings.Split(s, "|")
    if len(ss) == 0 {
        return 0, fmt.Errorf("invalid bitmask string")
    }
    v := uint64(0)
    for _, sn := range ss {
        vn, ok := this.fields[sn]
        if !ok {
            return 0, fmt.Errorf("field not found: %s", sn)
        }
        v |= vn
    }
    return v, nil
}

// Given a bitmask value, check if the value corresponding to the given
// string is set. Takes a single string with names separated by the '|'
// character.
func (this *BitMask) IsSet(v uint64, s string) bool {
    check, err := this.Parse(s)
    if err == nil && (v & check) != 0 {
        return true
    }
    return false
}

// Get the names for a bitmask value.
func (this *BitMask) Format(v uint64) (string, error) {
    if v == 0 {
        return "none", nil
    }
    s := make([]string, 0)
    vn := uint64(1)
    for vn != uint64(0x8000000000000000) {
        if (v & vn) != 0 {
            sn, ok := this.values[vn]
            if !ok {
                return "", fmt.Errorf("bitmask not found: %d\n", vn)
            }
            s = append(s, sn)
        }
        vn <<= 1
    }
    return strings.Join(s, "|"), nil
}
