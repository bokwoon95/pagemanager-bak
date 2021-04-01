package sq

import (
	"strings"
)

type VariadicQueryOperator string

const (
	QueryUnion        VariadicQueryOperator = "UNION"
	QueryUnionAll     VariadicQueryOperator = "UNION ALL"
	QueryIntersect    VariadicQueryOperator = "INTERSECT"
	QueryIntersectAll VariadicQueryOperator = "INTERSECT ALL"
	QueryExcept       VariadicQueryOperator = "EXCEPT"
	QueryExceptAll    VariadicQueryOperator = "EXCEPT ALL"
)

type VariadicQuery struct {
	TopLevel bool
	Operator VariadicQueryOperator
	Queries  []Query
}

func (vq VariadicQuery) ToSQL() (query string, args []interface{}, params map[string]int, err error) {
	buf := bufpool.Get().(*strings.Builder)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	params = make(map[string]int)
	err = vq.AppendSQL("", buf, &args, params)
	if err != nil {
		return query, args, params, err
	}
	return buf.String(), args, params, nil
}

func (vq VariadicQuery) AppendSQL(dialect string, buf *strings.Builder, args *[]interface{}, params map[string]int) error {
	var err error
	if vq.Operator == "" {
		vq.Operator = QueryUnion
	}
	if len(vq.Queries) == 0 {
		return nil
	}
	if len(vq.Queries) == 1 {
		switch q := vq.Queries[0].(type) {
		case nil:
			buf.WriteString("NULL")
		case VariadicQuery:
			q.TopLevel = true
			err = q.AppendSQL("", buf, args, params)
			if err != nil {
				return err
			}
		default:
			err = q.AppendSQL("", buf, args, params)
			if err != nil {
				return err
			}
		}
		return nil
	}
	if !vq.TopLevel {
		buf.WriteString("(")
	}
	for i, q := range vq.Queries {
		if i > 0 {
			buf.WriteString(" ")
			buf.WriteString(string(vq.Operator))
			buf.WriteString(" ")
		}
		switch q := q.(type) {
		case nil:
			buf.WriteString("NULL")
		case VariadicQuery:
			q.TopLevel = false
			err = q.AppendSQL("", buf, args, params)
			if err != nil {
				return err
			}
		default:
			err = q.AppendSQL("", buf, args, params)
			if err != nil {
				return err
			}
		}
	}
	if !vq.TopLevel {
		buf.WriteString(")")
	}
	return nil
}

func (vq VariadicQuery) SetFetchableFields(fields []Field) (Query, error) {
	return vq, ErrUnsupported
}

func (vq VariadicQuery) GetFetchableFields() ([]Field, error) {
	if len(vq.Queries) == 0 {
		return nil, nil
	}
	return vq.Queries[0].GetFetchableFields()
}

func (vq VariadicQuery) Dialect() string {
	if len(vq.Queries) == 0 {
		return ""
	}
	return vq.Queries[0].Dialect()
}

func Union(queries ...Query) VariadicQuery {
	return VariadicQuery{TopLevel: true, Operator: QueryUnion, Queries: queries}
}

func UnionAll(queries ...Query) VariadicQuery {
	return VariadicQuery{TopLevel: true, Operator: QueryUnionAll, Queries: queries}
}

func Intersect(queries ...Query) VariadicQuery {
	return VariadicQuery{TopLevel: true, Operator: QueryIntersect, Queries: queries}
}

func IntersectAll(queries ...Query) VariadicQuery {
	return VariadicQuery{TopLevel: true, Operator: QueryIntersectAll, Queries: queries}
}

func Except(queries ...Query) VariadicQuery {
	return VariadicQuery{TopLevel: true, Operator: QueryExcept, Queries: queries}
}

func ExceptAll(queries ...Query) VariadicQuery {
	return VariadicQuery{TopLevel: true, Operator: QueryExceptAll, Queries: queries}
}

func (vq VariadicQuery) CTE(name string, columns ...string) CTE {
	cte := NewCTE(vq, name, "", columns)
	if len(columns) > 0 {
		for _, column := range columns {
			cte[column] = Fieldf(name + "." + column)
		}
		return cte
	}
	if len(vq.Queries) > 0 && vq.Queries[0] != nil {
		fields, err := vq.Queries[0].GetFetchableFields()
		if err != nil {
			return cte
		}
		for _, field := range fields {
			column := getAliasOrName(field)
			cte[column] = Fieldf(name + "." + column)
		}
	}
	return cte
}
