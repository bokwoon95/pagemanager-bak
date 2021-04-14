package hypforms

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"runtime"

	"github.com/bokwoon95/erro"
	"github.com/bokwoon95/pagemanager/hyp"
)

type ValidationErrs struct {
	FormErrs  []error
	InputErrs map[string][]error
}

func (e ValidationErrs) Error() string {
	return fmt.Sprintf("form errors: %+v, input errors: %+v", e.FormErrs, e.InputErrs)
}

func MarshalForm(s hyp.Sanitizer, r *http.Request, fn func(*Form)) (template.HTML, error) {
	form := &Form{
		request:    r,
		inputNames: make(map[string]struct{}),
		inputErrs:  make(map[string][]error),
	}
	// read the cookies from the request and ungob any ValidationErrs
	fn(form)
	if len(form.formErrs) > 0 || len(form.inputErrs) > 0 {
		return "", erro.Wrap(ValidationErrs{FormErrs: form.formErrs, InputErrs: form.inputErrs})
	}
	output, err := hyp.Marshal(s, form)
	if err != nil {
		return output, erro.Wrap(err)
	}
	return output, nil
}

func UnmarshalForm(r *http.Request, fn func(*Form)) error {
	r.ParseForm()
	r.ParseMultipartForm(32 << 20)
	form := &Form{
		mode:       FormModeUnmarshal,
		request:    r,
		inputNames: make(map[string]struct{}),
		inputErrs:  make(map[string][]error),
	}
	fn(form)
	if len(form.formErrs) > 0 || len(form.inputErrs) > 0 {
		return erro.Wrap(ValidationErrs{FormErrs: form.formErrs, InputErrs: form.inputErrs})
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
	// write the err into w cookies
	http.Redirect(w, r, url, http.StatusMovedPermanently)
}

func validate(f *Form, name string, value interface{}, validators []Validator) {
	if len(validators) == 0 {
		return
	}
	var stop bool
	var err error
	ctx := f.request.Context()
	ctx = context.WithValue(ctx, contextKey("name"), name)
	for _, validator := range validators {
		stop, err = validator(ctx, value)
		if err != nil {
			f.inputErrs[name] = append(f.inputErrs[name], err)
		}
		if stop {
			return
		}
	}
}
