package sq

import (
	"strings"
)

func (f BooleanField) Not() Predicate {
	f.negative = !f.negative
	return f
}

type BooleanField struct {
	field
	negative bool
}

func NewBooleanField(name string, tableinfo TableInfo) BooleanField {
	f := BooleanField{field: field{name: name, tableQualifier: tableinfo.Name}}
	if tableinfo.Alias != "" {
		f.field.tableQualifier = tableinfo.Alias
	}
	return f
}

func (f BooleanField) As(alias string) BooleanField {
	f.field.alias = alias
	return f
}

func (f BooleanField) Asc() BooleanField        { f.field.asc(); return f }
func (f BooleanField) Desc() BooleanField       { f.field.desc(); return f }
func (f BooleanField) NullsFirst() BooleanField { f.field.nullsFirst(); return f }
func (f BooleanField) NullsLast() BooleanField  { f.field.nullsLast(); return f }

func (f BooleanField) AppendSQLExclude(dialect string, buf *strings.Builder, args *[]interface{}, params map[string]int, excludedTableQualifiers []string) error {
	if f.negative {
		buf.WriteString("NOT ")
	}
	_ = f.field.AppendSQLExclude("", buf, nil, nil, excludedTableQualifiers)
	return nil
}

func (f BooleanField) String() string {
	buf := bufpool.Get().(*strings.Builder)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	_ = f.AppendSQLExclude("", buf, nil, nil, nil)
	return buf.String()
}

func (f BooleanField) IsNull() Predicate               { return Predicatef("? IS NULL", f) }
func (f BooleanField) IsNotNull() Predicate            { return Predicatef("? IS NOT NULL", f) }
func (f BooleanField) Eq(field BooleanField) Predicate { return Eq(f, field) }
func (f BooleanField) Ne(field BooleanField) Predicate { return Ne(f, field) }

func (f BooleanField) SetBool(val bool) Assignment { return Assign(f, val) }
