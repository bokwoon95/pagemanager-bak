package hyperform

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bokwoon95/erro"
)

type input struct {
	form         *Form
	inputType    string
	name         string
	defaultValue string
	selector     string
	attributes   map[string]string
}

func (i input) appendHTML(buf *strings.Builder) error {
	s, err := parseSelector(i.selector, i.attributes)
	if err != nil {
		return erro.Wrap(err)
	}
	buf.WriteString(`<input`)
	if s.id != "" {
		buf.WriteString(` id="` + s.id + `"`)
	}
	buf.WriteString(` type="` + i.inputType + `"`)
	buf.WriteString(` name="` + i.name + `"`)
	if i.defaultValue != "" {
		buf.WriteString(` value="` + i.defaultValue + `"`)
	}
	if s.class != "" {
		buf.WriteString(` class="` + s.class + `"`)
	}
	for name, value := range s.attributes {
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
	return nil
}

func (i *input) validate(value interface{}, validators []func(interface{}) error) {
	if len(validators) == 0 {
		return
	}
	var err error
	for _, validator := range validators {
		err = validator(value)
		if err != nil {
			i.form.errors.inputErrors[i.name] = append(i.form.errors.inputErrors[i.name], err)
		}
	}
}

func (i *input) Name() string {
	return i.name
}

func (i *input) DefaultValue() string {
	return i.defaultValue
}

func (i *input) Value() string {
	if i.form.mode != FormModeUnmarshal {
		return ""
	}
	value := i.form.request.FormValue(i.name)
	return value
}

type Input struct{ input }

func (f *Form) Input(inputType string, name string, defaultValue string) *Input {
	if _, ok := f.names[name]; ok {
		file, line, _ := caller(1)
		f.errors.formErrors = append(f.errors.formErrors, fmt.Errorf("%s:%d duplicate name: %s", file, line, name))
	}
	f.names[name] = struct{}{}
	return &Input{input: input{
		inputType:    inputType,
		form:         f,
		name:         name,
		defaultValue: defaultValue,
	}}
}

func (i *Input) Set(selector string, attributes map[string]string) *Input {
	i.selector = selector
	i.attributes = attributes
	return i
}

func (i *Input) Errors() []error {
	if i.form.mode != FormModeUnmarshal {
		return nil
	}
	return i.form.errors.inputErrors[i.name]
}

func (i *Input) Validate(validators ...func(interface{}) error) *Input {
	if i.form.mode != FormModeUnmarshal {
		return i
	}
	value := i.form.request.FormValue(i.name)
	i.validate(value, validators)
	return i
}

type StringInput struct{ input }

func (f *Form) Text(name string, defaultValue string) *StringInput {
	if _, ok := f.names[name]; ok {
		file, line, _ := caller(1)
		f.errors.formErrors = append(f.errors.formErrors, fmt.Errorf("%s:%d duplicate name: %s", file, line, name))
	}
	f.names[name] = struct{}{}
	return &StringInput{input: input{
		inputType:    "text",
		form:         f,
		name:         name,
		defaultValue: defaultValue,
	}}
}

func (i *StringInput) Set(selector string, attributes map[string]string) *StringInput {
	i.selector = selector
	i.attributes = attributes
	return i
}

func (i *StringInput) Errors() []error {
	if i.form.mode != FormModeUnmarshal {
		return nil
	}
	return i.form.errors.inputErrors[i.name]
}

func (i *StringInput) Validate(validators ...func(interface{}) error) *StringInput {
	if i.form.mode != FormModeUnmarshal {
		return i
	}
	value := i.form.request.FormValue(i.name)
	i.validate(value, validators)
	return i
}

type NumberInput struct{ input }

func (f *Form) Number(name string, defaultValue float64) *NumberInput {
	if _, ok := f.names[name]; ok {
		file, line, _ := caller(1)
		f.errors.formErrors = append(f.errors.formErrors, fmt.Errorf("%s:%d duplicate name: %s", file, line, name))
	}
	f.names[name] = struct{}{}
	return &NumberInput{input: input{
		inputType:    "number",
		form:         f,
		name:         name,
		defaultValue: strconv.FormatFloat(defaultValue, 'f', -1, 64),
	}}
}

func (i *NumberInput) Set(selector string, attributes map[string]string) *NumberInput {
	i.selector = selector
	i.attributes = attributes
	return i
}

func (i *NumberInput) Errors() []error {
	if i.form.mode != FormModeUnmarshal {
		return nil
	}
	return i.form.errors.inputErrors[i.name]
}

func (i *NumberInput) Validate(validators ...func(interface{}) error) *NumberInput {
	if i.form.mode != FormModeUnmarshal {
		return i
	}
	value := i.form.request.FormValue(i.name)
	i.validate(value, validators)
	return i
}

func (i *NumberInput) Int(validators ...func(interface{}) error) int {
	if i.form.mode != FormModeUnmarshal {
		return 0
	}
	value := i.form.request.FormValue(i.name)
	num, err := strconv.Atoi(value)
	if err != nil {
		i.form.errors.inputErrors[i.name] = append(i.form.errors.inputErrors[i.name], fmt.Errorf("not a number: %s", value))
	} else {
		i.validate(num, validators)
	}
	return num
}

func (i *NumberInput) Float64(validators ...func(interface{}) error) float64 {
	if i.form.mode != FormModeUnmarshal {
		return 0
	}
	value := i.form.request.FormValue(i.name)
	num, err := strconv.ParseFloat(value, 64)
	if err != nil {
		i.form.errors.inputErrors[i.name] = append(i.form.errors.inputErrors[i.name], fmt.Errorf("not a number: %s", value))
	} else {
		i.validate(num, validators)
	}
	return num
}

type HiddenInput struct{ input }

func (f *Form) Hidden(name string, defaultValue string) *HiddenInput {
	if _, ok := f.names[name]; ok {
		file, line, _ := caller(1)
		f.errors.formErrors = append(f.errors.formErrors, fmt.Errorf("%s:%d duplicate name: %s", file, line, name))
	}
	f.names[name] = struct{}{}
	return &HiddenInput{input: input{
		inputType:    "hidden",
		form:         f,
		name:         name,
		defaultValue: defaultValue,
	}}
}

func (i *HiddenInput) Set(selector string, attributes map[string]string) *HiddenInput {
	i.selector = selector
	i.attributes = attributes
	return i
}

type CheckboxInput struct{ input }

func (f *Form) Checkbox(name string, defaultValue string) *CheckboxInput {
	return &CheckboxInput{input: input{
		inputType:    "checkbox",
		form:         f,
		name:         name,
		defaultValue: defaultValue,
	}}
}

func (i *CheckboxInput) Set(selector string, attributes map[string]string) *CheckboxInput {
	i.selector = selector
	i.attributes = attributes
	return i
}

func (i *CheckboxInput) Checked() bool {
	if i.form.mode != FormModeUnmarshal {
		return false
	}
	values, ok := i.form.request.Form[i.name]
	if !ok || len(values) == 0 {
		return false
	}
	for _, value := range values {
		if i.defaultValue == "" && value == "on" {
			return true
		}
		if i.defaultValue != "" && value == i.defaultValue {
			return true
		}
	}
	return false
}

type CheckboxInputs struct {
	form    *Form
	name    string
	options []string
}

func (f *Form) Checkboxes(name string, options []string) *CheckboxInputs {
	if _, ok := f.names[name]; ok {
		file, line, _ := caller(1)
		f.errors.formErrors = append(f.errors.formErrors, fmt.Errorf("%s:%d duplicate name: %s", file, line, name))
	}
	f.names[name] = struct{}{}
	return &CheckboxInputs{
		form:    f,
		name:    name,
		options: options,
	}
}

func (i *CheckboxInputs) Checkboxes() []*CheckboxInput {
	var inputs []*CheckboxInput
	for _, option := range i.options {
		inputs = append(inputs, &CheckboxInput{input: input{
			inputType:    "checkbox",
			form:         i.form,
			name:         i.name,
			defaultValue: option,
		}})
	}
	return inputs
}

func (i *CheckboxInputs) Options() []string {
	return i.options
}

func (i *CheckboxInputs) Values() []string {
	if i.form.mode != FormModeUnmarshal {
		return nil
	}
	return i.form.request.Form[i.name]
}

type RadioInput struct{ input }

func (i *RadioInput) Set(selector string, attributes map[string]string) *RadioInput {
	i.selector = selector
	i.attributes = attributes
	return i
}

type RadioInputs struct {
	form    *Form
	name    string
	options []string
}

func (f *Form) Radios(name string, options []string) *RadioInputs {
	if _, ok := f.names[name]; ok {
		file, line, _ := caller(1)
		f.errors.formErrors = append(f.errors.formErrors, fmt.Errorf("%s:%d duplicate name: %s", file, line, name))
	}
	f.names[name] = struct{}{}
	return &RadioInputs{
		form:    f,
		name:    name,
		options: options,
	}
}

func (i *RadioInputs) Radios() []RadioInput {
	var inputs []RadioInput
	for _, option := range i.options {
		inputs = append(inputs, RadioInput{input: input{
			inputType:    "checkbox",
			form:         i.form,
			name:         i.name,
			defaultValue: option,
		}})
	}
	return inputs
}

func (i *RadioInputs) Options() []string {
	return i.options
}

func (i *RadioInputs) Value() string {
	if i.form.mode != FormModeUnmarshal {
		return ""
	}
	return i.form.request.FormValue(i.name)
}

type SelectOption struct {
	Value      string
	Display    string
	Selected   bool
	Selector   string
	Attributes map[string]string
}

type SelectInput struct {
	form       *Form
	name       string
	Options    []*SelectOption
	selector   string
	attributes map[string]string
}

func (i SelectInput) appendHTML(buf *strings.Builder) error {
	s, err := parseSelector(i.selector, i.attributes)
	if err != nil {
		return erro.Wrap(err)
	}
	buf.WriteString(`<select`)
	if s.id != "" {
		buf.WriteString(` id="` + s.id + `"`)
	}
	buf.WriteString(` name="` + i.name + `"`)
	if s.class != "" {
		buf.WriteString(` class="` + s.class + `"`)
	}
	for name, value := range s.attributes {
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
	for _, option := range i.Options {
		buf.WriteString(`<option`)
		s, err := parseSelector(option.Selector, option.Attributes)
		if err != nil {
			return erro.Wrap(err)
		}
		if s.id != "" {
			buf.WriteString(` id="` + s.id + `"`)
		}
		buf.WriteString(` value="` + option.Value + `"`)
		if s.class != "" {
			buf.WriteString(` class="` + s.class + `"`)
		}
		for name, value := range s.attributes {
			if name == "selected" {
				continue
			}
			switch value {
			case Enabled:
				buf.WriteString(` ` + name)
			case Disabled:
				continue
			default:
				buf.WriteString(` ` + name + `="` + value + `"`)
			}
		}
		if option.Selected {
			buf.WriteString(` selected`)
		}
		buf.WriteString(`>`)
		buf.WriteString(option.Display + `</option>`)
	}
	buf.WriteString(`</select>`)
	return nil
}

func (i SelectInput) Name() string {
	return i.name
}

func (i SelectInput) Value() string {
	if i.form.mode != FormModeUnmarshal {
		return ""
	}
	value := i.form.request.FormValue(i.name)
	return value
}

func (i SelectInput) Values() []string {
	if i.form.mode != FormModeUnmarshal {
		return nil
	}
	values := i.form.request.Form[i.name]
	return values
}

func (f *Form) Select(name string, options []SelectOption) *SelectInput {
	if _, ok := f.names[name]; ok {
		file, line, _ := caller(1)
		f.errors.formErrors = append(f.errors.formErrors, fmt.Errorf("%s:%d duplicate name: %s", file, line, name))
	}
	f.names[name] = struct{}{}
	i := &SelectInput{form: f, name: name}
	for _, option := range options {
		i.Options = append(i.Options, &option)
	}
	return i
}

func (i *SelectInput) Set(selector string, attributes map[string]string) *SelectInput {
	i.selector = selector
	i.attributes = attributes
	return i
}

func (i *SelectInput) Errors() []error {
	if i.form.mode != FormModeUnmarshal {
		return nil
	}
	return i.form.errors.inputErrors[i.name]
}

// InputString: String() Strings()
// InputNumber: String() Strings() Int() Ints() Float64() Float64s()
// InputHidden:
// InputCheckbox:
// InputRadio:

// Input: String(), Strings()
// Text:
// Hidden:
// Number: Int(), Ints()
// Checkbox: Bool()
// File: ???

// radio, checkbox, select, date, range
