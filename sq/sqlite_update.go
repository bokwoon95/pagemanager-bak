package sq

import (
	"strings"
)

type SQLiteUpdateQuery struct {
	ColumnMapper func(*Column) error
	// WITH
	CTEs CTEs
	// UPDATE
	UpdateTable BaseTable
	// SET
	Assignments Assignments
	// FROM
	FromTable  Table
	JoinTables JoinTables
	// WHERE
	WherePredicate VariadicPredicate
}

func (q SQLiteUpdateQuery) AppendSQL(dialect string, buf *strings.Builder, args *[]interface{}, params map[string]int) error {
	var excludedTableQualifiers []string
	if q.ColumnMapper != nil {
		col := NewColumn(ColumnModeUpdate)
		err := q.ColumnMapper(col)
		if err != nil {
			return err
		}
		q.Assignments = ColumnUpdateResult(col)
	}
	// WITH
	if len(q.CTEs) > 0 {
		_ = q.CTEs.AppendCTEs(dialect, buf, args, params, q.FromTable, q.JoinTables)
	}
	// UPDATE
	var err error
	buf.WriteString("UPDATE ")
	if q.UpdateTable == nil {
		buf.WriteString("NULL")
	} else {
		err = q.UpdateTable.AppendSQL(dialect, buf, args, nil)
		if err != nil {
			return err
		}
		name := q.UpdateTable.GetName()
		alias := q.UpdateTable.GetAlias()
		if alias != "" {
			buf.WriteString(" AS ")
			buf.WriteString(alias)
			excludedTableQualifiers = append(excludedTableQualifiers, alias)
		} else {
			excludedTableQualifiers = append(excludedTableQualifiers, name)
		}
	}
	// SET
	if len(q.Assignments) > 0 {
		buf.WriteString(" SET ")
		_ = q.Assignments.AppendSQLExclude(dialect, buf, args, nil, excludedTableQualifiers)
	}
	// FROM
	if q.FromTable != nil {
		buf.WriteString(" FROM ")
		switch v := q.FromTable.(type) {
		case Subquery:
			buf.WriteString("(")
			err = v.AppendSQL(dialect, buf, args, nil)
			if err != nil {
				return err
			}
			buf.WriteString(")")
		default:
			err = q.FromTable.AppendSQL(dialect, buf, args, nil)
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
		_ = q.JoinTables.AppendSQL(dialect, buf, args, nil)
	}
	// WHERE
	if len(q.WherePredicate.Predicates) > 0 {
		buf.WriteString(" WHERE ")
		q.WherePredicate.Toplevel = true
		_ = q.WherePredicate.AppendSQLExclude(dialect, buf, args, nil, nil)
	}
	return nil
}

func (q SQLiteUpdateQuery) ToSQL() (query string, args []interface{}, params map[string]int, err error) {
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

func (q SQLiteUpdateQuery) SetFetchableFields(fields []Field) (Query, error) {
	return nil, ErrUnsupported
}

func (q SQLiteUpdateQuery) GetFetchableFields() ([]Field, error) {
	return nil, ErrUnsupported
}

func (q SQLiteUpdateQuery) Dialect() string { return "sqlite3" }

func (_ SQLiteDialect) WithUpdate(ctes ...CTE) SQLiteUpdateQuery {
	return SQLiteUpdateQuery{CTEs: ctes}
}

func (_ SQLiteDialect) Update(table BaseTable) SQLiteUpdateQuery {
	return SQLiteUpdateQuery{UpdateTable: table}
}

func (q SQLiteUpdateQuery) With(ctes ...CTE) SQLiteUpdateQuery {
	q.CTEs = append(q.CTEs, ctes...)
	return q
}

func (q SQLiteUpdateQuery) Update(table BaseTable) SQLiteUpdateQuery {
	q.UpdateTable = table
	return q
}

func (q SQLiteUpdateQuery) Set(assignments ...Assignment) SQLiteUpdateQuery {
	q.Assignments = append(q.Assignments, assignments...)
	return q
}

func (q SQLiteUpdateQuery) Setx(mapper func(*Column) error) SQLiteUpdateQuery {
	q.ColumnMapper = mapper
	return q
}

func (q SQLiteUpdateQuery) From(table Table) SQLiteUpdateQuery {
	q.FromTable = table
	return q
}

func (q SQLiteUpdateQuery) Join(table Table, predicate Predicate, predicates ...Predicate) SQLiteUpdateQuery {
	predicates = append([]Predicate{predicate}, predicates...)
	q.JoinTables = append(q.JoinTables, Join(table, predicates...))
	return q
}

func (q SQLiteUpdateQuery) LeftJoin(table Table, predicate Predicate, predicates ...Predicate) SQLiteUpdateQuery {
	predicates = append([]Predicate{predicate}, predicates...)
	q.JoinTables = append(q.JoinTables, LeftJoin(table, predicates...))
	return q
}

func (q SQLiteUpdateQuery) RightJoin(table Table, predicate Predicate, predicates ...Predicate) SQLiteUpdateQuery {
	predicates = append([]Predicate{predicate}, predicates...)
	q.JoinTables = append(q.JoinTables, RightJoin(table, predicates...))
	return q
}

func (q SQLiteUpdateQuery) FullJoin(table Table, predicate Predicate, predicates ...Predicate) SQLiteUpdateQuery {
	predicates = append([]Predicate{predicate}, predicates...)
	q.JoinTables = append(q.JoinTables, FullJoin(table, predicates...))
	return q
}

func (q SQLiteUpdateQuery) CustomJoin(joinType JoinType, table Table, predicates ...Predicate) SQLiteUpdateQuery {
	q.JoinTables = append(q.JoinTables, CustomJoin(joinType, table, predicates...))
	return q
}

func (q SQLiteUpdateQuery) Where(predicates ...Predicate) SQLiteUpdateQuery {
	q.WherePredicate.Predicates = append(q.WherePredicate.Predicates, predicates...)
	return q
}
