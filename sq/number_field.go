package sq

import "strings"

type NumberField struct {
	field
	format string
	values []interface{}
}

func NewNumberField(name string, tableinfo TableInfo) NumberField {
	f := NumberField{field: field{name: name, tableQualifier: tableinfo.Name}}
	if tableinfo.Alias != "" {
		f.field.tableQualifier = tableinfo.Alias
	}
	return f
}

func NumberFieldf(format string, values ...interface{}) NumberField {
	return NumberField{format: format, values: values}
}

func (f NumberField) As(alias string) NumberField {
	f.field.alias = alias
	return f
}

func (f NumberField) Asc() NumberField        { f.field.asc(); return f }
func (f NumberField) Desc() NumberField       { f.field.desc(); return f }
func (f NumberField) NullsFirst() NumberField { f.field.nullsFirst(); return f }
func (f NumberField) NullsLast() NumberField  { f.field.nullsLast(); return f }

func (f NumberField) AppendSQLExclude(dialect string, buf *strings.Builder, args *[]interface{}, params map[string]int, excludedTableQualifiers []string) error {
	if f.format != "" {
		_ = expandValues(buf, args, params, excludedTableQualifiers, f.format, f.values)
	}
	_ = f.field.AppendSQLExclude("", buf, args, params, excludedTableQualifiers)
	return nil
}

func (f NumberField) String() string {
	buf := bufpool.Get().(*strings.Builder)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	_ = f.AppendSQLExclude("", buf, nil, nil, nil)
	return buf.String()
}

func (f NumberField) IsNull() Predicate    { return Predicatef("? IS NULL", f) }
func (f NumberField) IsNotNull() Predicate { return Predicatef("? IS NOT NULL", f) }
func (f NumberField) In(v interface{}) Predicate {
	if v, ok := v.(RowValue); ok {
		return Predicatef("? IN ?", f, v)
	}
	return Predicatef("? IN (?)", f, v)
}
func (f NumberField) Eq(field NumberField) Predicate  { return Eq(f, field) }
func (f NumberField) Ne(field NumberField) Predicate  { return Ne(f, field) }
func (f NumberField) Gt(field NumberField) Predicate  { return Gt(f, field) }
func (f NumberField) Ge(field NumberField) Predicate  { return Ge(f, field) }
func (f NumberField) Lt(field NumberField) Predicate  { return Lt(f, field) }
func (f NumberField) Le(field NumberField) Predicate  { return Le(f, field) }
func (f NumberField) EqInt(val int) Predicate         { return Eq(f, val) }
func (f NumberField) NeInt(val int) Predicate         { return Ne(f, val) }
func (f NumberField) GtInt(val int) Predicate         { return Gt(f, val) }
func (f NumberField) GeInt(val int) Predicate         { return Ge(f, val) }
func (f NumberField) LtInt(val int) Predicate         { return Lt(f, val) }
func (f NumberField) LeInt(val int) Predicate         { return Le(f, val) }
func (f NumberField) EqInt64(val int64) Predicate     { return Eq(f, val) }
func (f NumberField) NeInt64(val int64) Predicate     { return Ne(f, val) }
func (f NumberField) GtInt64(val int64) Predicate     { return Gt(f, val) }
func (f NumberField) GeInt64(val int64) Predicate     { return Ge(f, val) }
func (f NumberField) LtInt64(val int64) Predicate     { return Lt(f, val) }
func (f NumberField) LeInt64(val int64) Predicate     { return Le(f, val) }
func (f NumberField) EqFloat64(val float64) Predicate { return Eq(f, val) }
func (f NumberField) NeFloat64(val float64) Predicate { return Ne(f, val) }
func (f NumberField) GtFloat64(val float64) Predicate { return Gt(f, val) }
func (f NumberField) GeFloat64(val float64) Predicate { return Ge(f, val) }
func (f NumberField) LtFloat64(val float64) Predicate { return Lt(f, val) }
func (f NumberField) LeFloat64(val float64) Predicate { return Le(f, val) }

func (f NumberField) SetInt(val int) Assignment         { return Assign(f, val) }
func (f NumberField) SetInt64(val int64) Assignment     { return Assign(f, val) }
func (f NumberField) SetFloat64(val float64) Assignment { return Assign(f, val) }
