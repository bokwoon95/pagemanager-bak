package hyforms

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/bokwoon95/pagemanager/hy"
)

type Validator func(ctx context.Context, value interface{}) (stop bool, errMsg string)

func Validate(value interface{}, validators ...Validator) (errMsgs []string) {
	return ValidateContext(context.Background(), value, validators...)
}

func ValidateContext(ctx context.Context, value interface{}, validators ...Validator) (errMsgs []string) {
	var stop bool
	var errMsg string
	for _, validator := range validators {
		stop, errMsg = validator(ctx, value)
		if errMsg != "" {
			errMsgs = append(errMsgs, errMsg)
		}
		if stop {
			return errMsgs
		}
	}
	return errMsgs
}

type ctxKey string

const ctxKeyName ctxKey = "name"

func deocrateErrMsg(ctx context.Context, errMsg string, value string) string {
	name, ok := ctx.Value(ctxKeyName).(string)
	if !ok {
		return fmt.Sprintf("%s: value=%v", errMsg, value)
	}
	return fmt.Sprintf("%s: value=%s, name=%s", errMsg, value, name)
}

const RequiredErrMsg = "\x00field required"

func Required(ctx context.Context, value interface{}) (stop bool, errMsg string) {
	var str string
	if value != nil {
		str = hy.Stringify(value)
	}
	if str == "" {
		return true, deocrateErrMsg(ctx, RequiredErrMsg, str)
	}
	return false, ""
}

// Optional

func Optional(ctx context.Context, value interface{}) (stop bool, errMsg string) {
	var str string
	if value != nil {
		str = hy.Stringify(value)
	}
	if str == "" {
		return true, ""
	}
	return false, ""
}

// IsRegexp

const IsRegexpErrMsg = "\x00value failed regexp match"

func IsRegexp(re *regexp.Regexp) Validator {
	return func(ctx context.Context, value interface{}) (stop bool, errMsg string) {
		var str string
		if value != nil {
			str = hy.Stringify(value)
		}
		if !re.MatchString(str) {
			return false, deocrateErrMsg(ctx, fmt.Sprintf("%s %s", IsRegexpErrMsg, re), str)
		}
		return false, ""
	}
}

// IsEmail

// https://emailregex.com/
var emailRegexp = regexp.MustCompile(`(?:[a-z0-9!#$%&'*+/=?^_` + "`" + `{|}~-]+(?:\.[a-z0-9!#$%&'*+/=?^_` + "`" + `{|}~-]+)*|"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\[(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\])`)

const IsEmailErrMsg = "\x00value is not an email"

func IsEmail(ctx context.Context, value interface{}) (stop bool, errMsg string) {
	var str string
	if value != nil {
		str = hy.Stringify(value)
	}
	if !emailRegexp.MatchString(str) {
		return false, deocrateErrMsg(ctx, IsEmailErrMsg, str)
	}
	return false, ""
}

// IsURL

// copied from govalidator:rxURL
var urlRegexp = regexp.MustCompile(`^((ftp|tcp|udp|wss?|https?):\/\/)?(\S+(:\S*)?@)?((([1-9]\d?|1\d\d|2[01]\d|22[0-3]|24\d|25[0-5])(\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])){2}(?:\.([0-9]\d?|1\d\d|2[0-4]\d|25[0-5]))|(\[(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))\])|(([a-zA-Z0-9]([a-zA-Z0-9-_]+)?[a-zA-Z0-9]([-\.][a-zA-Z0-9]+)*)|(((www\.)|([a-zA-Z0-9]+([-_\.]?[a-zA-Z0-9])*[a-zA-Z0-9]\.[a-zA-Z0-9]+))?))?(([a-zA-Z\x{00a1}-\x{ffff}0-9]+-?-?)*[a-zA-Z\x{00a1}-\x{ffff}0-9]+)(?:\.([a-zA-Z\x{00a1}-\x{ffff}]{1,}))?))\.?(:(\d{1,5}))?((\/|\?|#)[^\s]*)?$`)

const IsURLErrMsg = "\x00value is not a URL"

