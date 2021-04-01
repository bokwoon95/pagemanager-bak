package sq

// func Exists(query Query) CustomPredicate { return Predicatef("EXISTS(?)", query) }

func Count() NumberField                { return NumberFieldf("COUNT(*)") }
func Sum(field interface{}) NumberField { return NumberFieldf("SUM(?)", field) }
func Avg(field interface{}) NumberField { return NumberFieldf("AVG(?)", field) }
func Min(field interface{}) NumberField { return NumberFieldf("MIN(?)", field) }
func Max(field interface{}) NumberField { return NumberFieldf("MAX(?)", field) }

func CountOver(w Window) NumberField                  { return NumberFieldf("COUNT(*) OVER ?", w) }
func SumOver(field interface{}, w Window) NumberField { return NumberFieldf("SUM(?) OVER ?", field, w) }
func AvgOver(field interface{}, w Window) NumberField { return NumberFieldf("AVG(?) OVER ?", field, w) }
func MinOver(field interface{}, w Window) NumberField { return NumberFieldf("MIN(?) OVER ?", field, w) }
func MaxOver(field interface{}, w Window) NumberField { return NumberFieldf("MAX(?) OVER ?", field, w) }
func RowNumberOver(w Window) NumberField              { return NumberFieldf("ROW_NUMBER() OVER ?", w) }
func RankOver(w Window) NumberField                   { return NumberFieldf("RANK() OVER ?", w) }
func DenseRankOver(w Window) NumberField              { return NumberFieldf("DENSE_RANK() OVER ?", w) }
func PercentRankOver(w Window) NumberField            { return NumberFieldf("PERCENT_RANK() OVER ?", w) }
func CumeDistOver(w Window) NumberField               { return NumberFieldf("CUME_DIST() OVER ?", w) }
func LeadOver(field interface{}, offset interface{}, fallback interface{}, w Window) CustomField {
	if offset == nil {
		offset = 1
	}
	return Fieldf("LEAD(?, ?, ?) OVER ?", field, offset, fallback, w)
}
func LagOver(field interface{}, offset interface{}, fallback interface{}, w Window) CustomField {
	if offset == nil {
		offset = 1
	}
	return Fieldf("LAG(?, ?, ?) OVER ?", field, offset, fallback, w)
}
func NtileOver(n int, w Window) NumberField { return NumberFieldf("NTILE(?) OVER ?", n, w) }
func FirstValueOver(field interface{}, w Window) CustomField {
	return Fieldf("FIRST_VALUE(?) OVER ?", field, w)
}
func LastValueOver(field interface{}, w Window) CustomField {
	return Fieldf("LAST_VALUE(?) OVER ?", field, w)
}
func NthValueOver(field interface{}, n int, w Window) CustomField {
	return Fieldf("NTH_VALUE(?, ?) OVER ?", field, n, w)
}
