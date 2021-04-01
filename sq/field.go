package sq

import "strings"

type field struct {
	// field_name AS alias
	alias string
	// table_qualifier.field_name
	tableQualifier string
	tableName      string
	tableAlias     string
	// field_name
	name string
	// DESC/ASC
	descendingValid bool
	descending      bool
	// NULLS FIRST/NULLS LAST
	nullsfirstValid bool
	nullsfirst      bool
}

func (f field) GetAlias() string {
	return f.alias
}

func (f field) GetName() string {
	return f.name
}

func (f *field) asc() {
	f.descendingValid = true
	f.descending = false
}

func (f *field) desc() {
	f.descendingValid = true
	f.descending = true
}

func (f *field) nullsFirst() {
	f.nullsfirstValid = true
	f.nullsfirst = true
}

func (f *field) nullsLast() {
	f.nullsfirstValid = true
	f.nullsfirst = false
}

func (f field) AppendSQLExclude(dialect string, buf *strings.Builder, args *[]interface{}, params map[string]int, excludedTableQualifiers []string) error {
	tableQualifier := f.tableQualifier
	for _, excludedTableQualifier := range excludedTableQualifiers {
		if tableQualifier == excludedTableQualifier {
			tableQualifier = ""
			break
		}
	}
	if tableQualifier != "" {
		buf.WriteString(tableQualifier)
		buf.WriteString(".")
	}
	buf.WriteString(f.name)
	if f.descendingValid {
		if f.descending {
			buf.WriteString(" DESC")
		} else {
			buf.WriteString(" ASC")
		}
	}
	if f.nullsfirstValid {
		if f.nullsfirst {
			buf.WriteString(" NULLS FIRST")
		} else {
			buf.WriteString(" NULLS LAST")
		}
	}
	return nil
}

func (f field) String() string {
	buf := bufpool.Get().(*strings.Builder)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	_ = f.AppendSQLExclude("", buf, nil, nil, nil)
	return buf.String()
}

func (f field) IsNull() Predicate    { return Predicatef("? IS NULL", f) }
func (f field) IsNotNull() Predicate { return Predicatef("? IS NOT NULL", f) }
func (f field) In(v interface{}) Predicate {
	if v, ok := v.(RowValue); ok {
		return Predicatef("? IN ?", f, v)
	}
	return Predicatef("? IN (?)", f, v)
}

func (f field) Set(val interface{}) Assignment { return Assign(f, val) }

type CustomField struct {
	field
	format string
	values []interface{}
}

func Fieldf(format string, values ...interface{}) CustomField {
	return CustomField{format: format, values: values}
}

func FieldValue(value interface{}) CustomField { return Fieldf("?", value) }

func (f CustomField) As(alias string) CustomField {
	f.field.alias = alias
	return f
}

func (f CustomField) Asc() CustomField        { f.field.asc(); return f }
func (f CustomField) Desc() CustomField       { f.field.desc(); return f }
func (f CustomField) NullsFirst() CustomField { f.field.nullsFirst(); return f }
func (f CustomField) NullsLast() CustomField  { f.field.nullsLast(); return f }

func (f CustomField) AppendSQLExclude(dialect string, buf *strings.Builder, args *[]interface{}, params map[string]int, excludedTableQualifiers []string) error {
	if f.format == "" && len(f.values) == 0 {
		buf.WriteString(":blank:")
		return nil
	}
	_ = expandValues(buf, args, params, excludedTableQualifiers, f.format, f.values)
	_ = f.field.AppendSQLExclude("", buf, args, params, excludedTableQualifiers)
	return nil
}

func (f CustomField) String() string {
	buf := bufpool.Get().(*strings.Builder)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	var args []interface{}
	_ = f.AppendSQLExclude("", buf, &args, make(map[string]int), nil)
	return buf.String()
}

func (f CustomField) IsNull() Predicate    { return Predicatef("? IS NULL", f) }
func (f CustomField) IsNotNull() Predicate { return Predicatef("? IS NOT NULL", f) }
func (f CustomField) In(v interface{}) Predicate {
	if v, ok := v.(RowValue); ok {
		return Predicatef("? IN ?", f, v)
	}
	return Predicatef("? IN (?)", f, v)
}
func (f CustomField) Eq(v interface{}) Predicate { return Eq(f, v) }
func (f CustomField) Ne(v interface{}) Predicate { return Ne(f, v) }
func (f CustomField) Gt(v interface{}) Predicate { return Gt(f, v) }
func (f CustomField) Ge(v interface{}) Predicate { return Ge(f, v) }
func (f CustomField) Lt(v interface{}) Predicate { return Lt(f, v) }
func (f CustomField) Le(v interface{}) Predicate { return Le(f, v) }

type FieldLiteral string

func (f FieldLiteral) AppendSQLExclude(dialect string, buf *strings.Builder, args *[]interface{}, params map[string]int, excludedTableQualifiers []string) error {
	buf.WriteString(string(f))
	return nil
}

func (f FieldLiteral) GetAlias() string {
	return ""
}

func (f FieldLiteral) GetName() string {
	return string(f)
}

type Fields []Field

func (fs Fields) AppendSQLExclude(dialect string, buf *strings.Builder, args *[]interface{}, params map[string]int, excludedTableQualifiers []string) error {
	for i, field := range fs {
		if i > 0 {
			buf.WriteString(", ")
		}
		if field == nil {
			buf.WriteString("NULL")
		} else {
			_ = field.AppendSQLExclude(dialect, buf, args, params, excludedTableQualifiers)
		}
	}
	return nil
}

func (fs Fields) AppendSQLExcludeWithAlias(dialect string, buf *strings.Builder, args *[]interface{}, params map[string]int, excludedTableQualifiers []string) error {
	var alias string
	for i, f := range fs {
		if i > 0 {
			buf.WriteString(", ")
		}
		if f == nil {
			buf.WriteString("NULL")
		} else {
			_ = f.AppendSQLExclude(dialect, buf, args, params, excludedTableQualifiers)
			if alias = f.GetAlias(); alias != "" {
				buf.WriteString(" AS ")
				buf.WriteString(alias)
			}
		}
	}
	return nil
}
