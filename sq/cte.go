package sq

import "strings"

// https://www.topster.net/text/utf-schriften.html serif italics
const (
	metadataQuery     = "𝑞𝑢𝑒𝑟𝑦"
	metadataRecursive = "𝑟𝑒𝑐𝑢𝑟𝑠𝑖𝑣𝑒"
	metadataName      = "𝑛𝑎𝑚𝑒"
	metadataAlias     = "𝑎𝑙𝑖𝑎𝑠"
	metadataColumns   = "𝑐𝑜𝑙𝑢𝑚𝑛𝑠"
)

type CTE map[string]CustomField

func NewCTE(q Query, name, alias string, columns []string) CTE {
	cte := map[string]CustomField{
		metadataQuery:   Fieldf("", q),
		metadataName:    Fieldf("", name),
		metadataAlias:   Fieldf("", alias),
		metadataColumns: Fieldf("", columns),
	}
	for _, column := range columns {
		name := name
		if alias != "" {
			name = alias
		}
		cte[column] = Fieldf(name + "." + column)
	}
	return cte
}

func (cte CTE) GetName() string {
	field := cte[metadataName]
	if len(field.values) > 0 {
		if name, ok := field.values[0].(string); ok {
			return name
		}
	}
	return ""
}

func (cte CTE) GetAlias() string {
	field := cte[metadataAlias]
	if len(field.values) > 0 {
		if alias, ok := field.values[0].(string); ok {
			return alias
		}
	}
	return ""
}

func (cte CTE) AppendSQL(dialect string, buf *strings.Builder, args *[]interface{}, params map[string]int) error {
	buf.WriteString(cte.GetName())
	return nil
}

type CTEs []CTE

func (ctes CTEs) AppendCTEs(dialect string, buf *strings.Builder, args *[]interface{}, params map[string]int, fromTable Table, joinTables []JoinTable) error {
	type TmpCTE struct {
		name    string
		columns []string
		query   Query
	}
	var tmpCTEs []TmpCTE
	cteNames := map[string]bool{} // track CTE names we have already seen; used to remove duplicates
	hasRecursiveCTE := false
	addTmpCTE := func(table Table) {
		cte, ok := table.(CTE)
		if !ok {
			return // not a CTE, skip
		}
		name := cte.GetName()
		if cteNames[name] {
			return // already seen this CTE, skip
		}
		cteNames[name] = true
		if !hasRecursiveCTE && cte.IsRecursive() {
			hasRecursiveCTE = true
		}
		tmpCTEs = append(tmpCTEs, TmpCTE{
			name:    name,
			columns: cte.GetColumns(),
			query:   cte.GetQuery(),
		})
	}
	for _, cte := range ctes {
		addTmpCTE(cte)
	}
	addTmpCTE(fromTable)
	for _, joinTable := range joinTables {
		addTmpCTE(joinTable.Table)
	}
	if len(tmpCTEs) == 0 {
		return nil // there were no CTEs in the list of tables, return
	}
	if hasRecursiveCTE {
		buf.WriteString("WITH RECURSIVE ")
	} else {
		buf.WriteString("WITH ")
	}
	for i, cte := range tmpCTEs {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(cte.name)
		if len(cte.columns) > 0 {
			buf.WriteString(" (")
			buf.WriteString(strings.Join(cte.columns, ", "))
			buf.WriteString(")")
		}
		buf.WriteString(" AS (")
		switch q := cte.query.(type) {
		case nil:
			buf.WriteString("NULL")
		case VariadicQuery:
			q.TopLevel = true
			q.AppendSQL("", buf, args, params)
		default:
			q.AppendSQL("", buf, args, params)
		}
		buf.WriteString(")")
	}
	buf.WriteString(" ")
	return nil
}

func (cte CTE) IsRecursive() bool {
	field := cte[metadataRecursive]
	if len(field.values) > 0 {
		if recursive, ok := field.values[0].(bool); ok {
			return recursive
		}
	}
	return false
}

func (cte CTE) GetQuery() Query {
	field := cte[metadataQuery]
	if len(field.values) > 0 {
		if q, ok := field.values[0].(Query); ok {
			return q
		}
	}
	return nil
}

func (cte CTE) GetColumns() []string {
	field := cte[metadataColumns]
	if len(field.values) > 0 {
		if columns, ok := field.values[0].([]string); ok {
			return columns
		}
	}
	return nil
}

func (cte CTE) As(alias string) CTE {
	newcte := NewCTE(cte.GetQuery(), cte.GetName(), alias, cte.GetColumns())
	for column := range cte {
		switch column {
		case metadataQuery, metadataName, metadataAlias, metadataColumns:
			continue
		}
		newcte[column] = Fieldf(alias + "." + column)
	}
	return newcte
}

func RecursiveCTE(name string, columns ...string) CTE {
	cte := map[string]CustomField{
		metadataRecursive: Fieldf("", true),
		metadataName:      Fieldf("", name),
		metadataAlias:     Fieldf("", ""),
	}
	if len(columns) > 0 {
		cte[metadataColumns] = Fieldf("", columns)
		for _, column := range columns {
			cte[column] = Fieldf(name + "." + column)
		}
	}
	return cte
}

type IntermediateCTE map[string]CustomField

func (cte *CTE) Initial(query Query) IntermediateCTE {
	if !cte.IsRecursive() {
		return IntermediateCTE(*cte)
	}
	(*cte)[metadataQuery] = Fieldf("", query)
	name := cte.GetName()
	columns := cte.GetColumns()
	if len(columns) > 0 {
		return IntermediateCTE(*cte)
	}
	fields, err := query.GetFetchableFields()
	if err != nil {
		return IntermediateCTE(*cte)
	}
	for _, field := range fields {
		column := getAliasOrName(field)
		(*cte)[column] = Fieldf(name + "." + column)
	}
	return IntermediateCTE(*cte)
}

func (cte IntermediateCTE) Union(queries ...Query) CTE {
	if !CTE(cte).IsRecursive() {
		return CTE(cte)
	}
	return cte.union(queries, QueryUnion)
}

func (cte IntermediateCTE) UnionAll(queries ...Query) CTE {
	if !CTE(cte).IsRecursive() {
		return CTE(cte)
	}
	return cte.union(queries, QueryUnionAll)
}

func (cte *IntermediateCTE) union(queries []Query, operator VariadicQueryOperator) CTE {
	initialQuery := CTE(*cte).GetQuery()
	(*cte)[metadataQuery] = Fieldf("", VariadicQuery{
		Operator: operator,
		Queries:  append([]Query{initialQuery}, queries...),
	})
	return CTE(*cte)
}