// copied from govalidator:IsURL
func IsURL(ctx context.Context, value interface{}) (stop bool, errMsg string) {
	const maxURLRuneCount = 2083
	const minURLRuneCount = 3
	var str string
	if value != nil {
		str = hy.Stringify(value)
	}
	if str == "" || utf8.RuneCountInString(str) >= maxURLRuneCount || len(str) <= minURLRuneCount || strings.HasPrefix(str, ".") {
		return false, deocrateErrMsg(ctx, IsURLErrMsg, str)
	}
	strTemp := str
	if strings.Contains(str, ":") && !strings.Contains(str, "://") {
		// support no indicated urlscheme but with colon for port number
		// http:// is appended so url.Parse will succeed, strTemp used so it does not impact rxURL.MatchString
		strTemp = "http://" + str
	}
	u, err := url.Parse(strTemp)
	if err != nil {
		return false, deocrateErrMsg(ctx, IsURLErrMsg, str)
	}
	if strings.HasPrefix(u.Host, ".") {
		return false, deocrateErrMsg(ctx, IsURLErrMsg, str)
	}
	if u.Host == "" && (u.Path != "" && !strings.Contains(u.Path, ".")) {
		return false, deocrateErrMsg(ctx, IsURLErrMsg, str)
	}
	if !urlRegexp.MatchString(str) {
		return false, deocrateErrMsg(ctx, IsURLErrMsg, str)
	}
	return false, ""
}

// AnyOf

const AnyOfErrMsg = "\x00value is not any the allowed strings"

func AnyOf(targets ...string) Validator {
	return func(ctx context.Context, value interface{}) (stop bool, errMsg string) {
		var str string
		if value != nil {
			str = hy.Stringify(value)
		}
		for _, target := range targets {
			if target == str {
				return false, ""
			}
		}
		return false, deocrateErrMsg(ctx, fmt.Sprintf("%s (%s)", AnyOfErrMsg, strings.Join(targets, " | ")), str)
	}
}

// NoneOf

const NoneOfErrMsg = "\x00value is one of the disallowed strings"

func NoneOf(targets ...string) Validator {
	return func(ctx context.Context, value interface{}) (stop bool, errMsg string) {
		var str string
		if value != nil {
			str = hy.Stringify(value)
		}
		for _, target := range targets {
			if target == str {
				return false, deocrateErrMsg(ctx, fmt.Sprintf("%s (%s)", NoneOfErrMsg, strings.Join(targets, " | ")), str)
			}
		}
		return false, ""
	}
}

// LengthGt, LengthGe, LengthLt, LengthLe

const LengthGtErrMsg = "\x00value length is not greater than"

func LengthGt(length int) Validator {
	return func(ctx context.Context, value interface{}) (stop bool, errMsg string) {
		var str string
		if value != nil {
			str = hy.Stringify(value)
		}
		if utf8.RuneCountInString(str) <= length {
			return false, deocrateErrMsg(ctx, fmt.Sprintf("%s %d", LengthGtErrMsg, length), str)
		}
		return false, ""
	}
}

const LengthGeErrMsg = "\x00value length is not greater than or equal to"

func LengthGe(length int) Validator {
	return func(ctx context.Context, value interface{}) (stop bool, errMsg string) {
		var str string
		if value != nil {
			str = hy.Stringify(value)
		}
		if utf8.RuneCountInString(str) < length {
			return false, deocrateErrMsg(ctx, fmt.Sprintf("%s %d", LengthGeErrMsg, length), str)
		}
		return false, ""
	}
}

const LengthLtErrMsg = "\x00value length is not less than"

func LengthLt(length int) Validator {
	return func(ctx context.Context, value interface{}) (stop bool, errMsg string) {
		var str string
		if value != nil {
			str = hy.Stringify(value)
		}
		if utf8.RuneCountInString(str) >= length {
			return false, deocrateErrMsg(ctx, fmt.Sprintf("%s %d", LengthLtErrMsg, length), str)
		}
		return false, ""
	}
}

const LengthLeErrMsg = "\x00value length is not less than or equal to"

func LengthLe(length int) Validator {
	return func(ctx context.Context, value interface{}) (stop bool, errMsg string) {
		var str string
		if value != nil {
			str = hy.Stringify(value)
		}
		if utf8.RuneCountInString(str) > length {
			return false, deocrateErrMsg(ctx, fmt.Sprintf("%s %d", LengthLeErrMsg, length), str)
		}
		return false, ""
	}
}

// IsIPAddr
// IsMACAddr
// IsUUID
