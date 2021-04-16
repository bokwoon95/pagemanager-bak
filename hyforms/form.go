package hyforms

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bokwoon95/erro"
	"github.com/bokwoon95/pagemanager/hy"
)

type FormMode int

const (
	FormModeMarshal   FormMode = 0
	FormModeUnmarshal FormMode = 1
)

type Form struct {
	mode           FormMode
	attrs          hy.Attributes
	children       []hy.Element
	request        *http.Request
	inputNames     map[string]struct{}
	inputErrMsgs   map[string][]string
	formErrMsgs    []string
	marshalErrMsgs []string
}

func (f *Form) AppendHTML(buf *strings.Builder) error {
	if f.mode == FormModeUnmarshal {
		return nil
	}
	// check f.request.Context() for any CSRF token and prepend it into the form as necessary
	// or should this be done in a hook?
	f.attrs.Tag = "form"
	err := hy.AppendHTML(buf, f.attrs, f.children)
	if err != nil {
		return erro.Wrap(err)
	}
	return nil
}

func (f *Form) registerName(name string, skip int) {
	if f.mode == FormModeUnmarshal {
		return
	}
	if _, ok := f.inputNames[name]; ok {
		file, line, _ := caller(skip + 1)
		f.marshalErrMsgs = append(f.marshalErrMsgs, fmt.Sprintf("%s:%d duplicate name: %s", file, line, name))
	}
	f.inputNames[name] = struct{}{}
}

func (f *Form) Set(selector string, attributes map[string]string, children ...hy.Element) {
	if f.mode == FormModeUnmarshal {
		return
	}
	f.attrs = hy.ParseAttributes(selector, attributes)
	f.children = children
}

func (f *Form) Append(selector string, attributes map[string]string, children ...hy.Element) {
	if f.mode == FormModeUnmarshal {
		return
	}
	f.children = append(f.children, hy.H(selector, attributes, children...))
}

func (f *Form) AppendElements(children ...hy.Element) {
	if f.mode == FormModeUnmarshal {
		return
	}
	f.children = append(f.children, children...)
}

func (f *Form) Unmarshal(unmarshaller func()) {
	if f.mode != FormModeUnmarshal {
		return
	}
	unmarshaller()
}

func (f *Form) ErrMsgs() []string {
	return f.formErrMsgs
}

func (f *Form) AddErrMsgs(errMsgs ...string) {
	f.formErrMsgs = append(f.formErrMsgs, errMsgs...)
}

func (f *Form) AddInputErrMsgs(inputName string, errMsgs ...string) {
	f.inputErrMsgs[inputName] = append(f.inputErrMsgs[inputName], errMsgs...)
}
