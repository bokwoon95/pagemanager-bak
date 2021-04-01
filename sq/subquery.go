package sq

import (
	"fmt"
	"strings"
)

type Subquery map[string]CustomField

func NewSubquery(q Query, alias string) Subquery {
	subq := map[string]CustomField{
		metadataQuery: Fieldf("", q),
		metadataAlias: Fieldf("", alias),
	}
	fields, err := q.GetFetchableFields()
	if err != nil {
		return subq
	}
	for _, field := range fields {
		name := getAliasOrName(field)
		subq[name] = Fieldf(alias + "." + name)
	}
	return subq
}

func (subq Subquery) GetName() string {
	return ""
}

func (subq Subquery) GetAlias() string {
	field := subq[metadataAlias]
	if len(field.values) > 0 {
		if alias, ok := field.values[0].(string); ok {
			return alias
		}
	}
	return ""
}

func (subq Subquery) AppendSQL(dialect string, buf *strings.Builder, args *[]interface{}, params map[string]int) error {
	q := subq.GetQuery()
	if q == nil {
		return fmt.Errorf("empty subquery")
	}
	err := q.AppendSQL("", buf, args, nil)
	if err != nil {
		return err
	}
	return nil
}

func (subq Subquery) GetQuery() Query {
	field := subq[metadataQuery]
	if len(field.values) > 0 {
		if q, ok := field.values[0].(Query); ok {
			return q
		}
	}
	return nil
}
