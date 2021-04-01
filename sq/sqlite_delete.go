package sq

import (
	"strings"
)

type SQLiteDeleteQuery struct {
	// WITH
	CTEs CTEs
	// DELETE FROM
	FromTable BaseTable
	// WHERE
	WherePredicate VariadicPredicate
	// ORDER BY
	OrderByFields Fields
	// LIMIT
	LimitValid bool
	LimitValue int64
	// OFFSET
	OffsetValid bool
	OffsetValue int64
}

func (q SQLiteDeleteQuery) AppendSQL(dialect string, buf *strings.Builder, args *[]interface{}, params map[string]int) error {
	// WITH
	if len(q.CTEs) > 0 {
		_ = q.CTEs.AppendCTEs(dialect, buf, args, params, q.FromTable, nil)
	}
	// DELETE FROM
	var err error
	buf.WriteString("DELETE FROM ")
	if q.FromTable == nil {
		buf.WriteString("NULL")
	} else {
		err = q.FromTable.AppendSQL("", buf, args, params)
		if err != nil {
			return err
		}
		alias := q.FromTable.GetAlias()
		if alias != "" {
			buf.WriteString(" AS ")
			buf.WriteString(alias)
		}
	}
	// WHERE
	if len(q.WherePredicate.Predicates) > 0 {
		buf.WriteString(" WHERE ")
		q.WherePredicate.Toplevel = true
		_ = q.WherePredicate.AppendSQLExclude("", buf, args, nil, nil)
	}
	// ORDER BY
	if len(q.OrderByFields) > 0 {
		buf.WriteString(" ORDER BY ")
		_ = q.OrderByFields.AppendSQLExclude("", buf, args, nil, nil)
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

func (q SQLiteDeleteQuery) ToSQL() (query string, args []interface{}, params map[string]int, err error) {
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

func (q SQLiteDeleteQuery) SetFetchableFields(fields []Field) (Query, error) {
	return nil, ErrUnsupported
}

func (q SQLiteDeleteQuery) GetFetchableFields() ([]Field, error) {
	return nil, ErrUnsupported
}

func (q SQLiteDeleteQuery) Dialect() string { return "sqlite3" }

func (_ SQLiteDialect) WithDelete(ctes ...CTE) SQLiteDeleteQuery {
	return SQLiteDeleteQuery{CTEs: ctes}
}

func (_ SQLiteDialect) DeleteFrom(table BaseTable) SQLiteDeleteQuery {
	return SQLiteDeleteQuery{FromTable: table}
}

func (q SQLiteDeleteQuery) With(ctes ...CTE) SQLiteDeleteQuery {
	q.CTEs = append(q.CTEs, ctes...)
	return q
}

func (q SQLiteDeleteQuery) DeleteFrom(table BaseTable) SQLiteDeleteQuery {
	q.FromTable = table
	return q
}

func (q SQLiteDeleteQuery) Where(predicates ...Predicate) SQLiteDeleteQuery {
	q.WherePredicate.Predicates = append(q.WherePredicate.Predicates, predicates...)
	return q
}

func (q SQLiteDeleteQuery) OrderBy(fields ...Field) SQLiteDeleteQuery {
	q.OrderByFields = append(q.OrderByFields, fields...)
	return q
}

func (q SQLiteDeleteQuery) Limit(limit int64) SQLiteDeleteQuery {
	q.LimitValid = true
	q.LimitValue = limit
	return q
}

func (q SQLiteDeleteQuery) Offset(offset int64) SQLiteDeleteQuery {
	q.OffsetValid = true
	q.OffsetValue = offset
	return q
}
