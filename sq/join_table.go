package sq

import "strings"

type JoinType string

const (
	JoinTypeInner JoinType = "JOIN"
	JoinTypeLeft  JoinType = "LEFT JOIN"
	JoinTypeRight JoinType = "RIGHT JOIN"
	JoinTypeFull  JoinType = "FULL JOIN"
)

type JoinTable struct {
	JoinType     JoinType
	Table        Table
	OnPredicates VariadicPredicate
}

func Join(table Table, predicates ...Predicate) JoinTable {
	return JoinTable{
		JoinType: JoinTypeInner,
		Table:    table,
		OnPredicates: VariadicPredicate{
			Predicates: predicates,
		},
	}
}

func LeftJoin(table Table, predicates ...Predicate) JoinTable {
	return JoinTable{
		JoinType: JoinTypeLeft,
		Table:    table,
		OnPredicates: VariadicPredicate{
			Predicates: predicates,
		},
	}
}

func RightJoin(table Table, predicates ...Predicate) JoinTable {
	return JoinTable{
		JoinType: JoinTypeRight,
		Table:    table,
		OnPredicates: VariadicPredicate{
			Predicates: predicates,
		},
	}
}

func FullJoin(table Table, predicates ...Predicate) JoinTable {
	return JoinTable{
		JoinType: JoinTypeFull,
		Table:    table,
		OnPredicates: VariadicPredicate{
			Predicates: predicates,
		},
	}
}

func CustomJoin(joinType JoinType, table Table, predicates ...Predicate) JoinTable {
	return JoinTable{
		JoinType: joinType,
		Table:    table,
		OnPredicates: VariadicPredicate{
			Predicates: predicates,
		},
	}
}

func (join JoinTable) AppendSQL(dialect string, buf *strings.Builder, args *[]interface{}, params map[string]int) error {
	if join.JoinType == "" {
		join.JoinType = JoinTypeInner
	}
	buf.WriteString(string(join.JoinType) + " ")
	switch v := join.Table.(type) {
	case nil:
		buf.WriteString("NULL")
	case Subquery:
		buf.WriteString("(")
		_ = v.AppendSQL("", buf, args, nil)
		buf.WriteString(")")
	default:
		_ = v.AppendSQL("", buf, args, nil)
	}
	if join.Table != nil {
		alias := join.Table.GetAlias()
		if alias != "" {
			buf.WriteString(" AS ")
			buf.WriteString(alias)
		}
	}
	if len(join.OnPredicates.Predicates) > 0 {
		buf.WriteString(" ON ")
		join.OnPredicates.Toplevel = true
		_ = join.OnPredicates.AppendSQLExclude("", buf, args, nil, nil)
	}
	return nil
}

type JoinTables []JoinTable

func (joins JoinTables) AppendSQL(dialect string, buf *strings.Builder, args *[]interface{}, params map[string]int) error {
	for i, join := range joins {
		if i > 0 {
			buf.WriteString(" ")
		}
		_ = join.AppendSQL(dialect, buf, args, nil)
	}
	return nil
}
