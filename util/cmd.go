package util

import (
    "errors"
    "os/exec"
    "time"
)

// ErrReadTimeout is the error used when a command times out before completing.
var ErrCommandTimeout = errors.New("command timed out")

// Run a command, killing it if it does not finish before the specified timeout
func RunCommandWithTimeout(cmd *exec.Cmd, timeout time.Duration) error {
    done := make(chan error, 1)
    t    := time.After(timeout)

    err := cmd.Start()
    if err != nil {
        return err
    }
    go func() {
        done <- cmd.Wait()
    }()
    select {
    case err := <- done:
        return err
    case <- t:
        cmd.Process.Kill()
        return ErrCommandTimeout
    }
}
