package util

import (
    "io"
    "os"
)

// Copy a local file.
func CopyFile(dst, src string) error {
    sf, err := os.Open(src)
    if err != nil {
        return err
    }
    defer sf.Close()
    df, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer df.Close()
    _, err = io.Copy(df, sf)
    return err
}
