package sq

import (
	"strings"
)

type Assignment struct {
	LHS interface{}
	RHS interface{}
}

func Assign(LHS, RHS interface{}) Assignment {
	return Assignment{LHS: LHS, RHS: RHS}
}

func (a Assignment) AppendSQLExclude(dialect string, buf *strings.Builder, args *[]interface{}, params map[string]int, excludedTableQualifiers []string) error {
	_ = appendSQLValue(buf, args, params, excludedTableQualifiers, a.LHS)
	buf.WriteString(" = ")
	switch a.RHS.(type) {
	case Query:
		buf.WriteString("(")
		_ = appendSQLValue(buf, args, params, excludedTableQualifiers, a.RHS)
		buf.WriteString(")")
	default:
		_ = appendSQLValue(buf, args, params, excludedTableQualifiers, a.RHS)
	}
	return nil
}

type Assignments []Assignment

func (as Assignments) AppendSQLExclude(dialect string, buf *strings.Builder, args *[]interface{}, params map[string]int, excludedTableQualifiers []string) error {
	for i, a := range as {
		if i > 0 {
			buf.WriteString(", ")
		}
		_ = a.AppendSQLExclude(dialect, buf, args, params, excludedTableQualifiers)
	}
	return nil
}

func SetExcluded(field Field) Assignment {
	name := field.GetName()
	return Assignment{LHS: FieldLiteral(name), RHS: FieldLiteral("EXCLUDED." + name)}
}
