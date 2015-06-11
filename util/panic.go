package util

import (
    "fmt"
    "io"
    "os"
    "os/signal"
    "runtime"
    "syscall"
    "time"
)

// panic recovery handler that prints a backtrace to an io.Writer and exits.
// The default handler always prints to stderr and has no timestamp.
// To use, call this function with defer, e.g. in the main() function.
func PanicBacktrace(w io.Writer) {
    r := recover()
    if r != nil {
        b := make([]byte, 32768)
        runtime.Stack(b, true)
        fmt.Fprintf(w, "%s panic: %s\n", time.Now().UTC().Format("2006/01/02 15:04:05"), r)
        fmt.Fprintf(w, "%s\n", string(b))
        os.Exit(1)
    }
}

// Alternative SIGQUIT handler that dumps stack trace to any io.Writer and exits.
func SigQuitBacktrace(w io.Writer) {
    qsc := make(chan os.Signal, 1)
    signal.Notify(qsc, syscall.SIGQUIT)
    <- qsc
    signal.Stop(qsc)
    fmt.Fprintf(w, "SIGQUIT\n\n")
    buf := make([]byte, 65536)
    buf = buf[:runtime.Stack(buf, true)]
    fmt.Fprintln(w, string(buf))
    os.Exit(1)
}
