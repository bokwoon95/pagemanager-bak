package sq

import "strings"

func Not(predicate Predicate) Predicate {
	return predicate.Not()
}

type CustomPredicate struct {
	alias    string
	format   string
	values   []interface{}
	negative bool
}

func Predicatef(format string, values ...interface{}) CustomPredicate {
	return CustomPredicate{format: format, values: values}
}

func (p CustomPredicate) As(alias string) CustomPredicate {
	p.alias = alias
	return p
}

func (p CustomPredicate) GetAlias() string {
	return p.alias
}

func (p CustomPredicate) GetName() string {
	return ""
}

func (p CustomPredicate) AppendSQLExclude(dialect string, buf *strings.Builder, args *[]interface{}, params map[string]int, excludedTableQualifiers []string) error {
	if p.negative {
		buf.WriteString("NOT ")
	}
	_ = expandValues(buf, args, params, excludedTableQualifiers, p.format, p.values)
	return nil
}

func (p CustomPredicate) String() string {
	buf := bufpool.Get().(*strings.Builder)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	_ = p.AppendSQLExclude("", buf, nil, nil, nil)
	return buf.String()
}

func (p CustomPredicate) Not() Predicate {
	p.negative = !p.negative
	return p
}

type VariadicPredicateOperator string

const (
	PredicateOr  VariadicPredicateOperator = "OR"
	PredicateAnd VariadicPredicateOperator = "AND"
)

type VariadicPredicate struct {
	// Toplevel indicates if the variadic predicate is the top level predicate
	// i.e. it does not need enclosing brackets
	Toplevel   bool
	Alias      string
	Operator   VariadicPredicateOperator
	Predicates []Predicate
	Negative   bool
}

func (p VariadicPredicate) AppendSQLExclude(dialect string, buf *strings.Builder, args *[]interface{}, params map[string]int, excludedTableQualifiers []string) error {
	if p.Operator == "" {
		p.Operator = PredicateAnd
	}
	switch len(p.Predicates) {
	case 0:
		return nil
	case 1:
		if p.Negative {
			buf.WriteString("NOT ")
		}
		switch v := p.Predicates[0].(type) {
		case nil:
			buf.WriteString("NULL")
		case VariadicPredicate:
			if !p.Toplevel {
				buf.WriteString("(")
			}
			v.Toplevel = true // if only one predicate and it is variadic, hoist it to the top level
			_ = v.AppendSQLExclude("", buf, args, params, excludedTableQualifiers)
			if !p.Toplevel {
				buf.WriteString(")")
			}
		default:
			_ = p.Predicates[0].AppendSQLExclude("", buf, args, params, excludedTableQualifiers)
		}
	default:
		if p.Negative {
			buf.WriteString("NOT ")
		}
		if !p.Toplevel {
			buf.WriteString("(")
		}
		for i, predicate := range p.Predicates {
			if i > 0 {
				buf.WriteString(" ")
				buf.WriteString(string(p.Operator))
				buf.WriteString(" ")
			}
			if predicate == nil {
				buf.WriteString("NULL")
			} else {
				_ = predicate.AppendSQLExclude("", buf, args, params, excludedTableQualifiers)
			}
		}
		if !p.Toplevel {
			buf.WriteString(")")
		}
	}
	return nil
}

func (p VariadicPredicate) Not() Predicate {
	p.Negative = !p.Negative
	return p
}

func (p VariadicPredicate) GetAlias() string {
	return p.Alias
}

func (p VariadicPredicate) GetName() string {
	return ""
}

func And(predicates ...Predicate) VariadicPredicate {
	return VariadicPredicate{Operator: PredicateAnd, Predicates: predicates}
}

func Or(predicates ...Predicate) VariadicPredicate {
	return VariadicPredicate{Operator: PredicateOr, Predicates: predicates}
}

func IsNull(f Field) Predicate      { return Predicatef("? IS NULL", f) }
func IsNotNull(f Field) Predicate   { return Predicatef("? IS NOT NULL", f) }
func Eq(a, b interface{}) Predicate { return Predicatef("? = ?", a, b) }
func Ne(a, b interface{}) Predicate { return Predicatef("? <> ?", a, b) }
func Gt(a, b interface{}) Predicate { return Predicatef("? > ?", a, b) }
func Ge(a, b interface{}) Predicate { return Predicatef("? >= ?", a, b) }
func Lt(a, b interface{}) Predicate { return Predicatef("? < ?", a, b) }
func Le(a, b interface{}) Predicate { return Predicatef("? <= ?", a, b) }
