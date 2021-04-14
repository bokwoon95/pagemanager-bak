package hypforms

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/bokwoon95/pagemanager/hyp"
)

type ctxKey string

const ctxKeyName ctxKey = "name"

func decorateErr(ctx context.Context, err error, value string) error {
	name, ok := ctx.Value(ctxKeyName).(string)
	if !ok {
		return fmt.Errorf("%w: value=%v", err, value)
	}
	return fmt.Errorf("%w: value=%s, name=%s", err, value, name)
}

type Validator func(ctx context.Context, value interface{}) (stop bool, err error)

func Validate(value interface{}, validators ...Validator) []error {
	return ValidateContext(context.Background(), value, validators...)
}

func ValidateContext(ctx context.Context, value interface{}, validators ...Validator) []error {
	var stop bool
	var err error
	var errs []error
	for _, validator := range validators {
		stop, err = validator(ctx, value)
		if err != nil {
			errs = append(errs, err)
		}
		if stop {
			return errs
		}
	}
	return errs
}

var RequiredErr = errors.New("field required")

func Required(ctx context.Context, value interface{}) (stop bool, err error) {
	str := hyp.Stringify(value)
	if str == "" {
		return true, decorateErr(ctx, RequiredErr, str)
	}
	return false, nil
}

// Optional

func Optional(ctx context.Context, value interface{}) (stop bool, err error) {
	str := hyp.Stringify(value)
	if str == "" {
		return true, nil
	}
	return false, nil
}

// IsRegexp

var IsRegexpErr = errors.New("value failed regexp match")

func IsRegexp(re *regexp.Regexp) Validator {
	return func(ctx context.Context, value interface{}) (stop bool, err error) {
		str := hyp.Stringify(value)
		if !re.MatchString(str) {
			return false, decorateErr(ctx, fmt.Errorf("%w %s", IsRegexpErr, re), str)
		}
		return false, nil
	}
}

// IsEmail

var emailRegexp = regexp.MustCompile( // https://emailregex.com/
	`(?:[a-z0-9!#$%&'*+/=?^_` + "`" + `{|}~-]+(?:\.[a-z0-9!#$%&'*+/=?^_` + "`" +
		`{|}~-]+)*|"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*")` +
		`@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\[(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\])`)

var IsEmailErr = errors.New("value is not an email")

func IsEmail(ctx context.Context, value interface{}) (stop bool, err error) {
	str := hyp.Stringify(value)
	if !emailRegexp.MatchString(str) {
		return false, decorateErr(ctx, IsEmailErr, str)
	}
	return false, nil
}

// AnyOf

var AnyOfErr = errors.New("value is not any the allowed strings")

func AnyOf(targets ...string) Validator {
	return func(ctx context.Context, value interface{}) (stop bool, err error) {
		str := hyp.Stringify(value)
		for _, target := range targets {
			if target == str {
				return false, decorateErr(ctx, fmt.Errorf("%w: %s", AnyOfErr, strings.Join(targets, " | ")), str)
			}
		}
		return false, nil
	}
}

// NoneOf

var NoneOfErr = errors.New("value is one of the disallowed strings")

func NoneOf(targets ...string) Validator {
	return func(ctx context.Context, value interface{}) (stop bool, err error) {
		str := hyp.Stringify(value)
		for _, target := range targets {
			if target == str {
				return false, decorateErr(ctx, fmt.Errorf("%w: %s", NoneOfErr, strings.Join(targets, " | ")), str)
			}
		}
		return false, nil
	}
}

// LengthGt, LengthGe, LengthLt, LengthLe

var LengthGtErr = errors.New("value length is not greater than")

func LengthGt(length int) Validator {
	return func(ctx context.Context, value interface{}) (stop bool, err error) {
		str := hyp.Stringify(value)
		if utf8.RuneCountInString(str) <= length {
			return false, decorateErr(ctx, fmt.Errorf("%w %d", LengthGtErr, length), str)
		}
		return false, nil
	}
}

var LengthGeErr = errors.New("value length is not greater than or equal to")

func LengthGe(length int) Validator {
	return func(ctx context.Context, value interface{}) (stop bool, err error) {
		str := hyp.Stringify(value)
		if utf8.RuneCountInString(str) < length {
			return false, decorateErr(ctx, fmt.Errorf("%w %d", LengthGeErr, length), str)
		}
		return false, nil
	}
}

var LengthLtErr = errors.New("value length is not less than")

func LengthLt(length int) Validator {
	return func(ctx context.Context, value interface{}) (stop bool, err error) {
		str := hyp.Stringify(value)
		if utf8.RuneCountInString(str) >= length {
			return false, decorateErr(ctx, fmt.Errorf("%w %d", LengthLtErr, length), str)
		}
		return false, nil
	}
}

var LengthLeErr = errors.New("value length is not less than or equal to")

func LengthLe(length int) Validator {
	return func(ctx context.Context, value interface{}) (stop bool, err error) {
		str := hyp.Stringify(value)
		if utf8.RuneCountInString(str) > length {
			return false, decorateErr(ctx, fmt.Errorf("%w %d", LengthLeErr, length), str)
		}
		return false, nil
	}
}

// IsIPAddr
// IsMACAddr
// IsUUID
