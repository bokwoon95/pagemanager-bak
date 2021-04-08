package hg

import "strconv"

type Input struct {
	form         *Form
	inputType    string
	selector     string
	attributes   map[string]string
	name         string
	defaultValue string
	errs         []error
}

func (i Input) Name() string {
	return i.name
}

func (i Input) Errors() []error {
	return i.errs
}

func (i Input) String(validators ...func(interface{}) error) string {
	return i.form.request.FormValue(i.name)
}

func (i Input) Strings() []string {
	return i.form.request.Form[i.name]
}

func (f *Form) Input(inputType string, name string, defaultValue string, selector string, attributes map[string]string) Input {
	return Input{
		inputType:    inputType,
		form:         f,
		name:         name,
		defaultValue: defaultValue,
		selector:     selector,
		attributes:   attributes,
	}
}

type StringInput struct {
	Input
}

func (f *Form) Text(name string, defaultValue string, selector string, attributes map[string]string) StringInput {
	return StringInput{Input: Input{
		inputType:    "text",
		form:         f,
		name:         name,
		defaultValue: defaultValue,
		selector:     selector,
		attributes:   attributes,
	}}
}

type NumberInput struct {
	Input
}

func (f *Form) Number(name string, defaultValue string, selector string, attributes map[string]string) NumberInput {
	return NumberInput{Input: Input{
		inputType:    "number",
		form:         f,
		name:         name,
		defaultValue: defaultValue,
		selector:     selector,
		attributes:   attributes,
	}}
}

func (i NumberInput) Int() int {
	s := i.String()
	num, err := strconv.Atoi(s)
	if err != nil {
		i.errs = append(i.errs, err)
	}
	return num
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
