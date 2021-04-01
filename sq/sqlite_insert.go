package sq

import (
	"strings"
)

type SQLiteInsertQuery struct {
	ColumnMapper func(*Column) error
	// WITH
	CTEs CTEs
	// INSERT INTO
	IntoTable     BaseTable
	InsertColumns Fields
	// VALUES
	RowValues RowValues
	// SELECT
	SelectQuery *SQLiteSelectQuery
	// ON CONFLICT
	HandleConflict      bool
	ConflictFields      Fields
	ConflictPredicate   VariadicPredicate
	Resolution          Assignments
	ResolutionPredicate VariadicPredicate
}

type SQLiteInsertConflict struct{ insertQuery *SQLiteInsertQuery }

func (q SQLiteInsertQuery) AppendSQL(dialect string, buf *strings.Builder, args *[]interface{}, params map[string]int) error {
	var excludedTableQualifiers []string
	if q.ColumnMapper != nil {
		col := NewColumn(ColumnModeInsert)
		err := q.ColumnMapper(col)
		if err != nil {
			return err
		}
		q.InsertColumns, q.RowValues = ColumnInsertResult(col)
	}
	// WITH
	if len(q.CTEs) > 0 {
		var tbl Table
		var jointbls JoinTables
		if q.SelectQuery != nil {
			tbl = q.SelectQuery.FromTable
			jointbls = q.SelectQuery.JoinTables
		}
		_ = q.CTEs.AppendCTEs(dialect, buf, args, params, tbl, jointbls)
	}
	// INSERT INTO
	var err error
	buf.WriteString("INSERT INTO ")
	if q.IntoTable == nil {
		buf.WriteString("NULL")
	} else {
		err = q.IntoTable.AppendSQL("", buf, args, params)
		if err != nil {
			return err
		}
		name := q.IntoTable.GetName()
		alias := q.IntoTable.GetAlias()
		if alias != "" {
			buf.WriteString(" AS ")
			buf.WriteString(alias)
			excludedTableQualifiers = append(excludedTableQualifiers, alias)
		} else {
			excludedTableQualifiers = append(excludedTableQualifiers, name)
		}
	}
	if len(q.InsertColumns) > 0 {
		buf.WriteString(" (")
		_ = q.InsertColumns.AppendSQLExclude("", buf, args, params, excludedTableQualifiers)
		buf.WriteString(")")
	}
	// VALUES/SELECT
	switch {
	case len(q.RowValues) > 0:
		buf.WriteString(" VALUES ")
		_ = q.RowValues.AppendSQL("", buf, args, nil)
	case q.SelectQuery != nil:
		buf.WriteString(" ")
		_ = q.SelectQuery.AppendSQL("", buf, args, nil)
	}
	// ON CONFLICT
	if q.HandleConflict {
		buf.WriteString(" ON CONFLICT")
		if len(q.ConflictFields) > 0 {
			buf.WriteString(" (")
			_ = q.ConflictFields.AppendSQLExclude("", buf, args, params, excludedTableQualifiers)
			buf.WriteString(")")
			if len(q.ConflictPredicate.Predicates) > 0 {
				buf.WriteString(" WHERE ")
				q.ConflictPredicate.Toplevel = true
				_ = q.ConflictPredicate.AppendSQLExclude("", buf, args, params, excludedTableQualifiers)
			}
		}
		if len(q.Resolution) > 0 {
			buf.WriteString(" DO UPDATE SET ")
			_ = q.Resolution.AppendSQLExclude("", buf, args, params, excludedTableQualifiers)
			if len(q.ResolutionPredicate.Predicates) > 0 {
				buf.WriteString(" WHERE ")
				q.ResolutionPredicate.Toplevel = true
				_ = q.ResolutionPredicate.AppendSQLExclude("", buf, args, params, nil)
			}
		} else {
			buf.WriteString(" DO NOTHING")
		}
	}
	return nil
}

func (q SQLiteInsertQuery) ToSQL() (query string, args []interface{}, params map[string]int, err error) {
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

func (q SQLiteInsertQuery) SetFetchableFields(fields []Field) (Query, error) {
	return nil, ErrUnsupported
}

func (q SQLiteInsertQuery) GetFetchableFields() ([]Field, error) {
	return nil, ErrUnsupported
}

func (q SQLiteInsertQuery) Dialect() string { return "sqlite3" }

func (_ SQLiteDialect) WithInsert(ctes ...CTE) SQLiteInsertQuery {
	return SQLiteInsertQuery{CTEs: ctes}
}

func (_ SQLiteDialect) InsertInto(table BaseTable) SQLiteInsertQuery {
	return SQLiteInsertQuery{IntoTable: table}
}

func (q SQLiteInsertQuery) With(ctes ...CTE) SQLiteInsertQuery {
	q.CTEs = append(q.CTEs, ctes...)
	return q
}

func (q SQLiteInsertQuery) InsertInto(table BaseTable) SQLiteInsertQuery {
	q.IntoTable = table
	return q
}

func (q SQLiteInsertQuery) Columns(fields ...Field) SQLiteInsertQuery {
	q.InsertColumns = fields
	return q
}

func (q SQLiteInsertQuery) Values(values ...interface{}) SQLiteInsertQuery {
	q.RowValues = append(q.RowValues, values)
	return q
}

func (q SQLiteInsertQuery) Valuesx(mapper func(*Column) error) SQLiteInsertQuery {
	q.ColumnMapper = mapper
	return q
}

func (q SQLiteInsertQuery) Select(selectQuery SQLiteSelectQuery) SQLiteInsertQuery {
	q.SelectQuery = &selectQuery
	return q
}

func (q SQLiteInsertQuery) OnConflict(fields ...Field) SQLiteInsertConflict {
	q.HandleConflict = true
	q.ConflictFields = fields
	return SQLiteInsertConflict{insertQuery: &q}
}

func (c SQLiteInsertConflict) Where(predicates ...Predicate) SQLiteInsertConflict {
	c.insertQuery.ConflictPredicate.Predicates = append(c.insertQuery.ConflictPredicate.Predicates, predicates...)
	return c
}

func (c SQLiteInsertConflict) DoNothing() SQLiteInsertQuery {
	if c.insertQuery == nil {
		return SQLiteInsertQuery{}
	}
	return *c.insertQuery
}

func (c SQLiteInsertConflict) DoUpdateSet(assignments ...Assignment) SQLiteInsertQuery {
	if c.insertQuery == nil {
		return SQLiteInsertQuery{}
	}
	c.insertQuery.Resolution = assignments
	return *c.insertQuery
}

func (q SQLiteInsertQuery) Where(predicates ...Predicate) SQLiteInsertQuery {
	q.ResolutionPredicate.Predicates = append(q.ResolutionPredicate.Predicates, predicates...)
	return q
}
