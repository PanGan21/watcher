package main

import "io"

type bytesErr struct {
	bytes []byte
	err   error
}

// stdinReader bypasses stdin to child processes
//
// cmd.Wait() blocks until stdin.Read() returns.
// so stdinReader.Read() returns EOF when the child process exited.
type stdinReader struct {
	input    <-chan bytesErr
	chldDone <-chan struct{}
}

func (s *stdinReader) Read(b []byte) (int, error) {
	select {
	case be, ok := <-s.input:
		if !ok {
			return 0, io.EOF
		}
		return copy(b, be.bytes), be.err
	case <-s.chldDone:
		return 0, io.EOF
	}
}
