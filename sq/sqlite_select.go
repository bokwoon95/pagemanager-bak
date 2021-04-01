package sq

import (
	"strings"
)

type SQLiteSelectQuery struct {
	// WITH
	CTEs CTEs
	// SELECT
	SelectType   SelectType
	SelectFields Fields
	// FROM
	FromTable  Table
	JoinTables JoinTables
	// WHERE
	WherePredicate VariadicPredicate
	// GROUP BY
	GroupByFields Fields
	// HAVING
	HavingPredicate VariadicPredicate
	// WINDOW
	Windows Windows
	// ORDER BY
	OrderByFields Fields
	// LIMIT
	LimitValid bool
	LimitValue int64
	// OFFSET
	OffsetValid bool
	OffsetValue int64
}

func (q SQLiteSelectQuery) AppendSQL(dialect string, buf *strings.Builder, args *[]interface{}, params map[string]int) error {
	// WITH
	if len(q.CTEs) > 0 {
		_ = q.CTEs.AppendCTEs(dialect, buf, args, params, q.FromTable, q.JoinTables)
	}
	// SELECT
	if q.SelectType == "" {
		q.SelectType = SelectTypeDefault
	}
	buf.WriteString(string(q.SelectType))
	if len(q.SelectFields) > 0 {
		buf.WriteString(" ")
		_ = q.SelectFields.AppendSQLExcludeWithAlias(dialect, buf, args, params, nil)
	} else {
		buf.WriteString(" 1")
	}
	// FROM
	var err error
	if q.FromTable != nil {
		buf.WriteString(" FROM ")
		switch v := q.FromTable.(type) {
		case Subquery:
			buf.WriteString("(")
			err = v.AppendSQL("", buf, args, nil)
			if err != nil {
				return err
			}
			buf.WriteString(")")
		default:
			err = v.AppendSQL("", buf, args, nil)
			if err != nil {
				return err
			}
		}
		alias := q.FromTable.GetAlias()
		if alias != "" {
			buf.WriteString(" AS ")
			buf.WriteString(alias)
		}
	}
	// JOIN
	if len(q.JoinTables) > 0 {
		buf.WriteString(" ")
		_ = q.JoinTables.AppendSQL("", buf, args, params)
	}
	// WHERE
	if len(q.WherePredicate.Predicates) > 0 {
		buf.WriteString(" WHERE ")
		q.WherePredicate.Toplevel = true
		_ = q.WherePredicate.AppendSQLExclude("", buf, args, params, nil)
	}
	// GROUP BY
	if len(q.GroupByFields) > 0 {
		buf.WriteString(" GROUP BY ")
		_ = q.GroupByFields.AppendSQLExclude("", buf, args, params, nil)
	}
	// HAVING
	if len(q.HavingPredicate.Predicates) > 0 {
		buf.WriteString(" HAVING ")
		q.HavingPredicate.Toplevel = true
		_ = q.HavingPredicate.AppendSQLExclude("", buf, args, params, nil)
	}
	// WINDOW
	if len(q.Windows) > 0 {
		buf.WriteString(" WINDOW ")
		_ = q.Windows.AppendSQL("", buf, args, params)
	}
	// ORDER BY
	if len(q.OrderByFields) > 0 {
		buf.WriteString(" ORDER BY ")
		_ = q.OrderByFields.AppendSQLExclude("", buf, args, params, nil)
	}
	// LIMIT
	if q.LimitValid {
		buf.WriteString(" LIMIT ?")
		if q.LimitValue < 0 {
			q.LimitValue = -q.LimitValue
		}
		*args = append(*args, q.LimitValue)
	}
	// OFFSET
	if q.OffsetValid {
		buf.WriteString(" OFFSET ?")
		if q.OffsetValue < 0 {
			q.OffsetValue = -q.OffsetValue
		}
		*args = append(*args, q.OffsetValue)
	}
	return nil
}

func (q SQLiteSelectQuery) ToSQL() (query string, args []interface{}, params map[string]int, err error) {
	buf := bufpool.Get().(*strings.Builder)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	params = make(map[string]int)
	err = q.AppendSQL("", buf, &args, params)
	if err != nil {
		return query, args, params, err
	}
	query = buf.String()
	return query, args, params, nil
}

func (q SQLiteSelectQuery) SetFetchableFields(fields []Field) (Query, error) {
	q.SelectFields = fields
	return q, nil
}

func (q SQLiteSelectQuery) GetFetchableFields() ([]Field, error) {
	return q.SelectFields, nil
}

func (q SQLiteSelectQuery) Dialect() string { return "sqlite3" }

func (_ SQLiteDialect) From(table Table) SQLiteSelectQuery {
	return SQLiteSelectQuery{FromTable: table}
}

func (_ SQLiteDialect) Select(fields ...Field) SQLiteSelectQuery {
	return SQLiteSelectQuery{SelectFields: fields}
}

