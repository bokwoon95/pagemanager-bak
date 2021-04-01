package sq

import "strings"

type TableInfo struct {
	Schema string
	Name   string
	Alias  string
}

func (tbl TableInfo) AppendSQL(dialect string, buf *strings.Builder, args *[]interface{}, params map[string]int) error {
	if tbl.Schema != "" {
		buf.WriteString(tbl.Schema)
		buf.WriteString(".")
	}
	buf.WriteString(tbl.Name)
	return nil
}

func (tbl TableInfo) GetAlias() string  { return tbl.Alias }
func (tbl TableInfo) GetName() string   { return tbl.Name }
func (tbl TableInfo) GetSchema() string { return tbl.Schema }
