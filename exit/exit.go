// Copyright 2012, Google Inc.
// All rights reserved.
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
//    * Redistributions of source code must retain the above copyright
//      notice, this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above
//      copyright notice, this list of conditions and the following disclaimer
//      in the documentation and/or other materials provided with the
//      distribution.
//    * Neither the name of Google Inc. nor the names of its
//      contributors may be used to endorse or promote products derived from
//      this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

/*
Package exit provides an alternative to os.Exit(int).

Unlike os.Exit(int), exit.Return(int) will run deferred functions before
terminating. It's effectively like a return from main(), except you can specify
the exit code.

Defer a call to exit.Recover() at the beginning of main().
Use exit.Return(int) to initiate an exit.

	func main() {
		defer exit.Recover()
		defer cleanup()
		...
		if err != nil {
			// Return from main() with a non-zero exit code,
			// making sure to run deferred cleanup.
			exit.Return(1)
		}
		...
	}

All functions deferred *after* defer exit.Recover() will be executed before
the exit. This is why the defer for this package should be the first statement
in main().

NOTE: This mechanism only works if exit.Return() is called from the same
goroutine that deferred exit.Recover(). Usually this means Return() should
only be used from within main(), or within functions that are only ever
called from main(). See Recover() and Return() for more details.
*/
package exit

import (
	"os"
)

type exitCode int

var (
	exitFunc = os.Exit // can be faked out for testing
)

// Recover should be deferred as the first line of main(). It recovers the
// panic initiated by Return and converts it to a call to os.Exit. Any
// functions deferred after Recover in the main goroutine will be executed
// prior to exiting. Recover will re-panic anything other than the panic it
// expects from Return.
func Recover() {
	doRecover(recover())
}

func doRecover(err interface{}) {
	if err == nil {
		return
	}

	switch code := err.(type) {
	case exitCode:
		exitFunc(int(code))
	default:
		panic(err)
	}
}

// Return initiates a panic that sends the return code to the deferred Recover,
// executing other deferred functions along the way. When the panic reaches
// Recover, the return code will be passed to os.Exit. This should only be
// called from the main goroutine.
func Return(code int) {
	panic(exitCode(code))
}
