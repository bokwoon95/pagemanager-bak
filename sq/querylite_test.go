package sq

import "strings"

type querylite struct {
	fields     Fields
	writeQuery string
	readQuery  string
	args       []interface{}
}

// TODO: remove hardcoded selectFields of querylite, use fieldliterals instead.
// That way GetFetchableFields on querylite will not be a dud, and querylite
// can play nice with CTE.Initial which invokes GetFetchableFields in the event
// that a column list was not provided.
func fieldliterals(fields ...string) []Field {
	fs := make([]Field, len(fields))
	for i := range fields {
		fs[i] = FieldLiteral(fields[i])
	}
	return fs
}

func (q querylite) AppendSQL(dialect string, buf *strings.Builder, args *[]interface{}, params map[string]int) error {
	if q.readQuery != "" {
		if len(q.fields) > 0 {
			buf.WriteString("SELECT ")
			_ = q.fields.AppendSQLExcludeWithAlias(dialect, buf, args, make(map[string]int), nil)
			buf.WriteString(" ")
		}
		buf.WriteString(q.readQuery)
		*args = append(*args, q.args...)
		return nil
	}
	if q.writeQuery != "" {
		buf.WriteString(q.writeQuery)
		*args = append(*args, q.args...)
		_ = q.fields.AppendSQLExcludeWithAlias(dialect, buf, args, make(map[string]int), nil)
		if len(q.fields) > 0 {
			buf.WriteString(" RETURNING ")
			_ = q.fields.AppendSQLExcludeWithAlias(dialect, buf, args, make(map[string]int), nil)
		}
		return nil
	}
	buf.WriteString("SELECT ")
	_ = q.fields.AppendSQLExcludeWithAlias(dialect, buf, args, make(map[string]int), nil)
	return nil
}
func (q querylite) ToSQL() (query string, args []interface{}, params map[string]int, err error) {
	buf := &strings.Builder{}
	params = make(map[string]int)
	err = q.AppendSQL("", buf, &args, params)
	if err != nil {
		return buf.String(), args, params, err
	}
	return buf.String(), args, params, nil
}
func (q querylite) SetFetchableFields(fields []Field) (Query, error) {
	q.fields = fields
	return q, nil
}
func (q querylite) GetFetchableFields() ([]Field, error) {
	return q.fields, nil
}
func (q querylite) Dialect() string { return "sqlite3" }
