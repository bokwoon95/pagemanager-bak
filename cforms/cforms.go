package cforms

import (
	"net/http"
	"strings"

	"github.com/bokwoon95/erro"
)

// https://developer.mozilla.org/en-US/docs/Glossary/Empty_element
var singletonElements = map[string]struct{}{
	"AREA": {}, "BASE": {}, "BR": {}, "COL": {}, "EMBED": {}, "HR": {}, "IMG": {}, "INPUT": {},
	"LINK": {}, "META": {}, "PARAM": {}, "SOURCE": {}, "TRACK": {}, "WBR": {},
}

type Element interface {
	AppendHTML(*strings.Builder) error
}

type Attr map[string]string

type Txt struct{ Text string }

func (t Txt) AppendHTML(buf *strings.Builder) error {
	buf.WriteString(t.Text)
	return nil
}

type parsedSelector struct {
	tag        string
	id         string
	class      string
	attributes map[string]string
	body       string
}

func parseSelector(selector string) (parsedSelector, error) {
	// attrs["id"] overwrites selector id, and is deleted from the map
	// any attrs in attrs overwrites selector attrs, and is deleted from the map
	// attrs["class"] is concatenated with any classes in selector class, and is deleted from the map
	return parsedSelector{}, nil
}

type HTMLElement struct {
	selector   string
	attributes map[string]string
	children   []Element
}

func H(selector string, attributes map[string]string, children ...Element) HTMLElement {
	return HTMLElement{selector: selector, attributes: attributes, children: children}
}

func (el HTMLElement) AppendHTML(buf *strings.Builder) error {
	s, err := parseSelector(el.selector)
	if err != nil {
		return erro.Wrap(err)
	}
	if s.body != "" {
		buf.WriteString(s.body)
		return nil
	}
	if s.tag != "" {
		buf.WriteString(`<` + s.tag)
	} else {
		buf.WriteString(`<div`)
	}
	if s.id != "" {
		buf.WriteString(` id="` + s.id + `"`)
	}
	if s.class != "" {
		buf.WriteString(` class="` + s.class + `"`)
	}
	buf.WriteString(`>`)
	for name, value := range s.attributes {
		buf.WriteString(` ` + name + `="` + value + `"`)
	}
	for _, child := range el.children {
		err = child.AppendHTML(buf)
		if err != nil {
			return erro.Wrap(err)
		}
	}
	if _, ok := singletonElements[strings.ToUpper(s.tag)]; !ok {
		buf.WriteString("</" + s.tag + ">")
	}
	return nil
}

type Mode int

const (
	ModeFormbuilder Mode = 0
	ModeUnmarshal   Mode = 1
)

type Form struct {
	Mode       Mode
	r          *http.Request
	selector   string
	attributes map[string]string
	elements   []Element
}

func (f *Form) Attrs(selector string, attributes map[string]string) {
	if f.Mode == ModeUnmarshal {
		return
	}
	f.selector = selector
	f.attributes = attributes
}

func (f *Form) H(selector string, attributes map[string]string, children ...Element) {
	if f.Mode == ModeUnmarshal {
		return
	}
}

type HTMLInputElement struct {
	selector     string
	name         string
	defaultValue string
	values       []string
}

func (el HTMLInputElement) AppendHTML(buf *strings.Builder) error {
	return nil
}

type StringInputElement struct {
	HTMLInputElement
}

type IntInputElement struct {
	HTMLInputElement
}

type Float64InputElement struct {
	HTMLInputElement
}

type FileInputElement struct {
	HTMLInputElement
}

func (f *Form) Text(name string, defaultValue string, selector string, attributes map[string]string, validators ...func(interface{}) error) StringInputElement {
	return StringInputElement{}
}
