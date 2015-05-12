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

type readResponse struct {
    n int
    err error
}

// An io.ReadCloser that has a timeout for each underlying Read() function and
// optionally closes the underlying Reader on timeout.
type TimeoutReader struct {
    reader io.ReadCloser
    timeout time.Duration
    closeOnTimeout bool
    maxReadSize int
    done chan *readResponse
    timer *time.Timer
}

func NewTimeoutReaderSize(reader io.ReadCloser, timeout time.Duration, closeOnTimeout bool, maxReadSize int) *TimeoutReader {
    tr := new(TimeoutReader)
    tr.reader = reader
    tr.timeout = timeout
    tr.closeOnTimeout = closeOnTimeout
    tr.maxReadSize = maxReadSize
    tr.done = make(chan *readResponse, 1)
    if timeout > 0 {
        tr.timer = time.NewTimer(timeout)
    }
    return tr
}

// Create a new TimeoutReader.
func NewTimeoutReader(reader io.ReadCloser, timeout time.Duration, closeOnTimeout bool) *TimeoutReader {
    return NewTimeoutReaderSize(reader, timeout, closeOnTimeout, 0)
}

// Closes the TimeoutReader.
// Also closes the underlying Reader if it was not closed already at timeout.
func (this *TimeoutReader) Close() error {
    return this.reader.Close()
}

// Read from the underlying reader.
// If the underlying Read() does not return within the timeout, ErrReadTimeout
// is returned.
func (this *TimeoutReader) Read(p []byte) (int, error) {
    if this.timeout <= 0 {
        return this.reader.Read(p)
    }

    if this.maxReadSize > 0 && len(p) > this.maxReadSize {
        p = p[:this.maxReadSize]
    }

    // reset the timer
    select {
    case <- this.timer.C:
    default:
    }
    this.timer.Reset(this.timeout)

    // clear the done channel
    select {
    case <- this.done:
    default:
    }

    var timedOut bool
    var finished bool
    var mutex sync.Mutex

    go func() {
        n, err := io.ReadFull(this.reader, p)
        mutex.Lock()
        defer mutex.Unlock()
        finished = true
        if !timedOut {
            this.timer.Stop()
            if err == io.ErrUnexpectedEOF {
                err = nil
            }
            this.done <- &readResponse{n, err}
        }
    }()

    select {
    case <- this.timer.C:
        mutex.Lock()
        defer mutex.Unlock()
        if finished {
            resp := <- this.done
            return resp.n, resp.err
        }
        timedOut = true
        if this.closeOnTimeout {
            this.reader.Close()
        }
        return 0, ErrReadTimeout
    case resp := <- this.done:
        return resp.n, resp.err
    }
}
