package sq

import (
	"strings"
)

type PostgresSelectQuery struct {
	Alias string
	// WITH
	CTEs CTEs
	// SELECT
	SelectType   SelectType
	SelectFields Fields
	DistinctOn   Fields
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
	// WITH TIES
	WithTiesValid bool
	// OFFSET
	OffsetValid bool
	OffsetValue int64
}

func (q PostgresSelectQuery) AppendSQL(dialect string, buf *strings.Builder, args *[]interface{}, params map[string]int) error {
	// WITH
	if len(q.CTEs) > 0 {
		_ = q.CTEs.AppendCTEs(dialect, buf, args, params, q.FromTable, q.JoinTables)
	}
	// SELECT
	if q.SelectType == "" {
		q.SelectType = SelectTypeDefault
	}
	buf.WriteString(string(q.SelectType))
	if q.SelectType == SelectTypeDistinctOn {
		buf.WriteString(" (")
		_ = q.DistinctOn.AppendSQLExclude("", buf, args, params, nil)
		buf.WriteString(")")
	}
	if len(q.SelectFields) > 0 {
		buf.WriteString(" ")
		_ = q.SelectFields.AppendSQLExcludeWithAlias("", buf, args, params, nil)
	}
	// FROM
	if q.FromTable != nil {
		buf.WriteString(" FROM ")
		switch v := q.FromTable.(type) {
		case Query:
			buf.WriteString("(")
			_ = v.AppendSQL("", buf, args, nil)
			buf.WriteString(")")
		default:
			_ = v.AppendSQL("", buf, args, nil)
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
	if q.WithTiesValid {
		buf.WriteString(" WITH TIES")
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

func (q PostgresSelectQuery) ToSQL() (query string, args []interface{}, params map[string]int, err error) {
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
	query = QuestionToDollarPlaceholders(buf.String())
	return query, args, params, nil
}

func (q PostgresSelectQuery) SetFetchableFields(fields []Field) (Query, error) {
	q.SelectFields = fields
	return q, nil
}

func (q PostgresSelectQuery) GetFetchableFields() ([]Field, error) {
	return q.SelectFields, nil
}

func (q PostgresSelectQuery) Dialect() string { return "postgres" }

func (_ PostgresDialect) From(table Table) PostgresSelectQuery {
	return PostgresSelectQuery{FromTable: table}
}

func (_ PostgresDialect) Select(fields ...Field) PostgresSelectQuery {
	return PostgresSelectQuery{SelectFields: fields}
}

func (_ PostgresDialect) WithSelect(ctes ...CTE) PostgresSelectQuery {
	return PostgresSelectQuery{CTEs: ctes}
}

func (_ PostgresDialect) SelectOne() PostgresSelectQuery {
	return PostgresSelectQuery{SelectFields: Fields{FieldLiteral("1")}}
}

func (_ PostgresDialect) SelectDistinct(fields ...Field) PostgresSelectQuery {
	return PostgresSelectQuery{SelectType: SelectTypeDistinct, SelectFields: fields}
}

func (_ PostgresDialect) SelectDistinctOn(distinctFields ...Field) func(...Field) PostgresSelectQuery {
	return func(fields ...Field) PostgresSelectQuery {
		return PostgresSelectQuery{SelectType: SelectTypeDistinctOn, SelectFields: fields, DistinctOn: distinctFields}
	}
}

func (q PostgresSelectQuery) With(ctes ...CTE) PostgresSelectQuery {
	q.CTEs = append(q.CTEs, ctes...)
	return q
}

func (q PostgresSelectQuery) Select(fields ...Field) PostgresSelectQuery {
	q.SelectFields = append(q.SelectFields, fields...)
	return q
}

func (q PostgresSelectQuery) SelectOne() PostgresSelectQuery {
	q.SelectFields = Fields{FieldLiteral("1")}
	return q
}

func (q PostgresSelectQuery) SelectAll() PostgresSelectQuery {
	q.SelectFields = Fields{FieldLiteral("*")}
	return q
}

func (q PostgresSelectQuery) SelectCount() PostgresSelectQuery {
	q.SelectFields = Fields{FieldLiteral("COUNT(*)")}
	return q
}

func (q PostgresSelectQuery) SelectDistinct(fields ...Field) PostgresSelectQuery {
	q.SelectType = SelectTypeDistinct
	q.SelectFields = fields
	return q
}

func (q PostgresSelectQuery) SelectDistinctOn(distinctFields ...Field) func(...Field) PostgresSelectQuery {
	return func(fields ...Field) PostgresSelectQuery {
		q.SelectType = SelectTypeDistinctOn
		q.SelectFields = fields
		q.DistinctOn = distinctFields
		return q
	}
}

func (q PostgresSelectQuery) From(table Table) PostgresSelectQuery {
	q.FromTable = table
	return q
}

func (q PostgresSelectQuery) Join(table Table, predicate Predicate, predicates ...Predicate) PostgresSelectQuery {
	predicates = append([]Predicate{predicate}, predicates...)
	q.JoinTables = append(q.JoinTables, Join(table, predicates...))
	return q
}

func (q PostgresSelectQuery) LeftJoin(table Table, predicate Predicate, predicates ...Predicate) PostgresSelectQuery {
	predicates = append([]Predicate{predicate}, predicates...)
	q.JoinTables = append(q.JoinTables, LeftJoin(table, predicates...))
	return q
}

func (q PostgresSelectQuery) RightJoin(table Table, predicate Predicate, predicates ...Predicate) PostgresSelectQuery {
	predicates = append([]Predicate{predicate}, predicates...)
	q.JoinTables = append(q.JoinTables, RightJoin(table, predicates...))
	return q
}

func (q PostgresSelectQuery) FullJoin(table Table, predicate Predicate, predicates ...Predicate) PostgresSelectQuery {
	predicates = append([]Predicate{predicate}, predicates...)
	q.JoinTables = append(q.JoinTables, FullJoin(table, predicates...))
	return q
}

func (q PostgresSelectQuery) CustomJoin(joinType JoinType, table Table, predicates ...Predicate) PostgresSelectQuery {
	q.JoinTables = append(q.JoinTables, CustomJoin(joinType, table, predicates...))
	return q
}

func (q PostgresSelectQuery) Where(predicates ...Predicate) PostgresSelectQuery {
	q.WherePredicate.Predicates = append(q.WherePredicate.Predicates, predicates...)
	return q
}

func (q PostgresSelectQuery) GroupBy(fields ...Field) PostgresSelectQuery {
	q.GroupByFields = append(q.GroupByFields, fields...)
	return q
}

func (q PostgresSelectQuery) Having(predicates ...Predicate) PostgresSelectQuery {
	q.HavingPredicate.Predicates = append(q.HavingPredicate.Predicates, predicates...)
	return q
}

func (q PostgresSelectQuery) Window(windows ...Window) PostgresSelectQuery {
	q.Windows = append(q.Windows, windows...)
	return q
}

func (q PostgresSelectQuery) OrderBy(fields ...Field) PostgresSelectQuery {
	q.OrderByFields = append(q.OrderByFields, fields...)
	return q
}

func (q PostgresSelectQuery) Limit(limit int) PostgresSelectQuery {
	q.LimitValid = true
	q.LimitValue = int64(limit)
	return q
}

func (q PostgresSelectQuery) WithTies() PostgresSelectQuery {
	q.WithTiesValid = true
	return q
}

func (q PostgresSelectQuery) Offset(offset int) PostgresSelectQuery {
	q.OffsetValid = true
	q.OffsetValue = int64(offset)
	return q
}

func (q PostgresSelectQuery) CTE(name string, columns ...string) CTE {
	cte := NewCTE(q, name, "", columns)
	if len(columns) == 0 {
		for _, field := range q.SelectFields {
			column := getAliasOrName(field)
			cte[column] = Fieldf(name + "." + column)
		}
	}
	return cte
}
