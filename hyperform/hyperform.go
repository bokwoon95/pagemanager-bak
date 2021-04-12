package hyperform

import (
	"fmt"
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
const Disabled = "\x01"

var sanitizer = func() *bluemonday.Policy {
	p := bluemonday.UGCPolicy()
	p.AllowStyling()
	p.AllowElements("form", "input", "button", "label", "select", "option")
	p.AllowAttrs("name", "value", "type", "for").Globally()
	p.AllowAttrs("method", "action").OnElements("form")
	return p
}()

var bufpool = sync.Pool{
	New: func() interface{} { return &strings.Builder{} },
}

type Element interface {
	appendHTML(*strings.Builder) error
}

type Attr map[string]string

type text string

func (t text) appendHTML(buf *strings.Builder) error {
	buf.WriteString(string(t))
	return nil
}

func Txt(txt string) Element {
	return text(txt)
}

type parsedSelector struct {
	tag        string
	id         string
	class      string
	attributes map[string]string
	body       string
}

func parseSelector(selector string, attributes map[string]string) (parsedSelector, error) {
	const (
		StateEmpty = iota
		StateTag
		StateID
		StateClass
		StateAttrName
		StateAttrValue
	)
	s := parsedSelector{attributes: make(map[string]string)}
	if strings.HasPrefix(selector, "<") && strings.HasSuffix(selector, ">") {
		s.body = selector
		return s, nil
	}
	state := StateTag
	var classes []string
	var name []rune
	var value []rune
	for _, c := range selector {
		if c == '#' || c == '.' || c == '[' {
			switch state {
			case StateTag:
				s.tag = string(value)
			case StateID:
				s.id = string(value)
			case StateClass:
				if len(value) > 0 {
					classes = append(classes, string(value))
				}
			case StateAttrName, StateAttrValue:
				return s, erro.Wrap(fmt.Errorf("unclosed attribute"))
			}
			value = value[:0]
			switch c {
			case '#':
				state = StateID
			case '.':
				state = StateClass
			case '[':
				state = StateAttrName
			}
			continue
		}
		if c == '=' {
			switch state {
			case StateAttrName:
				state = StateAttrValue
			default:
				return s, erro.Wrap(fmt.Errorf("unopened attribute"))
			}
			continue
		}
		if c == ']' {
			switch state {
			case StateAttrName:
				s.attributes[string(name)] = Enabled
			case StateAttrValue:
				s.attributes[string(name)] = string(value)
			default:
				return s, erro.Wrap(fmt.Errorf("unopened attribute"))
			}
			name = name[:0]
			value = value[:0]
			state = StateEmpty
			continue
		}
		switch state {
		case StateTag, StateID, StateClass, StateAttrValue:
			value = append(value, c)
		case StateAttrName:
			name = append(name, c)
		case StateEmpty:
			return s, erro.Wrap(fmt.Errorf("unknown state (please prepend with '#', '.' or '['"))
		}
	}
	// flush value
	if len(value) > 0 {
		switch state {
		case StateTag:
			s.tag = string(value)
		case StateID:
			s.id = string(value)
		case StateClass:
			classes = append(classes, string(value))
		case StateEmpty:
			// do nothing, drop the value
		case StateAttrName, StateAttrValue:
			return s, erro.Wrap(fmt.Errorf("unclosed attribute"))
		}
		value = value[:0]
	}
	if s.tag == "" {
		s.tag = "div"
	}
	if len(classes) > 0 {
		s.class = strings.Join(classes, " ")
	}
	for name, value := range attributes {
		switch name {
		case "id":
			s.id = value
		case "class":
			if value != "" {
				s.class += " " + value
			}
		default:
			s.attributes[name] = value
		}
	}
	return s, nil
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
	for name, value := range s.attributes {
		buf.WriteString(` ` + name + `="` + value + `"`)
	}
	buf.WriteString(`>`)
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

func appendHTMLv2(buf *strings.Builder, s parsedSelector, children []Element) error {
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
	for name, value := range s.attributes {
		buf.WriteString(` ` + name + `="` + value + `"`)
	}
	buf.WriteString(`>`)
	var err error
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
	output := buf.String()
	output = sanitizer.Sanitize(buf.String())
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
