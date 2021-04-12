package hg

import "strconv"

type ValidatorFunc func(value interface{}, errs *[]error)
type Validator func(ValidatorFunc) ValidatorFunc

func NonEmpty(fn ValidatorFunc) ValidatorFunc {
	return func(value interface{}, errs *[]error) {
		switch value := value.(type) {
		case string:
			if value == "" {
			}
		case int:
			if value == 0 {
			}
		default:
			// dk what to do here
		}
		fn(value, errs)
	}
}

func CastToNumber(fn ValidatorFunc) ValidatorFunc {
	return func(value interface{}, errs *[]error) {
		switch value := value.(type) {
		case string:
			num, err := strconv.ParseFloat(value, 64)
			if err != nil {
				*errs = append(*errs, err)
			} else {
				fn(num, errs)
			}
		default:
			fn(value, errs)
		}
	}
}
