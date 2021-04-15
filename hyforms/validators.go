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

type Validator func(ctx context.Context, value interface{}) (stop bool, msg string)

func Validate(value interface{}, validators ...Validator) (msgs []string) {
	return ValidateContext(context.Background(), value, validators...)
}

func ValidateContext(ctx context.Context, value interface{}, validators ...Validator) (msgs []string) {
	var stop bool
	var msg string
	for _, validator := range validators {
		stop, msg = validator(ctx, value)
		if msg != "" {
			msgs = append(msgs, msg)
		}
		if stop {
			return msgs
		}
	}
	return msgs
}

type Error string

func (e Error) Error() string { return string(e) }

type ctxKey string

const ctxKeyName ctxKey = "name"

func deocrateMsg(ctx context.Context, msg string, value string) string {
	name, ok := ctx.Value(ctxKeyName).(string)
	if !ok {
		return fmt.Sprintf("%s: value=%v", msg, value)
	}
	return fmt.Sprintf("%s: value=%s, name=%s", msg, value, name)
}

const RequiredErr = "field required"

func Required(ctx context.Context, value interface{}) (stop bool, msg string) {
	var str string
	if value != nil {
		str = hy.Stringify(value)
	}
	if str == "" {
		return true, deocrateMsg(ctx, RequiredErr, str)
	}
	return false, ""
}

// Optional

