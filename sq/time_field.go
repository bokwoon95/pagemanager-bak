package sq

import "time"

type TimeField struct {
	field
}

func NewTimeField(name string, tableinfo TableInfo) TimeField {
	f := TimeField{field: field{name: name, tableQualifier: tableinfo.Name}}
	if tableinfo.Alias != "" {
		f.field.tableQualifier = tableinfo.Alias
	}
	return f
}

func (f TimeField) As(alias string) TimeField {
	f.field.alias = alias
	return f
}

func (f TimeField) Asc() TimeField        { f.field.asc(); return f }
func (f TimeField) Desc() TimeField       { f.field.desc(); return f }
func (f TimeField) NullsFirst() TimeField { f.field.nullsFirst(); return f }
func (f TimeField) NullsLast() TimeField  { f.field.nullsLast(); return f }

func (f TimeField) Eq(field TimeField) Predicate   { return Eq(f, field) }
func (f TimeField) Ne(field TimeField) Predicate   { return Ne(f, field) }
func (f TimeField) Gt(field TimeField) Predicate   { return Gt(f, field) }
func (f TimeField) Ge(field TimeField) Predicate   { return Ge(f, field) }
func (f TimeField) Lt(field TimeField) Predicate   { return Lt(f, field) }
func (f TimeField) Le(field TimeField) Predicate   { return Le(f, field) }
func (f TimeField) EqTime(val time.Time) Predicate { return Eq(f, val) }
func (f TimeField) NeTime(val time.Time) Predicate { return Ne(f, val) }
func (f TimeField) GtTime(val time.Time) Predicate { return Gt(f, val) }
func (f TimeField) GeTime(val time.Time) Predicate { return Ge(f, val) }
func (f TimeField) LtTime(val time.Time) Predicate { return Lt(f, val) }
func (f TimeField) LeTime(val time.Time) Predicate { return Le(f, val) }
func (f TimeField) Between(start, end TimeField) Predicate {
	return Predicatef("? BETWEEN ? AND ?", f, start, end)
}
func (f TimeField) NotBetween(start, end TimeField) Predicate {
	return Predicatef("? NOT BETWEEN ? AND ?", f, start, end)
}
func (f TimeField) BetweenTime(start, end TimeField) Predicate {
	return Predicatef("? BETWEEN ? AND ?", f, start, end)
}
func (f TimeField) NotBetweenTime(start, end TimeField) Predicate {
	return Predicatef("? NOT BETWEEN ? AND ?", f, start, end)
}

func (f TimeField) SetTime(val time.Time) Assignment { return Assign(f, val) }
