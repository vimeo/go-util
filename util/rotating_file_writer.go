package util

import (
    "fmt"
    "os"
    "path/filepath"
    "sync"
    "time"
)

// Callback function called from within a Write() or Close() call after rotation.
// `filename` is the name of the file the rotation copied data to.
// `startTime` is the time the file started getting output.
// `closing` indicates if this is the final rotation called on Close().
// `opaque` is the object passed into `NewRotatingFileWriter`.
type RotateCallbackFunc func(filename string, startTime time.Time, closing bool, opaque interface{})

// io.WriteCloser that writes to a rotating file.
type RotatingFileWriter struct {
    mutex sync.Mutex
    filename string
    byteCount int64
    startTime time.Time
    currentFile *os.File
    rotateCallback RotateCallbackFunc
    opaque interface{}
    maxSize int64
    maxDuration time.Duration
}

// Create a new RotatingFileWriter.
// All new writes go to `filename`. If a new write would make `filename` larger
// than `maxSize` or if `maxDuration` has passed since the last rotation,
// `filename` is copied to `filename.<timestamp>`, then truncated before the
// new data is written to it. After rotation, `callback` is called with the
// given `opaque` parameter.
func NewRotatingFileWriter(filename string, callback RotateCallbackFunc,
                           maxSize int64, maxDuration time.Duration,
                           opaque interface{}) (*RotatingFileWriter, error) {
    rfw := new(RotatingFileWriter)
    rfw.filename       = filename
    rfw.rotateCallback = callback
    rfw.opaque         = opaque
    rfw.maxSize        = maxSize
    rfw.maxDuration    = maxDuration

    err := os.MkdirAll(filepath.Dir(filename), 0755)
    if err != nil {
        return nil, err
    }

    fi, err := os.Stat(filename)
    if err == nil {
        rfw.byteCount = fi.Size()
    }

    f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
    if err != nil {
        return nil, err
    }
    rfw.currentFile = f
    rfw.startTime   = time.Now()

    return rfw, nil
}

func (this *RotatingFileWriter) rotate(closing bool) error {
    now := time.Now()

    err := this.currentFile.Close()
    if err != nil {
        return err
    }

    newFilename := this.filename + fmt.Sprintf(".%d", now.UnixNano() / 1000000)
    err = os.Rename(this.filename, newFilename)
    if err != nil {
        return err
    }

    f, err := os.OpenFile(this.filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
    if err != nil {
        return err
    }
    this.currentFile = f
    startTime       := this.startTime
    this.startTime   = now
    this.byteCount   = 0

    if this.rotateCallback != nil {
        this.rotateCallback(newFilename, startTime, closing, this.opaque)
    }

    return nil
}

func (this *RotatingFileWriter) Write(p []byte) (int, error) {
    var err error

    this.mutex.Lock()
    defer this.mutex.Unlock()

    c := int64(len(p))
    if c == 0 {
        return 0, nil
    }

    rotate := false
    if this.maxSize > 0 && this.byteCount > this.maxSize - c {
        rotate = true
    }
    now := time.Now()
    if this.maxDuration > 0 && now.Sub(this.startTime) > this.maxDuration {
        rotate = true
    }

    if rotate {
        err = this.rotate(false)
        if err != nil {
            return 0, err
        }
    }

    n, err := this.currentFile.Write(p)
    this.byteCount += int64(n)
    return n, err
}

func (this *RotatingFileWriter) Close() error {
    this.mutex.Lock()
    defer this.mutex.Unlock()
    err := this.rotate(true)
    if err != nil {
        return err
    }
    return this.currentFile.Close()
}