func Optional(ctx context.Context, value interface{}) (stop bool, msg string) {
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

const IsRegexpErr = "value failed regexp match"

func IsRegexp(re *regexp.Regexp) Validator {
	return func(ctx context.Context, value interface{}) (stop bool, msg string) {
		var str string
		if value != nil {
			str = hy.Stringify(value)
		}
		if !re.MatchString(str) {
			return false, deocrateMsg(ctx, fmt.Sprintf("%s %s", IsRegexpErr, re), str)
		}
		return false, ""
	}
}

// IsEmail

var emailRegexp = regexp.MustCompile( // https://emailregex.com/
	`(?:[a-z0-9!#$%&'*+/=?^_` + "`" + `{|}~-]+(?:\.[a-z0-9!#$%&'*+/=?^_` + "`" +
		`{|}~-]+)*|"(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21\x23-\x5b\x5d-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])*")` +
		`@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\[(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?|[a-z0-9-]*[a-z0-9]:(?:[\x01-\x08\x0b\x0c\x0e-\x1f\x21-\x5a\x53-\x7f]|\\[\x01-\x09\x0b\x0c\x0e-\x7f])+)\])`)

const IsEmailErr = "value is not an email"

func IsEmail(ctx context.Context, value interface{}) (stop bool, msg string) {
	var str string
	if value != nil {
		str = hy.Stringify(value)
	}
	if !emailRegexp.MatchString(str) {
		return false, deocrateMsg(ctx, IsEmailErr, str)
	}
	return false, ""
}

// IsURL

// copied from govalidator:rxURL
var urlRegexp = regexp.MustCompile(`^((ftp|tcp|udp|wss?|https?):\/\/)?(\S+(:\S*)?@)?((([1-9]\d?|1\d\d|2[01]\d|22[0-3]|24\d|25[0-5])(\.(\d{1,2}|1\d\d|2[0-4]\d|25[0-5])){2}(?:\.([0-9]\d?|1\d\d|2[0-4]\d|25[0-5]))|(\[(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))\])|(([a-zA-Z0-9]([a-zA-Z0-9-_]+)?[a-zA-Z0-9]([-\.][a-zA-Z0-9]+)*)|(((www\.)|([a-zA-Z0-9]+([-_\.]?[a-zA-Z0-9])*[a-zA-Z0-9]\.[a-zA-Z0-9]+))?))?(([a-zA-Z\x{00a1}-\x{ffff}0-9]+-?-?)*[a-zA-Z\x{00a1}-\x{ffff}0-9]+)(?:\.([a-zA-Z\x{00a1}-\x{ffff}]{1,}))?))\.?(:(\d{1,5}))?((\/|\?|#)[^\s]*)?$`)

const IsURLErr = "value is not a URL"

// copied from govalidator:IsURL
func IsURL(ctx context.Context, value interface{}) (stop bool, msg string) {
	const maxURLRuneCount = 2083
	const minURLRuneCount = 3
	var str string
	if value != nil {
		str = hy.Stringify(value)
	}
	if str == "" || utf8.RuneCountInString(str) >= maxURLRuneCount || len(str) <= minURLRuneCount || strings.HasPrefix(str, ".") {
		return false, deocrateMsg(ctx, IsURLErr, str)
	}
	strTemp := str
	if strings.Contains(str, ":") && !strings.Contains(str, "://") {
		// support no indicated urlscheme but with colon for port number
		// http:// is appended so url.Parse will succeed, strTemp used so it does not impact rxURL.MatchString
		strTemp = "http://" + str
	}
	u, err := url.Parse(strTemp)
	if err != nil {
		return false, deocrateMsg(ctx, IsURLErr, str)
	}
	if strings.HasPrefix(u.Host, ".") {
		return false, deocrateMsg(ctx, IsURLErr, str)
	}
	if u.Host == "" && (u.Path != "" && !strings.Contains(u.Path, ".")) {
		return false, deocrateMsg(ctx, IsURLErr, str)
	}
	if !urlRegexp.MatchString(str) {
		return false, deocrateMsg(ctx, IsURLErr, str)
	}
	return false, ""
}

// AnyOf

const AnyOfErr = "value is not any the allowed strings"

func AnyOf(targets ...string) Validator {
	return func(ctx context.Context, value interface{}) (stop bool, msg string) {
		var str string
		if value != nil {
			str = hy.Stringify(value)
		}
		for _, target := range targets {
			if target == str {
				return false, ""
			}
		}
		return false, deocrateMsg(ctx, fmt.Sprintf("%s (%s)", AnyOfErr, strings.Join(targets, " | ")), str)
	}
}

// NoneOf

const NoneOfErr = "value is one of the disallowed strings"

func NoneOf(targets ...string) Validator {
	return func(ctx context.Context, value interface{}) (stop bool, msg string) {
		var str string
		if value != nil {
			str = hy.Stringify(value)
		}
		for _, target := range targets {
			if target == str {
				return false, deocrateMsg(ctx, fmt.Sprintf("%s (%s)", NoneOfErr, strings.Join(targets, " | ")), str)
			}
		}
		return false, ""
	}
}

// LengthGt, LengthGe, LengthLt, LengthLe

const LengthGtErr = "value length is not greater than"

func LengthGt(length int) Validator {
	return func(ctx context.Context, value interface{}) (stop bool, msg string) {
		var str string
		if value != nil {
			str = hy.Stringify(value)
		}
		if utf8.RuneCountInString(str) <= length {
			return false, deocrateMsg(ctx, fmt.Sprintf("%s %d", LengthGtErr, length), str)
		}
		return false, ""
	}
}

const LengthGeErr = "value length is not greater than or equal to"

func LengthGe(length int) Validator {
	return func(ctx context.Context, value interface{}) (stop bool, msg string) {
		var str string
		if value != nil {
			str = hy.Stringify(value)
		}
		if utf8.RuneCountInString(str) < length {
			return false, deocrateMsg(ctx, fmt.Sprintf("%s %d", LengthGeErr, length), str)
		}
		return false, ""
	}
}

const LengthLtErr = "value length is not less than"

func LengthLt(length int) Validator {
	return func(ctx context.Context, value interface{}) (stop bool, msg string) {
		var str string
		if value != nil {
			str = hy.Stringify(value)
		}
		if utf8.RuneCountInString(str) >= length {
			return false, deocrateMsg(ctx, fmt.Sprintf("%s %d", LengthLtErr, length), str)
		}
		return false, ""
	}
}

const LengthLeErr = "value length is not less than or equal to"

func LengthLe(length int) Validator {
	return func(ctx context.Context, value interface{}) (stop bool, msg string) {
		var str string
		if value != nil {
			str = hy.Stringify(value)
		}
		if utf8.RuneCountInString(str) > length {
			return false, deocrateMsg(ctx, fmt.Sprintf("%s %d", LengthLeErr, length), str)
		}
		return false, ""
	}
}

// IsIPAddr
// IsMACAddr
// IsUUID
