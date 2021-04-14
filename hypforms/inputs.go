package hypforms

import (
	"errors"
	"strconv"
	"strings"

	"github.com/bokwoon95/erro"
	"github.com/bokwoon95/pagemanager/hyp"
)

type Input struct {
	form         *Form
	attrs        hyp.Attributes
	inputType    string
	name         string
	defaultValue string
}

func (i *Input) AppendHTML(buf *strings.Builder) error {
	i.attrs.Tag = "input"
	i.attrs.Dict["type"] = i.inputType
	i.attrs.Dict["name"] = i.name
	i.attrs.Dict["value"] = i.defaultValue
	err := hyp.AppendHTML(buf, i.attrs, nil)
	if err != nil {
		return erro.Wrap(err)
	}
	return nil
}

func (i *Input) Type() string         { return i.inputType }
func (i *Input) Name() string         { return i.name }
func (i *Input) DefaultValue() string { return i.defaultValue }

func (i *Input) Set(selector string, attributes map[string]string) *Input {
	i.attrs = hyp.ParseAttributes(selector, attributes)
	return i
}

func (i *Input) Validate(validators ...Validator) *Input {
	var value interface{}
	if len(i.form.request.Form[i.name]) > 0 {
		value = i.form.request.Form[i.name][0]
	}
	validate(i.form, i.name, value, validators)
	return i
}

type wrapError struct {
	msg string
	err error
}

func (e *wrapError) Error() string {
	return e.msg
}

func (e *wrapError) Unwrap() error {
	return e.err
}

func ErrOneOf(errs []error, target error) bool {
	for _, err := range errs {
		if errors.Is(err, target) {
			return true
		}
	}
	return false
}

func (i *Input) Errors() []error {
	if i.form.mode != FormModeUnmarshal {
		return nil
	}
	return i.form.inputErrs[i.name]
}

func (i *Input) Value() (value string, exists bool) {
	if i.form.mode != FormModeUnmarshal {
		return "", false
	}
	if len(i.form.request.Form[i.name]) == 0 {
		return "", false
	}
	return i.form.request.Form[i.name][0], true
}

func (i *Input) Int(validators ...Validator) (num int, err error) {
	if i.form.mode != FormModeUnmarshal {
		return 0, nil
	}
	value := i.form.request.FormValue(i.name)
	num, err = strconv.Atoi(value)
	if err != nil {
		return 0, erro.Wrap(err)
	}
	validate(i.form, i.name, num, validators)
	return num, nil
}

func (i *Input) Float64(validators ...Validator) (num float64, err error) {
	if i.form.mode != FormModeUnmarshal {
		return 0, nil
	}
	value := i.form.request.FormValue(i.name)
	num, err = strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, erro.Wrap(err)
	}
	validate(i.form, i.name, num, validators)
	return num, nil
}

func (f *Form) Input(inputType string, name string, defaultValue string) *Input {
	f.registerName(name, 1)
	return &Input{form: f, inputType: inputType, name: name, defaultValue: defaultValue}
}

func (f *Form) Text(name string, defaultValue string) *Input {
	f.registerName(name, 1)
	return &Input{form: f, inputType: "text", name: name, defaultValue: defaultValue}
}

func (f *Form) Hidden(name string, defaultValue string) *Input {
	f.registerName(name, 1)
	return &Input{form: f, inputType: "hidden", name: name, defaultValue: defaultValue}
}

type ToggledInput struct {
	form      *Form
	attrs     hyp.Attributes
	inputType string
	name      string
	value     string
	checked   bool
}

func (i *ToggledInput) AppendHTML(buf *strings.Builder) error {
	i.attrs.Tag = "input"
	i.attrs.Dict["type"] = i.inputType
	i.attrs.Dict["name"] = i.name
	if i.value != "" {
		i.attrs.Dict["value"] = i.value
	}
	if i.checked {
		i.attrs.Dict["checked"] = hyp.Enabled
	} else {
		i.attrs.Dict["checked"] = hyp.Disabled
	}
	err := hyp.AppendHTML(buf, i.attrs, nil)
	if err != nil {
		return erro.Wrap(err)
	}
	return nil
}

func (f *Form) Checkbox(name string, value string, checked bool) *ToggledInput {
	f.registerName(name, 1)
	return &ToggledInput{form: f, inputType: "checkbox", name: name, value: value, checked: checked}
}

func (f *Form) Radio(name string, value string, checked bool) *ToggledInput {
	f.registerName(name, 1)
	return &ToggledInput{form: f, inputType: "radio", name: name, value: value, checked: checked}
}

