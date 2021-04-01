package sq

type JSONField struct {
	field
}

func NewJSONField(name string, tableinfo TableInfo) JSONField {
	f := JSONField{field: field{name: name, tableQualifier: tableinfo.Name}}
	if tableinfo.Alias != "" {
		f.field.tableQualifier = tableinfo.Alias
	}
	return f
}

func (f JSONField) As(alias string) JSONField {
	f.field.alias = alias
	return f
}

func (f JSONField) Asc() JSONField        { f.field.asc(); return f }
func (f JSONField) Desc() JSONField       { f.field.desc(); return f }
func (f JSONField) NullsFirst() JSONField { f.field.nullsFirst(); return f }
func (f JSONField) NullsLast() JSONField  { f.field.nullsLast(); return f }
