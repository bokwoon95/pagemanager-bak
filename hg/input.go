package hg

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
		buf.WriteString(` ` + name + `="` + value + `"`)
	}
	buf.WriteString(`>`)
	return nil
}

func (i *input) validate(value interface{}, validators []Validator) {
	if len(validators) == 0 {
		return
	}
	validatorfunc := func(interface{}, *[]error) {}
	for i := len(validators) - 1; i >= 0; i-- {
		validatorfunc = validators[i](validatorfunc)
	}
	var errs []error
	validatorfunc(value, &errs)
	i.form.errors.inputErrors[i.name] = append(i.form.errors.inputErrors[i.name], errs...)
}

func (i *input) Name() string {
	return i.name
}

func (i *input) DefaultValue() string {
	return i.defaultValue
}

func (i *input) Value(validators ...Validator) string {
	if i.form.mode != FormModeUnmarshal {
		return ""
	}
	value := i.form.request.FormValue(i.name)
	i.validate(value, validators)
	return value
}

type Input struct{ input }

func (f *Form) Input(inputType string, name string, defaultValue string) *Input {
	if _, ok := f.names[name]; ok {
		file, line, _ := caller(1)
		f.errors.customErrors = append(f.errors.customErrors, fmt.Errorf("%s:%d duplicate name: %s", file, line, name))
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

type StringInput struct{ input }

func (f *Form) Text(name string, defaultValue string) *StringInput {
	if _, ok := f.names[name]; ok {
		file, line, _ := caller(1)
		f.errors.customErrors = append(f.errors.customErrors, fmt.Errorf("%s:%d duplicate name: %s", file, line, name))
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

type NumberInput struct{ input }

func (f *Form) Number(name string, defaultValue float64) *NumberInput {
	if _, ok := f.names[name]; ok {
		file, line, _ := caller(1)
		f.errors.customErrors = append(f.errors.customErrors, fmt.Errorf("%s:%d duplicate name: %s", file, line, name))
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

// type validator func(next validator, errs *[]error)

// TODO: if the field is not a number, is it okay to skip the validators?
// if number field is optional, it is good that the validators get skipped
// if the number field is required, you mark it with a hg.NonEmpty but it would never get called
// age.IntValue(hg.IsNumber, hg.NonEmpty)
func (i *NumberInput) IntValue(validators ...Validator) int {
	if i.form.mode != FormModeUnmarshal {
		return 0
	}
	value := i.form.request.FormValue(i.name)
	validators = append([]Validator{CastToNumber}, validators...)
	i.validate(value, validators)
	num, _ := strconv.Atoi(value)
	return num
}

func (i *NumberInput) Float64Value(validators ...Validator) float64 {
	if i.form.mode != FormModeUnmarshal {
		return 0
	}
	s := i.form.request.FormValue(i.name)
	num, err := strconv.ParseFloat(s, 64)
	if err != nil {
		i.form.errors.inputErrors[i.name] = append(i.form.errors.inputErrors[i.name], err)
	} else {
		i.validate(num, validators)
	}
	return num
}

type HiddenInput struct{ input }

func (f *Form) Hidden(name string, defaultValue string) *HiddenInput {
	if _, ok := f.names[name]; ok {
		file, line, _ := caller(1)
		f.errors.customErrors = append(f.errors.customErrors, fmt.Errorf("%s:%d duplicate name: %s", file, line, name))
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

// CheckboxInput should just be an Element impl that contains Set() (but not Append())
type CheckboxInput struct{ input }

// CheckboxInputs must function like a single input
type CheckboxInputs struct {
	form           *Form
	name           string
	values         []string
	set            map[string]struct{}
	setInitialized bool
}

func (f *Form) Checkboxes(name string, values []string) *CheckboxInputs {
	if _, ok := f.names[name]; ok {
		file, line, _ := caller(1)
		f.errors.customErrors = append(f.errors.customErrors, fmt.Errorf("%s:%d duplicate name: %s", file, line, name))
	}
	f.names[name] = struct{}{}
	return &CheckboxInputs{
		form:   f,
		name:   name,
		values: values,
		set:    make(map[string]struct{}),
	}
}

func (i *CheckboxInputs) Checkboxes() []CheckboxInput {
	var inputs []CheckboxInput
	for _, value := range i.values {
		inputs = append(inputs, CheckboxInput{input: input{
			inputType:    "checkbox",
			form:         i.form,
			name:         i.name,
			defaultValue: value,
		}})
	}
	return inputs
}

// No need for this, you just return the list of values and it's up to the user to manually save it into a map for O(1) lookup
func (i *CheckboxInputs) IsSet(value string) bool {
	if !i.setInitialized {
		values := i.form.request.Form[i.name]
		for _, value := range values {
			i.set[value] = struct{}{}
		}
		i.setInitialized = true
	}
	_, ok := i.set[value]
	return ok
}

type RadioInput struct{ input }

type RadioInputs struct {
	form           *Form
	name           string
	values         []string
	set            map[string]struct{}
	setInitialized bool
}

func (f *Form) Radios(name string, values []string) *RadioInputs {
	return &RadioInputs{form: f, name: name, values: values, set: make(map[string]struct{})}
}

func (i *RadioInputs) Radios() []RadioInput {
	var inputs []RadioInput
	for _, value := range i.values {
		inputs = append(inputs, RadioInput{input: input{
			inputType:    "checkbox",
			form:         i.form,
			name:         i.name,
			defaultValue: value,
		}})
	}
	return inputs
}

func (i *RadioInputs) IsSet(value string) bool {
	if !i.setInitialized {
		values := i.form.request.Form[i.name]
		for _, value := range values {
			i.set[value] = struct{}{}
		}
		i.setInitialized = true
	}
	_, ok := i.set[value]
	return ok
}

type SelectOption struct {
	Value   string
	Display string
}

type SelectInput struct {
	form         *Form
	name         string
	defaultValue string
	selector     string
	attributes   map[string]string
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
