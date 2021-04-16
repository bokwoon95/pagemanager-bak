package hy

import (
	"fmt"
	"html/template"
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

var bufpool = sync.Pool{
	New: func() interface{} { return &strings.Builder{} },
}

const Enabled = "\x00"
const Disabled = "\x01"

type Element interface {
	AppendHTML(*strings.Builder) error
}

type Attr map[string]string

type Sanitizer interface {
	Sanitize(string) string
}

var defaultSanitizer = func() Sanitizer {
	p := bluemonday.UGCPolicy()
	p.AllowStyling()
	p.AllowDataAttributes()
	// https://developer.mozilla.org/en-US/docs/Web/HTML/Element/form#attributes
	p.AllowElements("form")
	p.AllowAttrs("accept-charset", "autocomplete", "name", "rel", "action", "enctype", "method", "novalidate", "target").OnElements("form")
	// https://developer.mozilla.org/en-US/docs/Web/HTML/Element/input#attributes
	p.AllowElements("input")
	p.AllowAttrs(
		"accept", "alt", "autocomplete", "autofocus", "capture", "checked",
		"dirname", "disabled", "form", "formaction", "formenctype", "formmethod",
		"formnovalidate", "formtarget", "height", "list", "max", "maxlength", "min",
		"minlength", "multiple", "name", "pattern", "placeholder", "readonly",
		"required", "size", "src", "step", "type", "value", "width",
	).OnElements("input")
	// https://developer.mozilla.org/en-US/docs/Web/HTML/Element/button#attributes
	p.AllowElements("button")
	p.AllowAttrs(
		"autofocus", "disabled", "form", "formaction", "formenctype",
		"formmethod", "formnovalidate", "formtarget", "name", "type", "value",
	).OnElements("button")
	// https://developer.mozilla.org/en-US/docs/Web/HTML/Element/label#attributes
	p.AllowElements("label")
	p.AllowAttrs("for").OnElements("label")
	// https://developer.mozilla.org/en-US/docs/Web/HTML/Element/select#attributes
	p.AllowElements("select")
	p.AllowAttrs("autocomplete", "autofocus", "disabled", "form", "multiple", "name", "required", "size").OnElements("select")
	// https://developer.mozilla.org/en-US/docs/Web/HTML/Element/option#attributes
	p.AllowElements("option")
	p.AllowAttrs("disabled", "label", "selected", "value").OnElements("option")
	// https://developer.mozilla.org/en-US/docs/Web/HTML/Global_attributes/inputmode
	p.AllowAttrs("inputmode").Globally()
	// https://developer.mozilla.org/en-US/docs/Web/HTML/Element/link#attributes
	p.AllowElements("link")
	p.AllowAttrs(
		"as", "crossorigin", "disabled", "href", "hreflang", "imagesizes",
		"imagesrcset", "media", "rel", "sizes", "title", "type",
	).OnElements("link")
	p.AllowStandardURLs()
	// https://developer.mozilla.org/en-US/docs/Web/HTML/Element/script#attributes
	p.AllowElements("script")
	p.AllowAttrs("async", "crossorigin", "defer", "integrity", "nomodule", "nonce", "referrerpolicy", "src", "type").OnElements("script")

	p.AllowElements("svg")

	p.AllowImages()
	p.AllowLists()
	p.AllowTables()

	// settings which bluemonday loves to turn back on, leave it for the last
	p.RequireNoFollowOnLinks(false)
	return p
}()

type Attributes struct {
	ParseErr error
	Selector string
	Tag      string
	ID       string
	Class    string
	Dict     map[string]string
}

func ParseAttributes(selector string, attributes map[string]string) Attributes {
	type State int
	const (
		StateNone State = iota
		StateTag
		StateID
		StateClass
		StateAttrName
		StateAttrValue
	)
	attrs := Attributes{Selector: selector, Dict: make(map[string]string)}
	defer func() {
		if attrs.ParseErr != nil {
			for k, v := range attributes {
				attrs.Dict[k] = v
			}
		}
	}()
	state := StateTag
	var classes []string
	var name []rune
	var value []rune
	for _, c := range selector {
		if c == '#' || c == '.' || c == '[' {
			switch state {
			case StateTag:
				attrs.Tag = string(value)
			case StateID:
				attrs.ID = string(value)
			case StateClass:
				if len(value) > 0 {
					classes = append(classes, string(value))
				}
			case StateAttrName, StateAttrValue:
				attrs.ParseErr = erro.Wrap(fmt.Errorf("unclosed attribute"))
				return attrs
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
				attrs.ParseErr = erro.Wrap(fmt.Errorf("unopened attribute"))
				return attrs
			}
			continue
		}
		if c == ']' {
			switch state {
			case StateAttrName:
				if _, ok := attrs.Dict[string(name)]; ok {
					break
				}
				attrs.Dict[string(name)] = Enabled
			case StateAttrValue:
				if _, ok := attrs.Dict[string(name)]; ok {
					break
				}
				attrs.Dict[string(name)] = string(value)
			default:
				attrs.ParseErr = erro.Wrap(fmt.Errorf("unopened attribute"))
				return attrs
			}
			name = name[:0]
			value = value[:0]
			state = StateNone
			continue
		}
		switch state {
		case StateTag, StateID, StateClass, StateAttrValue:
			value = append(value, c)
		case StateAttrName:
			name = append(name, c)
		case StateNone:
			attrs.ParseErr = erro.Wrap(fmt.Errorf("unknown state (please prepend with '#', '.' or '['"))
			return attrs
		}
	}
	// flush value
	if len(value) > 0 {
		switch state {
		case StateTag:
			attrs.Tag = string(value)
		case StateID:
			attrs.ID = string(value)
		case StateClass:
			classes = append(classes, string(value))
		case StateNone: // do nothing i.e. drop the value
		case StateAttrName, StateAttrValue:
			attrs.ParseErr = erro.Wrap(fmt.Errorf("unclosed attribute"))
			return attrs
		}
		value = value[:0]
	}
	if len(classes) > 0 {
		attrs.Class = strings.Join(classes, " ")
	}
	for name, value := range attributes {
		switch name {
		case "id":
			attrs.ID = value
		case "class":
			if value != "" {
				attrs.Class += " " + value
			}
		default:
			attrs.Dict[name] = value
		}
	}
	return attrs
}

func AppendHTML(buf *strings.Builder, attrs Attributes, children []Element) error {
	var err error
	if attrs.ParseErr != nil {
		return erro.Wrap(attrs.ParseErr)
	}
	if attrs.Tag != "" {
		buf.WriteString(`<` + attrs.Tag)
	} else {
		buf.WriteString(`<div`)
	}
	if attrs.ID != "" {
		buf.WriteString(` id="` + attrs.ID + `"`)
	}
	if attrs.Class != "" {
		buf.WriteString(` class="` + attrs.Class + `"`)
	}
	for name, value := range attrs.Dict {
		switch value {
		case Enabled:
			buf.WriteString(` ` + name)
		case Disabled:
			continue
		default:
			buf.WriteString(` ` + name + `="` + value + `"`)
		}
	}
	buf.WriteString(`>`)
	if _, ok := singletonElements[strings.ToUpper(attrs.Tag)]; !ok {
		for _, child := range children {
			err = child.AppendHTML(buf)
			if err != nil {
				return erro.Wrap(err)
			}
		}
		buf.WriteString("</" + attrs.Tag + ">")
	}
	return nil
}

func MarshalElement(s Sanitizer, el Element) (template.HTML, error) {
	buf := bufpool.Get().(*strings.Builder)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	err := el.AppendHTML(buf)
	if err != nil {
		return "", erro.Wrap(err)
	}
	if s == nil {
		s = defaultSanitizer
	}
	output := s.Sanitize(buf.String())
	return template.HTML(output), nil
}

type ElementMap map[*template.HTML]Element

func MarshalElements(s Sanitizer, elements map[*template.HTML]Element) error {
	var err error
	for dest, element := range elements {
		if dest == nil {
			continue
		}
		*dest, err = MarshalElement(s, element)
		if err != nil {
			return erro.Wrap(err)
		}
	}
	return nil
}