func (i *ToggledInput) Name() string  { return i.name }
func (i *ToggledInput) Type() string  { return i.inputType }
func (i *ToggledInput) Value() string { return i.value }

func (i *ToggledInput) Set(selector string, attributes map[string]string) *ToggledInput {
	i.attrs = hyp.ParseAttributes(selector, attributes)
	return i
}

func (i *ToggledInput) Check(b bool) *ToggledInput {
	i.checked = b
	return i
}

func (i *ToggledInput) Checked() bool {
	if i.form.mode != FormModeUnmarshal {
		return false
	}
	values, ok := i.form.request.Form[i.name]
	if !ok || len(values) == 0 {
		return false
	}
	for _, v := range values {
		if i.value == "" && v == "on" {
			return true
		}
		if i.value != "" && v == i.value {
			return true
		}
	}
	return false
}

type ToggledInputs struct {
	form      *Form
	inputType string
	name      string
	options   []string
}

func (f *Form) Checkboxes(name string, options []string) *ToggledInputs {
	f.registerName(name, 1)
	return &ToggledInputs{form: f, inputType: "checkbox", name: name, options: options}
}

func (f *Form) Radios(name string, options []string) *ToggledInputs {
	f.registerName(name, 1)
	return &ToggledInputs{form: f, inputType: "radio", name: name, options: options}
}

func (i *ToggledInputs) Inputs() []*ToggledInput {
	var inputs []*ToggledInput
	for _, option := range i.options {
		inputs = append(inputs, &ToggledInput{form: i.form, inputType: i.inputType, name: i.name, value: option})
	}
	return inputs
}

func (i *ToggledInputs) Options() []string {
	return i.options
}

func (i *ToggledInputs) Value() string {
	if i.form.mode != FormModeUnmarshal {
		return ""
	}
	return i.form.request.FormValue(i.name)
}

func (i *ToggledInputs) Values() []string {
	if i.form.mode != FormModeUnmarshal {
		return nil
	}
	return i.form.request.Form[i.name]
}

type Opt interface {
	hyp.Element
	Opt()
}
type Opts []Opt

type Option struct {
	Value      string
	Display    string
	Disabled   bool
	Selected   bool
	Selector   string
	Attributes map[string]string
}

func (o Option) Opt() {}

func (o Option) AppendHTML(buf *strings.Builder) error {
	attrs := hyp.ParseAttributes(o.Selector, o.Attributes)
	attrs.Tag = "option"
	attrs.Dict["value"] = o.Value
	if o.Disabled {
		attrs.Dict["disabled"] = hyp.Enabled
	}
	if o.Selected {
		attrs.Dict["selected"] = hyp.Enabled
	}
	err := hyp.AppendHTML(buf, attrs, []hyp.Element{hyp.Txt(o.Display)})
	if err != nil {
		return erro.Wrap(err)
	}
	return nil
}

type OptGroup struct {
	Label      string
	Disabled   bool
	Options    []Option
	Selector   string
	Attributes map[string]string
}

func (o OptGroup) Opt() {}

func (o OptGroup) AppendHTML(buf *strings.Builder) error {
	attrs := hyp.ParseAttributes(o.Selector, o.Attributes)
	attrs.Tag = "option"
	attrs.Dict["label"] = o.Label
	if o.Disabled {
		attrs.Dict["disabled"] = hyp.Enabled
	}
	elements := make([]hyp.Element, len(o.Options))
	for i := range o.Options {
		elements[i] = o.Options[i]
	}
	err := hyp.AppendHTML(buf, attrs, elements)
	if err != nil {
		return erro.Wrap(err)
	}
	return nil
}

type SelectInput struct {
	form  *Form
	attrs hyp.Attributes
	name  string
	Opts  Opts
}

func (i *SelectInput) AppendHTML(buf *strings.Builder) error {
	elements := make([]hyp.Element, len(i.Opts))
	for j := range i.Opts {
		elements[j] = i.Opts[j]
	}
	err := hyp.AppendHTML(buf, i.attrs, elements)
	if err != nil {
		return erro.Wrap(err)
	}
	return nil
}

func (i *SelectInput) Options() []Option {
	var options []Option
	for _, opt := range i.Opts {
		switch opt := opt.(type) {
		case Option:
			options = append(options, opt)
		case OptGroup:
			options = append(options, opt.Options...)
		}
	}
	return options
}

func (i *SelectInput) Values() []string {
	if i.form.mode != FormModeUnmarshal {
		return nil
	}
	return i.form.request.Form[i.name]
}
