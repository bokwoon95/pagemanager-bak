package sq

type StringField struct {
	field
}

func NewStringField(name string, tableinfo TableInfo) StringField {
	f := StringField{field: field{name: name, tableQualifier: tableinfo.Name}}
	if tableinfo.Alias != "" {
		f.field.tableQualifier = tableinfo.Alias
	}
	return f
}

func (f StringField) As(alias string) StringField {
	f.field.alias = alias
	return f
}

func (f StringField) Asc() StringField        { f.field.asc(); return f }
func (f StringField) Desc() StringField       { f.field.desc(); return f }
func (f StringField) NullsFirst() StringField { f.field.nullsFirst(); return f }
func (f StringField) NullsLast() StringField  { f.field.nullsLast(); return f }

func (f StringField) Eq(field StringField) Predicate      { return Eq(f, field) }
func (f StringField) Ne(field StringField) Predicate      { return Ne(f, field) }
func (f StringField) Gt(field StringField) Predicate      { return Gt(f, field) }
func (f StringField) Ge(field StringField) Predicate      { return Ge(f, field) }
func (f StringField) Lt(field StringField) Predicate      { return Lt(f, field) }
func (f StringField) Le(field StringField) Predicate      { return Le(f, field) }
func (f StringField) EqString(val string) Predicate       { return Eq(f, val) }
func (f StringField) NeString(val string) Predicate       { return Ne(f, val) }
func (f StringField) GtString(val string) Predicate       { return Gt(f, val) }
func (f StringField) GeString(val string) Predicate       { return Ge(f, val) }
func (f StringField) LtString(val string) Predicate       { return Lt(f, val) }
func (f StringField) LeString(val string) Predicate       { return Le(f, val) }
func (f StringField) LikeString(val string) Predicate     { return Predicatef("? LIKE ?", f, val) }
func (f StringField) NotLikeString(val string) Predicate  { return Predicatef("? NOT LIKE ?", f, val) }
func (f StringField) ILikeString(val string) Predicate    { return Predicatef("? ILIKE ?", f, val) }
func (f StringField) NotILikeString(val string) Predicate { return Predicatef("? NOT ILIKE ?", f, val) }

func (f StringField) SetString(val string) Assignment { return Assign(f, val) }
