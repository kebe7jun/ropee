package test

import (
	"fmt"
	"io"
)

type Body struct {
	s             string
	b             []byte
	isOpen        bool
	closeAttempts int
}

// NewBody creates a new instance of Body.
func NewBody(s string) *Body {
	return (&Body{s: s}).reset()
}

func (body *Body) Read(b []byte) (n int, err error) {
	if !body.IsOpen() {
		return 0, fmt.Errorf("ERROR: Body has been closed")
	}
	if len(body.b) == 0 {
		return 0, io.EOF
	}
	n = copy(b, body.b)
	body.b = body.b[n:]
	return n, nil
}

// Close closes the body.
func (body *Body) Close() error {
	if body.isOpen {
		body.isOpen = false
		body.closeAttempts++
	}
	return nil
}

// IsOpen returns true if the Body has not been closed, false otherwise.
func (body *Body) IsOpen() bool {
	return body.isOpen
}

func (body *Body) reset() *Body {
	body.isOpen = true
	body.b = []byte(body.s)
	return body
}

// Length returns the number of bytes in the body.
func (body *Body) Length() int64 {
	if body == nil {
		return 0
	}
	return int64(len(body.b))
}
