package hyperform

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/bokwoon95/erro"
)

type FormMode int

const (
	FormModeMarshal   FormMode = 0
	FormModeUnmarshal FormMode = 1
)

type FormErrors struct {
	inputErrors map[string][]error
	formErrors  []error
}

func (e FormErrors) Error() string {
	return fmt.Sprintf("%#v", e)
}

type Form struct {
	selector   string
	attributes map[string]string
	children   []Element
	mode       FormMode
	request    *http.Request
	names      map[string]struct{}
	errors     FormErrors
}

func (f *Form) Mode() FormMode         { return f.mode }
func (f *Form) Request() *http.Request { return f.request }

func (f *Form) appendHTML(buf *strings.Builder) error {
	if f.mode == FormModeUnmarshal {
		return nil
	}
	// check f.request.Context() for any CSRF token and prepend it into the form as necessary
	return appendHTML(buf, f.selector, f.attributes, f.children)
}

func (f *Form) Set(selector string, attributes map[string]string, children ...Element) {
	if f.mode == FormModeUnmarshal {
		return
	}
	f.selector = selector
	f.attributes = attributes
	f.children = children
}

func (f *Form) Append(selector string, attributes map[string]string, children ...Element) {
	if f.mode == FormModeUnmarshal {
		return
	}
	f.children = append(f.children, H(selector, attributes, children...))
}

func (f *Form) AppendElements(children ...Element) {
	if f.mode == FormModeUnmarshal {
		return
	}
	f.children = append(f.children, children...)
}

func (f *Form) Unmarshal(fn func()) {
	if f.mode != FormModeUnmarshal {
		return
	}
	fn()
}

func (f *Form) FormError(err error) {
	f.errors.formErrors = append(f.errors.formErrors, err)
}

func (f *Form) InputError(name string, err error) {
	f.errors.inputErrors[name] = append(f.errors.inputErrors[name], err)
}

func MarshalForm(w http.ResponseWriter, r *http.Request, fn func(*Form)) (template.HTML, error) {
	form := &Form{
		selector:   "form",
		attributes: make(map[string]string),
		names:      make(map[string]struct{}),
	}
	fn(form)
	return Marshal(w, r, form)
}

func UnmarshalForm(w http.ResponseWriter, r *http.Request, fn func(*Form)) error {
	err := r.ParseForm()
	if err != nil {
		return erro.Wrap(err)
	}
	form := &Form{
		selector:   "form",
		attributes: make(map[string]string),
		mode:       FormModeUnmarshal,
		request:    r,
		names:      make(map[string]struct{}),
	}
	fn(form)
	return form.errors
}

// TODO: remove this,
// https://medium.com/@owlwalks/dont-parse-everything-from-client-multipart-post-golang-9280d23cd4ad
func UnmarshalMultipartForm(w http.ResponseWriter, r *http.Request, fn func(*Form), maxMemory int64) error {
	err := r.ParseMultipartForm(maxMemory)
	if err != nil {
		return erro.Wrap(err)
	}
	form := &Form{
		selector:   "form",
		attributes: make(map[string]string),
		mode:       FormModeUnmarshal,
		request:    r,
		names:      make(map[string]struct{}),
	}
	fn(form)
	return form.errors
}
