package hg

import (
	"html/template"
	"net/http"
	"runtime"
	"strings"
	"sync"

	"github.com/bokwoon95/erro"
	"github.com/microcosm-cc/bluemonday"
)

// https://developer.mozilla.org/en-US/docs/Glossary/Empty_element
var singletonElements = map[string]struct{}{
	"AREA": {}, "BASE": {}, "BR": {}, "COL": {}, "EMBED": {}, "HR": {}, "IMG": {}, "INPUT": {},
	"LINK": {}, "META": {}, "PARAM": {}, "SOURCE": {}, "TRACK": {}, "WBR": {},
}

const Enabled = "\x00"

var sanitizer = func() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()
	p.AllowStyling()
	return p
}()

var bufpool = sync.Pool{
	New: func() interface{} { return &strings.Builder{} },
}

type Element interface {
	appendHTML(*strings.Builder) error
}

type Attr map[string]string

type Txt struct{ Text string }

func (t Txt) appendHTML(buf *strings.Builder) error {
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

func parseSelector(selector string, attributes map[string]string) (parsedSelector, error) {
	// attrs["id"] overwrites selector id, and is deleted from the map
	// any attrs in attrs overwrites selector attrs, and is deleted from the map
	// attrs["class"] is concatenated with any classes in selector class, and is deleted from the map
	return parsedSelector{}, nil
}

func appendHTML(buf *strings.Builder, selector string, attributes map[string]string, children []Element) error {
	s, err := parseSelector(selector, attributes)
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
	for _, child := range children {
		err = child.appendHTML(buf)
		if err != nil {
			return erro.Wrap(err)
		}
	}
	if _, ok := singletonElements[strings.ToUpper(s.tag)]; !ok {
		buf.WriteString("</" + s.tag + ">")
	}
	return nil
}

type HTMLElement struct {
	selector   string
	attributes map[string]string
	children   []Element
}

func H(selector string, attributes map[string]string, children ...Element) HTMLElement {
	return HTMLElement{selector: selector, attributes: attributes, children: children}
}

func (el HTMLElement) appendHTML(buf *strings.Builder) error {
	return appendHTML(buf, el.selector, el.attributes, el.children)
}

func (el *HTMLElement) Set(selector string, attributes map[string]string, children ...Element) {
	el.selector = selector
	el.attributes = attributes
	el.children = children
}

func (el *HTMLElement) Append(selector string, attributes map[string]string, children ...Element) {
	el.children = append(el.children, H(selector, attributes, children...))
}

func (el *HTMLElement) AppendElements(elements ...Element) {
	el.children = append(el.children, elements...)
}

func Marshal(w http.ResponseWriter, r *http.Request, el Element) (template.HTML, error) {
	buf := bufpool.Get().(*strings.Builder)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	err := el.appendHTML(buf)
	if err != nil {
		return "", erro.Wrap(err)
	}
	output := sanitizer.Sanitize(buf.String())
	return template.HTML(output), nil
}

func Redirect(w http.ResponseWriter, r *http.Request, url string, err error) {
	// cast err into a FormError, then serialize it into a cookie
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
