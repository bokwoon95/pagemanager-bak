package hy

import (
	"context"
	"strings"
)

type Validator func(ctx context.Context, value interface{}) (stop bool, err *Error)

func Validate(value interface{}, validators ...Validator) []*Error {
	return ValidateContext(context.Background(), value, validators...)
}

func ValidateContext(ctx context.Context, value interface{}, validators ...Validator) []*Error {
	var stop bool
	var err *Error
	var errs []*Error
	for _, validator := range validators {
		stop, err = validator(ctx, value)
		if err != nil {
			errs = append(errs, err)
		}
		if stop {
			return errs
		}
	}
	return errs
}

// errors are delimited by RS(30)

type Error struct {
	Msg string
	Err *Error
}

const rs = 30

func NewError(text string) *Error {
	return &Error{Msg: text}
}

func WrapError(childErr *Error, text string) *Error {
	return &Error{Msg: text, Err: childErr}
}

func (e *Error) String() string {
	err := e
	buf := &strings.Builder{}
	buf.WriteString(err.Msg)
	for err.Err != nil {
		err = err.Err
		buf.WriteString(": ")
		buf.WriteString(err.Msg)
	}
	return buf.String()
}

func (e *Error) Error() string {
	return e.Msg
}

func (e *Error) Unwrap() error {
	return e.Err
}

func (e *Error) MarshalBinary() (data []byte, err error) {
	data = append(data, e.Msg...)
	data = append(data, rs)
	if e.Err != nil {
		b, err := e.Err.MarshalBinary()
		if err != nil {
			return data, err
		}
		data = append(data, b...)
	}
	return data, nil
}

func (e *Error) UnmarshalBinary(data []byte) (err error) {
	i := len(data)
	for j, c := range data {
		if c == rs {
			i = j
			break
		}
	}
	e.Msg = string(data[:i])
	if e.Err != nil {
		return e.Err.UnmarshalBinary(data[i+1:])
	}
	return nil
}
