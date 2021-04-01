package sq

import "strings"

type RowValue []interface{}

func (r RowValue) AppendSQLExclude(dialect string, buf *strings.Builder, args *[]interface{}, params map[string]int, excludedTableQualifiers []string) error {
	buf.WriteString("(")
	for i, value := range r {
		if i > 0 {
			buf.WriteString(", ")
		}
		_ = appendSQLValue(buf, args, params, excludedTableQualifiers, value)
	}
	buf.WriteString(")")
	return nil
}

func (r RowValue) GetName() string  { return "" }
func (r RowValue) GetAlias() string { return "" }

func (r RowValue) AppendSQL(dialect string, buf *strings.Builder, args *[]interface{}, params map[string]int) error {
	return r.AppendSQLExclude(dialect, buf, args, params, nil)
}

func (r RowValue) In(v interface{}) CustomPredicate {
	if v, ok := v.(RowValue); ok {
		return Predicatef("? IN ?", r, v)
	}
	return Predicatef("? IN (?)", r, v)
}

type RowValues []RowValue

func (rs RowValues) AppendSQL(dialect string, buf *strings.Builder, args *[]interface{}, params map[string]int) error {
	for i, rowvalue := range rs {
		if i > 0 {
			buf.WriteString(", ")
		}
		_ = rowvalue.AppendSQL(dialect, buf, args, params)
	}
	return nil
}
