package hyp

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

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

type textValue struct {
	v interface{}
}

func (txt textValue) AppendHTML(buf *strings.Builder) error {
	switch v := txt.v.(type) {
	case fmt.Stringer:
		buf.WriteString(v.String())
	case string:
		buf.WriteString(v)
	case []byte:
		buf.WriteString(string(v))
	case time.Time:
		buf.WriteString(v.Format(time.RFC3339Nano))
	case error:
		buf.WriteString(v.Error())
	default:
		rv := reflect.ValueOf(txt.v)
		switch rv.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			buf.WriteString(strconv.FormatInt(rv.Int(), 10))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			buf.WriteString(strconv.FormatUint(rv.Uint(), 10))
		case reflect.Float64:
			buf.WriteString(strconv.FormatFloat(rv.Float(), 'g', -1, 64))
		case reflect.Float32:
			buf.WriteString(strconv.FormatFloat(rv.Float(), 'g', -1, 32))
		case reflect.Bool:
			buf.WriteString(strconv.FormatBool(rv.Bool()))
		default:
			buf.WriteString(fmt.Sprintf("%v", txt.v))
		}
	}
	return nil
}

func Txt(v interface{}) Element {
	return textValue{v: v}
}

type Elements []Element

func (l Elements) AppendHTML(buf *strings.Builder) error {
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
