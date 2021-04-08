package hg

import "strconv"

type input struct {
	form         *Form
	inputType    string
	name         string
	defaultValue string
	selector     string
	attributes   map[string]string
}

func (i input) Name() string {
	return i.name
}

func (i input) DefaultValue() string {
	return i.defaultValue
}

func (i input) Errors() []error {
	if i.form.mode != FormModeUnmarshal {
		return nil
	}
	return i.form.errors.inputErrors[i.name]
}

func (i input) String(validators ...func(interface{}) error) string {
	if i.form.mode != FormModeUnmarshal {
		return ""
	}
	s := i.form.request.FormValue(i.name)
	var err error
	for _, validator := range validators {
		err = validator(s)
		if err != nil {
			i.form.errors.inputErrors[i.name] = append(i.form.errors.inputErrors[i.name], err)
		}
	}
	return s
}

func (i input) Strings(validators ...func(interface{}) error) []string {
	if i.form.mode != FormModeUnmarshal {
		return nil
	}
	ss := i.form.request.Form[i.name]
	var err error
	if len(validators) > 0 {
		for _, s := range ss {
			for _, validator := range validators {
				err = validator(s)
				if err != nil {
					i.form.errors.inputErrors[i.name] = append(i.form.errors.inputErrors[i.name], err)
				}
			}
		}
	}
	return ss
}

type Input struct{ input }

func (f *Form) Input(inputType string, name string, defaultValue string) *Input {
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

type StringInput struct{ input }

func (f *Form) Text(name string, defaultValue string) *StringInput {
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

type NumberInput struct{ input }

func (f *Form) Number(name string, defaultValue float64) *NumberInput {
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

func (i *NumberInput) Int(validators ...func(interface{}) error) int {
	if i.form.mode != FormModeUnmarshal {
		return 0
	}
	s := i.String()
	num, err := strconv.Atoi(s)
	if err != nil {
		i.form.errors.inputErrors[i.name] = append(i.form.errors.inputErrors[i.name], err)
	} else {
		for _, validator := range validators {
			err = validator(num)
			if err != nil {
				i.form.errors.inputErrors[i.name] = append(i.form.errors.inputErrors[i.name], err)
			}
		}
	}
	return num
}

func (i *NumberInput) Ints(validators ...func(interface{}) error) []int {
	if i.form.mode != FormModeUnmarshal {
		return nil
	}
	ss := i.Strings()
	var nums []int
	for _, s := range ss {
		num, err := strconv.Atoi(s)
		if err != nil {
			i.form.errors.inputErrors[i.name] = append(i.form.errors.inputErrors[i.name], err)
		} else {
			for _, validator := range validators {
				err = validator(num)
				if err != nil {
					i.form.errors.inputErrors[i.name] = append(i.form.errors.inputErrors[i.name], err)
				}
			}
		}
		nums = append(nums, num)
	}
	return nums
}

func (i *NumberInput) Float64(validators ...func(interface{}) error) float64 {
	if i.form.mode != FormModeUnmarshal {
		return 0
	}
	s := i.String()
	num, err := strconv.ParseFloat(s, 64)
	if err != nil {
		i.form.errors.inputErrors[i.name] = append(i.form.errors.inputErrors[i.name], err)
	} else {
		for _, validator := range validators {
			err = validator(num)
			if err != nil {
				i.form.errors.inputErrors[i.name] = append(i.form.errors.inputErrors[i.name], err)
			}
		}
	}
	return num
}

func (i *NumberInput) Float64s(validators ...func(interface{}) error) []float64 {
	if i.form.mode != FormModeUnmarshal {
		return nil
	}
	ss := i.Strings()
	var nums []float64
	for _, s := range ss {
		num, err := strconv.ParseFloat(s, 64)
		if err != nil {
			i.form.errors.inputErrors[i.name] = append(i.form.errors.inputErrors[i.name], err)
		} else {
			for _, validator := range validators {
				err = validator(num)
				if err != nil {
					i.form.errors.inputErrors[i.name] = append(i.form.errors.inputErrors[i.name], err)
				}
			}
		}
		nums = append(nums, num)
	}
	return nums
}

type HiddenInput struct{ input }

func (f *Form) Hidden(name string, defaultValue string) *HiddenInput {
	return &HiddenInput{input: input{
		inputType:    "number",
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

type CheckboxInput struct{ input } // no! checkbox input should never return Strings()!

type CheckboxInputs struct {
	form           *Form
	name           string
	values         []string
	set            map[string]struct{}
	setInitialized bool
}

func (f *Form) Checkboxes(name string, values []string) *CheckboxInputs {
	return &CheckboxInputs{form: f, name: name, values: values, set: make(map[string]struct{})}
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
