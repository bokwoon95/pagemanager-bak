package sq

import (
	"strings"
	"testing"

	"github.com/bokwoon95/pagemanager/testutil"
)

func Test_PredicateCases(t *testing.T) {
	type TT struct {
		excludedTableQualifiers []string
		wantQuery               string
		wantArgs                []interface{}
	}

	assertField := func(t *testing.T, f PredicateCases, tt TT) {
		Is := testutil.New(t)
		var _ Field = f
		buf := &strings.Builder{}
		var args []interface{}
		_ = f.AppendSQLExclude("", buf, &args, make(map[string]int), tt.excludedTableQualifiers)
		Is.Equal(tt.wantQuery, buf.String())
		Is.Equal(f.alias, f.GetAlias())
		Is.Equal("", f.GetName())
	}
	t.Run("empty", func(t *testing.T) {
		f := PredicateCases{}
		tt := TT{wantQuery: "CASE END"}
		assertField(t, f, tt)
	})
	t.Run("1 case", func(t *testing.T) {
		u := NEW_USERS("u")
		f := CaseWhen(u.USER_ID.IsNull(), 5)
		tt := TT{wantQuery: "CASE WHEN u.user_id IS NULL THEN ? END", wantArgs: []interface{}{5}}
		assertField(t, f, tt)
	})
	t.Run("2 cases", func(t *testing.T) {
		u := NEW_USERS("u")
		f := CaseWhen(u.USER_ID.IsNull(), 5).When(u.PASSWORD.EqString("abc"), u.EMAIL)
		tt := TT{
			wantQuery: "CASE WHEN u.user_id IS NULL THEN ? WHEN u.password = ? THEN u.email END",
			wantArgs:  []interface{}{5, "abc"},
		}
		assertField(t, f.As("alias"), tt)
	})
	t.Run("2 cases, fallback", func(t *testing.T) {
		u := NEW_USERS("u")
		f := CaseWhen(u.USER_ID.IsNull(), 5).When(u.PASSWORD.EqString("abc"), u.EMAIL).Else(6789)
		tt := TT{
			wantQuery: "CASE WHEN u.user_id IS NULL THEN ? WHEN u.password = ? THEN u.email ELSE ? END",
			wantArgs:  []interface{}{5, "abc", 6789},
		}
		assertField(t, f, tt)
	})
}

func Test_SimpleCases(t *testing.T) {
	type TT struct {
		excludedTableQualifiers []string
		wantQuery               string
		wantArgs                []interface{}
	}

	assertField := func(t *testing.T, f SimpleCases, tt TT) {
		Is := testutil.New(t)
		var _ Field = f
		buf := &strings.Builder{}
		var args []interface{}
		_ = f.AppendSQLExclude("", buf, &args, make(map[string]int), tt.excludedTableQualifiers)
		Is.Equal(tt.wantQuery, buf.String())
		Is.Equal(f.alias, f.GetAlias())
		Is.Equal("", f.GetName())
	}
	t.Run("empty", func(t *testing.T) {
		f := SimpleCases{}
		tt := TT{wantQuery: "CASE NULL END"}
		assertField(t, f, tt)
	})
	t.Run("expression only", func(t *testing.T) {
		u := NEW_USERS("u")
		f := Case(u.USER_ID)
		tt := TT{wantQuery: "CASE u.user_id END"}
		assertField(t, f, tt)
	})
	t.Run("expression, 1 case", func(t *testing.T) {
		u := NEW_USERS("u")
		f := Case(u.USER_ID).When(99, 97)
		tt := TT{
			wantQuery: "CASE u.user_id WHEN ? THEN ? END",
			wantArgs:  []interface{}{99, 97},
		}
		assertField(t, f, tt)
	})
	t.Run("expression, 2 cases", func(t *testing.T) {
		u := NEW_USERS("u")
		f := Case(u.USER_ID).When(99, 97).When(u.PASSWORD, u.EMAIL)
		tt := TT{
			wantQuery: "CASE u.user_id WHEN ? THEN ? WHEN u.password THEN u.email END",
			wantArgs:  []interface{}{99, 97},
		}
		assertField(t, f.As("alias"), tt)
	})
	t.Run("expression, 2 cases, fallback", func(t *testing.T) {
		u := NEW_USERS("u")
		f := Case(u.USER_ID).When(99, 97).When(u.PASSWORD, u.EMAIL).Else("abcde")
		tt := TT{
			wantQuery: "CASE u.user_id WHEN ? THEN ? WHEN u.password THEN u.email ELSE ? END",
			wantArgs:  []interface{}{99, 97, "abcde"},
		}
		assertField(t, f, tt)
	})
}
