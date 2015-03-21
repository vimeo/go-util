package util

import (
    "errors"
    "io"
    "os"
    "sync"
    "time"
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

// ErrReadTimeout is the error used when a read times out before completing.
var ErrReadTimeout = errors.New("read timed out")

// An io.ReadCloser that has a timeout for each underlying Read() function and
// optionally closes the underlying Reader on timeout.
type TimeoutReader struct {
    reader io.ReadCloser
    timeout time.Duration
    closeOnTimeout bool

    timedOut bool
    mutex sync.Mutex
}

// Create a new TimeoutReader.
func NewTimeoutReader(reader io.ReadCloser, timeout time.Duration, closeOnTimeout bool) *TimeoutReader {
    return &TimeoutReader{
        reader: reader,
        timeout: timeout,
        closeOnTimeout: closeOnTimeout,
    }
}

// Closes the TimeoutReader.
// Also closes the underlying Reader if it was not closed already at timeout.
func (this *TimeoutReader) Close() error {
    this.mutex.Lock()
    defer this.mutex.Unlock()
    if this.timedOut && this.closeOnTimeout {
        return nil
    } else {
        return this.reader.Close()
    }
    return nil
}

// Read from the underlying reader.
// If the underlying Read() does not return within the timeout, ErrReadTimeout
// is returned.
func (this *TimeoutReader) Read(p []byte) (int, error) {
    type ReadResponse struct {
        n int
        err error
    }

    this.mutex.Lock()
    if this.timedOut {
        defer this.mutex.Unlock()
        return 0, ErrReadTimeout
    }
    this.mutex.Unlock()

    if this.timeout <= 0 {
        return this.reader.Read(p)
    }

    done := make(chan *ReadResponse, 1)
    defer close(done)
    t := time.After(this.timeout)

    go func() {
        n, err := this.reader.Read(p)
        this.mutex.Lock()
        defer this.mutex.Unlock()
        if !this.timedOut {
            done <- &ReadResponse{n, err}
        }
    }()

    select {
    case <- t:
        this.mutex.Lock()
        this.timedOut = true
        this.mutex.Unlock()
        if this.closeOnTimeout {
            this.reader.Close()
        }
        return 0, ErrReadTimeout
    case resp := <- done:
        return resp.n, resp.err
    }
}
