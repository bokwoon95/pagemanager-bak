package sq

type BlobField struct {
	field
}

func NewBlobField(name string, tableinfo TableInfo) BlobField {
	f := BlobField{field: field{name: name, tableQualifier: tableinfo.Name}}
	if tableinfo.Alias != "" {
		f.tableQualifier = tableinfo.Alias
	}
	return f
}

func (f BlobField) As(alias string) BlobField {
	f.field.alias = alias
	return f
}

func (f BlobField) Asc() BlobField        { f.field.asc(); return f }
func (f BlobField) Desc() BlobField       { f.field.desc(); return f }
func (f BlobField) NullsFirst() BlobField { f.field.nullsFirst(); return f }
func (f BlobField) NullsLast() BlobField  { f.field.nullsLast(); return f }

func (f BlobField) SetBlob(val []byte) Assignment { return Assign(f, val) }
