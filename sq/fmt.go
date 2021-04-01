package sq

import (
	"database/sql"
	"database/sql/driver"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func expandValues(buf *strings.Builder, args *[]interface{}, params map[string]int, excludedTableQualifiers []string, format string, values []interface{}) error {
	var value interface{}
	for i := strings.Index(format, "?"); i >= 0 && len(values) > 0; i = strings.Index(format, "?") {
		buf.WriteString(format[:i])
		format = format[i+1:]
		value, values = values[0], values[1:] // pop value from values
		_ = appendSQLValue(buf, args, params, excludedTableQualifiers, value)
	}
	buf.WriteString(format)
	return nil
}

func appendSQLValue(buf *strings.Builder, args *[]interface{}, params map[string]int, excludedTableQualifiers []string, value interface{}) error {
	switch v := value.(type) {
	case nil:
		buf.WriteString("NULL")
		return nil
	case interface {
		AppendSQLExclude(string, *strings.Builder, *[]interface{}, map[string]int, []string) error
	}:
		_ = v.AppendSQLExclude("", buf, args, params, excludedTableQualifiers)
		return nil
	case interface {
		AppendSQL(string, *strings.Builder, *[]interface{}, map[string]int) error
	}:
		// TODO: propogate this error (ugh this will pollute the entire system)
		_ = v.AppendSQL("", buf, args, params)
		return nil
	}
	typ := reflect.TypeOf(value)
	switch typ.Kind() {
	case reflect.Slice:
		if typ.Elem().Kind() == reflect.Uint8 {
			break
		}
		s := reflect.ValueOf(value)
		if l := s.Len(); l == 0 {
			buf.WriteString("NULL")
		} else {
			buf.WriteString("?")
			buf.WriteString(strings.Repeat(", ?", l-1))
			for i := 0; i < l; i++ {
				*args = append(*args, s.Index(i).Interface())
			}
		}
		return nil
	}
	buf.WriteString("?")
	*args = append(*args, value)
	return nil
}

func randomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const (
		letterIdxBits = 6                    // 6 bits to represent a letter index
		letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
		letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
	)
	var src = rand.NewSource(time.Now().UnixNano())
	sb := strings.Builder{}
	sb.Grow(n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			sb.WriteByte(letterBytes[idx])
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return sb.String()
}

func QuestionInterpolate(query string, args ...interface{}) string {
	buf := bufpool.Get().(*strings.Builder)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	for i := strings.Index(query, "?"); i >= 0 && len(args) > 0; i = strings.Index(query, "?") {
		buf.WriteString(query[:i])
		if len(query[i:]) > 1 && query[i:i+2] == "??" {
			buf.WriteString("?")
			query = query[i+2:]
			continue
		}
		interpolateSQLValue(buf, args[0])
		query = query[i+1:]
		args = args[1:]
	}
	buf.WriteString(query)
	return buf.String()
}

func DollarInterpolate(query string, args ...interface{}) string {
	buf := bufpool.Get().(*strings.Builder)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	oldnewSets := make(map[int][]string)
	for i, arg := range args {
		interpolateSQLValue(buf, arg)
		placeholder := "$" + strconv.Itoa(i+1)
		oldnewSets[len(placeholder)] = append(oldnewSets[len(placeholder)], placeholder, buf.String())
	}
	result := query
	for i := len(oldnewSets) + 1; i >= 2; i-- {
		result = strings.NewReplacer(oldnewSets[i]...).Replace(result)
	}
	return result
}

func interpolateSQLValue(buf *strings.Builder, value interface{}) {
	switch v := value.(type) {
	case nil:
		buf.WriteString("NULL")
	case bool:
		if v {
			buf.WriteString("TRUE")
		} else {
			buf.WriteString("FALSE")
		}
	case []byte:
		buf.WriteString(`x'`)
		buf.WriteString(hex.EncodeToString(v))
		buf.WriteString(`'`)
	case string:
		buf.WriteString("'")
		buf.WriteString(v)
		buf.WriteString("'")
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		buf.WriteString(fmt.Sprint(value))
	case time.Time:
		buf.WriteString("'")
		buf.WriteString(v.Format(time.RFC3339Nano))
		buf.WriteString("'")
	case driver.Valuer:
		Interface, err := v.Value()
		if err != nil {
			buf.WriteString(":")
			buf.WriteString(err.Error())
			buf.WriteString(":")
		} else {
			switch Concrete := Interface.(type) {
			case string:
				buf.WriteString("'")
				buf.WriteString(Concrete)
				buf.WriteString("'")
			case nil:
				buf.WriteString("NULL")
			default:
				buf.WriteString(":")
				buf.WriteString(fmt.Sprintf("%#v", value)) // give up, don't know what it is, resort to fmt.Sprintf
				buf.WriteString(":")
			}
		}
	default:
		b, err := json.Marshal(value)
		if err != nil {
			buf.WriteString(":")
			buf.WriteString(fmt.Sprintf("%#v", value)) // give up, don't know what it is, resort to fmt.Sprintf
			buf.WriteString(":")
		} else {
			buf.WriteString("'")
			buf.Write(b)
			buf.WriteString("'")
		}
	}
}

func appendSQLDisplay(buf *strings.Builder, value interface{}) {
	switch v := value.(type) {
	case nil:
		buf.WriteString("𝗡𝗨𝗟𝗟")
	case *sql.NullBool:
		if v.Valid {
			if v.Bool {
				buf.WriteString("true")
			} else {
				buf.WriteString("false")
			}
		} else {
			buf.WriteString("𝗡𝗨𝗟𝗟")
		}
	case *sql.NullFloat64:
		if v.Valid {
			buf.WriteString(fmt.Sprintf("%f", v.Float64))
		} else {
			buf.WriteString("𝗡𝗨𝗟𝗟")
		}
	case *sql.NullInt64:
		if v.Valid {
			buf.WriteString(strconv.FormatInt(v.Int64, 10))
		} else {
			buf.WriteString("𝗡𝗨𝗟𝗟")
		}
	case *sql.NullString:
		if v.Valid {
			buf.WriteString(v.String)
		} else {
			buf.WriteString("𝗡𝗨𝗟𝗟")
		}
	case *sql.NullTime:
		if v.Valid {
			buf.WriteString(v.Time.String())
		} else {
			buf.WriteString("𝗡𝗨𝗟𝗟")
		}
	case *[]byte:
		if v != nil && len(*v) > 0 {
			buf.WriteString("0x" + hex.EncodeToString(*v))
		} else {
			buf.WriteString("𝗡𝗨𝗟𝗟")
		}
	default:
		buf.WriteString(fmt.Sprintf("%#v", value))
	}
}

func QuestionToDollarPlaceholders(query string) string {
	buf := bufpool.Get().(*strings.Builder)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	var count int
	for i := strings.Index(query, "?"); i >= 0; i = strings.Index(query, "?") {
		buf.WriteString(query[:i])
		if len(query[i:]) > 1 && query[i:i+2] == "??" {
			buf.WriteString("?")
			query = query[i+2:]
		} else {
			count++
			buf.WriteString("$" + strconv.Itoa(count))
			query = query[i+1:]
		}
	}
	buf.WriteString(query)
	return buf.String()
}