func (_ SQLiteDialect) WithSelect(ctes ...CTE) SQLiteSelectQuery {
	return SQLiteSelectQuery{CTEs: ctes}
}

func (_ SQLiteDialect) SelectOne() SQLiteSelectQuery {
	return SQLiteSelectQuery{SelectFields: Fields{FieldLiteral("1")}}
}

func (_ SQLiteDialect) SelectDistinct(fields ...Field) SQLiteSelectQuery {
	return SQLiteSelectQuery{SelectType: SelectTypeDistinct, SelectFields: fields}
}

func (q SQLiteSelectQuery) With(ctes ...CTE) SQLiteSelectQuery {
	q.CTEs = append(q.CTEs, ctes...)
	return q
}

func (q SQLiteSelectQuery) Select(fields ...Field) SQLiteSelectQuery {
	q.SelectFields = append(q.SelectFields, fields...)
	return q
}

func (q SQLiteSelectQuery) SelectOne() SQLiteSelectQuery {
	q.SelectFields = Fields{FieldLiteral("1")}
	return q
}

func (q SQLiteSelectQuery) SelectAll() SQLiteSelectQuery {
	q.SelectFields = Fields{FieldLiteral("*")}
	return q
}

func (q SQLiteSelectQuery) SelectCount() SQLiteSelectQuery {
	q.SelectFields = Fields{FieldLiteral("COUNT(*)")}
	return q
}

func (q SQLiteSelectQuery) SelectDistinct(fields ...Field) SQLiteSelectQuery {
	q.SelectType = SelectTypeDistinct
	q.SelectFields = fields
	return q
}

func (q SQLiteSelectQuery) From(table Table) SQLiteSelectQuery {
	q.FromTable = table
	return q
}

func (q SQLiteSelectQuery) Join(table Table, predicate Predicate, predicates ...Predicate) SQLiteSelectQuery {
	predicates = append([]Predicate{predicate}, predicates...)
	q.JoinTables = append(q.JoinTables, Join(table, predicates...))
	return q
}

func (q SQLiteSelectQuery) LeftJoin(table Table, predicate Predicate, predicates ...Predicate) SQLiteSelectQuery {
	predicates = append([]Predicate{predicate}, predicates...)
	q.JoinTables = append(q.JoinTables, LeftJoin(table, predicates...))
	return q
}

func (q SQLiteSelectQuery) RightJoin(table Table, predicate Predicate, predicates ...Predicate) SQLiteSelectQuery {
	predicates = append([]Predicate{predicate}, predicates...)
	q.JoinTables = append(q.JoinTables, RightJoin(table, predicates...))
	return q
}

func (q SQLiteSelectQuery) FullJoin(table Table, predicate Predicate, predicates ...Predicate) SQLiteSelectQuery {
	predicates = append([]Predicate{predicate}, predicates...)
	q.JoinTables = append(q.JoinTables, FullJoin(table, predicates...))
	return q
}

func (q SQLiteSelectQuery) CustomJoin(joinType JoinType, table Table, predicates ...Predicate) SQLiteSelectQuery {
	q.JoinTables = append(q.JoinTables, CustomJoin(joinType, table, predicates...))
	return q
}

func (q SQLiteSelectQuery) Where(predicates ...Predicate) SQLiteSelectQuery {
	q.WherePredicate.Predicates = append(q.WherePredicate.Predicates, predicates...)
	return q
}

func (q SQLiteSelectQuery) GroupBy(fields ...Field) SQLiteSelectQuery {
	q.GroupByFields = append(q.GroupByFields, fields...)
	return q
}

func (q SQLiteSelectQuery) Having(predicates ...Predicate) SQLiteSelectQuery {
	q.HavingPredicate.Predicates = append(q.HavingPredicate.Predicates, predicates...)
	return q
}

func (q SQLiteSelectQuery) Window(windows ...Window) SQLiteSelectQuery {
	q.Windows = append(q.Windows, windows...)
	return q
}

func (q SQLiteSelectQuery) OrderBy(fields ...Field) SQLiteSelectQuery {
	q.OrderByFields = append(q.OrderByFields, fields...)
	return q
}

func (q SQLiteSelectQuery) Limit(limit int64) SQLiteSelectQuery {
	q.LimitValid = true
	q.LimitValue = limit
	return q
}

func (q SQLiteSelectQuery) Offset(offset int64) SQLiteSelectQuery {
	q.OffsetValid = true
	q.OffsetValue = offset
	return q
}

func (q SQLiteSelectQuery) CTE(name string, columns ...string) CTE {
	cte := NewCTE(q, name, "", columns)
	if len(columns) == 0 {
		for _, field := range q.SelectFields {
			column := getAliasOrName(field)
			cte[column] = Fieldf(name + "." + column)
		}
	}
	return cte
}

func (q SQLiteSelectQuery) Subquery(alias string) Subquery { return NewSubquery(q, alias) }
