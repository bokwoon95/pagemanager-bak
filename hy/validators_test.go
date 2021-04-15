package hy

import (
	"errors"
	"io"
	"testing"

	"github.com/bokwoon95/pagemanager/testutil"
)

func Test_Error(t *testing.T) {
	is := testutil.New(t)
	is.True(!errors.Is(io.EOF, errors.New("EOF")))
	is.True(io.EOF != errors.New("EOF"))
	e1 := errors.New("bruh")
	e2 := errors.New("bruh")
	is.True(e1 != e2)
	is.True(errors.Is(e1, e2))
}
