package hyp

import (
	"encoding/json"
	"strings"

	"github.com/bokwoon95/erro"
)

type HTMLElement struct {
	attrs    Attributes
	children []Element
}

func (el HTMLElement) AppendHTML(buf *strings.Builder) error {
	err := AppendHTML(buf, el.attrs, el.children)
	if err != nil {
		return erro.Wrap(err)
	}
	return nil
}

func H(selector string, attributes map[string]string, children ...Element) HTMLElement {
	return HTMLElement{
		attrs:    ParseAttributes(selector, attributes),
		children: children,
	}
}

func (el *HTMLElement) ID() string {
	return el.attrs.ID
}

func (el *HTMLElement) Set(selector string, attributes map[string]string, children ...Element) {
	el.attrs = ParseAttributes(selector, attributes)
	el.children = children
}

func (el *HTMLElement) Append(selector string, attributes map[string]string, children ...Element) {
	el.children = append(el.children, H(selector, attributes, children...))
}

func (el *HTMLElement) AppendElements(elements ...Element) *HTMLElement {
	el.children = append(el.children, elements...)
	return el
}

type text string

func (t text) AppendHTML(buf *strings.Builder) error {
	buf.WriteString(string(t))
	return nil
}

func Txt(txt string) Element {
	return text(txt)
}

type List []Element

func (l List) AppendHTML(buf *strings.Builder) error {
	var err error
	for _, el := range l {
		err = el.AppendHTML(buf)
		if err != nil {
			return erro.Wrap(err)
		}
	}
	return nil
}

type JSONElement struct {
	attrs Attributes
	value interface{}
}

func (el JSONElement) AppendHTML(buf *strings.Builder) error {
	el.attrs.Tag = "script"
	el.attrs.Dict["type"] = "application/json"
	b, err := json.Marshal(el.value)
	if err != nil {
		return erro.Wrap(err)
	}
	err = AppendHTML(buf, el.attrs, []Element{text(string(b))})
	if err != nil {
		return erro.Wrap(err)
	}
	return nil
}

func JSON(selector string, attributes map[string]string, value interface{}) JSONElement {
	return JSONElement{
		attrs: ParseAttributes(selector, attributes),
		value: value,
	}
}
