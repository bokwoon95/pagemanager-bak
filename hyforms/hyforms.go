package hyforms

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"runtime"

	"github.com/bokwoon95/erro"
	"github.com/bokwoon95/pagemanager/hy"
)

type ValidationError struct {
	FormErrMsgs  []string
	InputErrMsgs map[string][]string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("form errors: %+v, input errors: %+v", e.FormErrMsgs, e.InputErrMsgs)
}

func MarshalForm(s hy.Sanitizer, w http.ResponseWriter, r *http.Request, fn func(*Form)) (template.HTML, error) {
	form := &Form{
		request:    r,
		inputNames: make(map[string]struct{}),
		inputErrMsgs:  make(map[string][]string),
	}
	func() {
		c, _ := r.Cookie("hyforms.ValidationError")
		if c == nil {
			return
		}
		defer http.SetCookie(w, &http.Cookie{Name: "hyforms.ValidationError", MaxAge: -1})
		b, err := base64.RawURLEncoding.DecodeString(c.Value)
		if err != nil {
			return
		}
		validationErr := &ValidationError{}
		err = gob.NewDecoder(bytes.NewReader(b)).Decode(validationErr)
		if err != nil {
			return
		}
		form.formErrMsgs = validationErr.FormErrMsgs
		form.inputErrMsgs = validationErr.InputErrMsgs
	}()
	fn(form)
	if len(form.marshalErrMsgs) > 0 {
		return "", erro.Wrap(fmt.Errorf("marshal errors %v", form.marshalErrMsgs))
	}
	output, err := hy.MarshalElement(s, form)
	if err != nil {
		return output, erro.Wrap(err)
	}
	return output, nil
}

func UnmarshalForm(w http.ResponseWriter, r *http.Request, fn func(*Form)) error {
	r.ParseForm()
	form := &Form{
		mode:       FormModeUnmarshal,
		request:    r,
		inputNames: make(map[string]struct{}),
		inputErrMsgs:  make(map[string][]string),
	}
	fn(form)
	if len(form.formErrMsgs) > 0 || len(form.inputErrMsgs) > 0 {
		validationErr := ValidationError{FormErrMsgs: form.formErrMsgs, InputErrMsgs: form.inputErrMsgs}
		func() {
			buf := &bytes.Buffer{}
			err := gob.NewEncoder(buf).Encode(validationErr)
			if err != nil {
				return
			}
			http.SetCookie(w, &http.Cookie{
				Name:   "hyforms.ValidationError",
				Value:  base64.RawURLEncoding.EncodeToString(buf.Bytes()),
				MaxAge: 5,
			})
		}()
		return erro.Wrap(&ValidationError{FormErrMsgs: form.formErrMsgs, InputErrMsgs: form.inputErrMsgs})
	}
	return nil
}

func caller(skip int) (file string, line int, function string) {
	var pc [1]uintptr
	n := runtime.Callers(skip+2, pc[:])
	if n == 0 {
		return "???", 1, "???"
	}
	frame, _ := runtime.CallersFrames(pc[:n]).Next()
	return frame.File, frame.Line, frame.Function
}

func Redirect(w http.ResponseWriter, r *http.Request, url string, err error) {
	defer http.Redirect(w, r, url, http.StatusMovedPermanently)
	validationErr := &ValidationError{}
	if !errors.As(err, &validationErr) {
		return
	}
	buf := &bytes.Buffer{}
	err = gob.NewEncoder(buf).Encode(*validationErr)
	if err != nil {
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:  "hyforms.ValidationError",
		Value: base64.RawURLEncoding.EncodeToString(buf.Bytes()),
	})
}

func validate(f *Form, name string, value interface{}, validators []Validator) {
	if len(validators) == 0 {
		return
	}
	var stop bool
	var errMsg string
	ctx := f.request.Context()
	ctx = context.WithValue(ctx, ctxKeyName, name)
	for _, validator := range validators {
		stop, errMsg = validator(ctx, value)
		if errMsg != "" {
			f.inputErrMsgs[name] = append(f.inputErrMsgs[name], errMsg)
		}
		if stop {
			return
		}
	}
}
