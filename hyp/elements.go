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

// adapted from database/sql:asString, text/template:printableValue,printValue
func Stringify(v interface{}) string {
	switch v := v.(type) {
	case fmt.Stringer:
		return v.String()
	case string:
		return v
	case []byte:
		return string(v)
	case time.Time:
		return v.Format(time.RFC3339Nano)
	case error:
		return v.Error()
	case int:
		return strconv.FormatInt(int64(v), 10)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'g', -1, 64)
	case float64:
		return strconv.FormatFloat(v, 'g', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	}
	rv := reflect.ValueOf(v)
	for {
		if rv.Kind() != reflect.Ptr && rv.Kind() != reflect.Interface {
			break
		}
		rv = rv.Elem()
	}
	if !rv.IsValid() {
		return "<no value>"
	}
	if rv.Kind() == reflect.Chan {
		return "<channel>"
	}
	if rv.Kind() == reflect.Func {
		return "<function>"
	}
	return fmt.Sprint(v)
}

func (txt textValue) AppendHTML(buf *strings.Builder) error {
	buf.WriteString(Stringify(txt.v))
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
	err = AppendHTML(buf, el.attrs, []Element{textValue{v: string(b)}})
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
