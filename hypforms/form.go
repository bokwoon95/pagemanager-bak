package hypforms

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bokwoon95/erro"
	"github.com/bokwoon95/pagemanager/hyp"
)

type FormMode int

const (
	FormModeMarshal   FormMode = 0
	FormModeUnmarshal FormMode = 1
)

type Form struct {
	mode       FormMode
	attrs      hyp.Attributes
	children   []hyp.Element
	request    *http.Request
	inputNames map[string]struct{}
	inputErrs  map[string][]error
	formErrs   []error
}

func (f *Form) AppendHTML(buf *strings.Builder) error {
	if f.mode == FormModeUnmarshal {
		return nil
	}
	// check f.request.Context() for any CSRF token and prepend it into the form as necessary
	// or should this be done in a hook?
	f.attrs.Tag = "form"
	err := hyp.AppendHTML(buf, f.attrs, f.children)
	if err != nil {
		return erro.Wrap(err)
	}
	return nil
}

func (f *Form) registerName(name string, skip int) {
	if _, ok := f.inputNames[name]; ok {
		file, line, _ := caller(skip + 1)
		f.formErrs = append(f.formErrs, fmt.Errorf("%s:%d duplicate name: %s", file, line, name))
	}
	f.inputNames[name] = struct{}{}
}

func (f *Form) Set(selector string, attributes map[string]string, children ...hyp.Element) {
	if f.mode == FormModeUnmarshal {
		return
	}
	f.attrs = hyp.ParseAttributes(selector, attributes)
	f.children = children
}

func (f *Form) Append(selector string, attributes map[string]string, children ...hyp.Element) {
	if f.mode == FormModeUnmarshal {
		return
	}
	f.children = append(f.children, hyp.H(selector, attributes, children...))
}

func (f *Form) AppendElements(children ...hyp.Element) {
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

func (f *Form) FormError(err error) {
	f.formErrs = append(f.formErrs, err)
}

func (f *Form) InputError(name string, err error) {
	f.inputErrs[name] = append(f.inputErrs[name], err)
}
