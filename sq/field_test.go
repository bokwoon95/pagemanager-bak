package sq

import (
	"strings"
	"testing"

	"github.com/bokwoon95/pagemanager/testutil"
)

func Test_field(t *testing.T) {
	type TT struct {
		excludedTableQualifiers []string
		wantQuery               string
		wantArgs                []interface{}
	}

	assertField := func(t *testing.T, f field, tt TT) {
		is := testutil.New(t)
		var _ Field = f
		buf := &strings.Builder{}
		var args []interface{}
		f.AppendSQLExclude("", buf, &args, make(map[string]int), tt.excludedTableQualifiers)
		is.Equal(tt.wantQuery, buf.String())
		is.Equal(f.alias, f.GetAlias())
		is.Equal(f.name, f.GetName())
		if len(tt.excludedTableQualifiers) == 0 {
			is.Equal(f.String(), buf.String())
		}
	}
	t.Run("table qualified", func(t *testing.T) {
		u := NEW_USERS("")
		tt := TT{wantQuery: "users.user_id"}
		assertField(t, u.USER_ID.field, tt)
	})
	t.Run("table alias qualified", func(t *testing.T) {
		u := NEW_USERS("u")
		tt := TT{wantQuery: "u.user_id"}
		assertField(t, u.USER_ID.field, tt)
	})
	t.Run("excludedTableQualifiers (name)", func(t *testing.T) {
		u := NEW_USERS("")
		tt := TT{wantQuery: "user_id", excludedTableQualifiers: []string{"users"}}
		assertField(t, u.USER_ID.field, tt)
	})
	t.Run("excludedTableQualifiers (alias)", func(t *testing.T) {
		u := NEW_USERS("u")
		tt := TT{wantQuery: "user_id", excludedTableQualifiers: []string{"u"}}
		assertField(t, u.USER_ID.field, tt)
	})
	t.Run("ASC", func(t *testing.T) {
		u := NEW_USERS("u")
		u.USER_ID.field.asc()
		tt := TT{wantQuery: "user_id ASC", excludedTableQualifiers: []string{"u"}}
		assertField(t, u.USER_ID.field, tt)
	})
	t.Run("DESC", func(t *testing.T) {
		u := NEW_USERS("u")
		u.USER_ID.field.desc()
		tt := TT{wantQuery: "user_id DESC", excludedTableQualifiers: []string{"u"}}
		assertField(t, u.USER_ID.field, tt)
	})
	t.Run("NULLS FIRST", func(t *testing.T) {
		u := NEW_USERS("u")
		u.USER_ID.field.nullsFirst()
		tt := TT{wantQuery: "user_id NULLS FIRST", excludedTableQualifiers: []string{"u"}}
		assertField(t, u.USER_ID.field, tt)
	})
	t.Run("NULLS LAST", func(t *testing.T) {
		u := NEW_USERS("u")
		u.USER_ID.field.nullsLast()
		tt := TT{wantQuery: "user_id NULLS LAST", excludedTableQualifiers: []string{"u"}}
		assertField(t, u.USER_ID.field, tt)
	})

	assertPredicate := func(t *testing.T, p Predicate, tt TT) {
		Is := testutil.New(t)
		buf := &strings.Builder{}
		var args []interface{}
		p.AppendSQLExclude("", buf, &args, make(map[string]int), tt.excludedTableQualifiers)
		Is.Equal(tt.wantQuery, buf.String())
		Is.Equal(tt.wantArgs, args)
	}
	t.Run("IS NULL", func(t *testing.T) {
		u := NEW_USERS("u")
		tt := TT{wantQuery: "user_id IS NULL", excludedTableQualifiers: []string{"u"}}
		assertPredicate(t, u.USER_ID.field.IsNull(), tt)
	})
	t.Run("IS NOT NULL", func(t *testing.T) {
		u := NEW_USERS("u")
		tt := TT{wantQuery: "user_id IS NOT NULL", excludedTableQualifiers: []string{"u"}}
		assertPredicate(t, u.USER_ID.field.IsNotNull(), tt)
	})
	t.Run("IN (rowvalues)", func(t *testing.T) {
		u := NEW_USERS("u")
		tt := TT{wantQuery: "u.user_id IN (u.user_id, u.displayname, u.password)"}
		assertPredicate(t, u.USER_ID.field.In(RowValue{u.USER_ID, u.DISPLAYNAME, u.PASSWORD}), tt)
	})
	t.Run("IN (slice)", func(t *testing.T) {
		u := NEW_USERS("u")
		tt := TT{wantQuery: "u.user_id IN (?, ?, ?)", wantArgs: []interface{}{5, 6, 7}}
		assertPredicate(t, u.USER_ID.field.In([]int{5, 6, 7}), tt)
	})

	assertAssignment := func(t *testing.T, a Assignment, tt TT) {
		Is := testutil.New(t)
		buf := &strings.Builder{}
		var args []interface{}
		a.AppendSQLExclude("", buf, &args, make(map[string]int), tt.excludedTableQualifiers)
		Is.Equal(tt.wantQuery, buf.String())
		Is.Equal(tt.wantArgs, args)
	}
	t.Run("Set", func(t *testing.T) {
		u := NEW_USERS("")
		tt := TT{wantQuery: "users.user_id = ?", wantArgs: []interface{}{99}}
		assertAssignment(t, u.USER_ID.Set(99), tt)
	})
	t.Run("Set (excludedTableQualifiers)", func(t *testing.T) {
		u := NEW_USERS("u")
		tt := TT{wantQuery: "user_id = ?", wantArgs: []interface{}{22}, excludedTableQualifiers: []string{"u"}}
		assertAssignment(t, u.USER_ID.Set(22), tt)
	})
}

func Test_CustomField(t *testing.T) {
	type TT struct {
		excludedTableQualifiers []string
		wantQuery               string
		wantArgs                []interface{}
	}

	assertField := func(t *testing.T, f CustomField, tt TT) {
		Is := testutil.New(t)
		var _ Field = f
		buf := &strings.Builder{}
		var args []interface{}
		f.AppendSQLExclude("", buf, &args, make(map[string]int), tt.excludedTableQualifiers)
		Is.Equal(tt.wantQuery, buf.String())
		Is.Equal(tt.wantArgs, args)
		Is.Equal(f.alias, f.GetAlias())
		Is.Equal(f.name, f.GetName())
		if len(tt.excludedTableQualifiers) == 0 {
			Is.Equal(f.String(), buf.String())
		}
	}
	t.Run("empty", func(t *testing.T) {
		tt := TT{wantQuery: ":blank:"}
		assertField(t, CustomField{}, tt)
	})
	t.Run("FieldValue", func(t *testing.T) {
		tt := TT{wantQuery: "?", wantArgs: []interface{}{"abcd"}}
		assertField(t, FieldValue("abcd"), tt)
	})
	t.Run("Fieldf", func(t *testing.T) {
		tt := TT{wantQuery: "lorem ipsum ? ?", wantArgs: []interface{}{1, "a"}}
		assertField(t, Fieldf("lorem ipsum ? ?", 1, "a"), tt)
	})
	t.Run("alias", func(t *testing.T) {
		tt := TT{wantQuery: "my_field"}
		assertField(t, Fieldf("my_field").As("ggggggg"), tt)
	})
	t.Run("ASC NULLS LAST", func(t *testing.T) {
		tt := TT{wantQuery: "my_field ASC NULLS LAST"}
		assertField(t, Fieldf("my_field").Asc().NullsLast(), tt)
	})
	t.Run("DESC NULLS FIRST", func(t *testing.T) {
		tt := TT{wantQuery: "my_field DESC NULLS FIRST"}
		assertField(t, Fieldf("my_field").Desc().NullsFirst(), tt)
	})

	assertPredicate := func(t *testing.T, p Predicate, tt TT) {
		Is := testutil.New(t)
		buf := &strings.Builder{}
		var args []interface{}
		p.AppendSQLExclude("", buf, &args, make(map[string]int), tt.excludedTableQualifiers)
		Is.Equal(tt.wantQuery, buf.String())
		Is.Equal(tt.wantArgs, args)
	}
	t.Run("IS NULL", func(t *testing.T) {
		tt := TT{wantQuery: "my_field IS NULL", excludedTableQualifiers: []string{"u"}}
		assertPredicate(t, Fieldf("my_field").IsNull(), tt)
	})
	t.Run("IS NOT NULL", func(t *testing.T) {
		tt := TT{wantQuery: "my_field IS NOT NULL", excludedTableQualifiers: []string{"u"}}
		assertPredicate(t, Fieldf("my_field").IsNotNull(), tt)
	})
	t.Run("IN (rowvalues)", func(t *testing.T) {
		u := NEW_USERS("u")
		tt := TT{wantQuery: "my_field IN (u.user_id, u.displayname, u.password)"}
		assertPredicate(t, Fieldf("my_field").In(RowValue{u.USER_ID, u.DISPLAYNAME, u.PASSWORD}), tt)
	})
	t.Run("IN (slice)", func(t *testing.T) {
		tt := TT{wantQuery: "my_field IN (?, ?, ?)", wantArgs: []interface{}{5, 6, 7}}
		assertPredicate(t, Fieldf("my_field").In([]int{5, 6, 7}), tt)
	})
	t.Run("Eq", func(t *testing.T) {
		tt := TT{wantQuery: "my_field = ?", wantArgs: []interface{}{123}}
		assertPredicate(t, Fieldf("my_field").Eq(123), tt)
	})
	t.Run("Ne", func(t *testing.T) {
		tt := TT{wantQuery: "my_field <> ?", wantArgs: []interface{}{123}}
		assertPredicate(t, Fieldf("my_field").Ne(123), tt)
	})
	t.Run("Gt", func(t *testing.T) {
		tt := TT{wantQuery: "my_field > ?", wantArgs: []interface{}{123}}
		assertPredicate(t, Fieldf("my_field").Gt(123), tt)
	})
	t.Run("Ge", func(t *testing.T) {
		tt := TT{wantQuery: "my_field >= ?", wantArgs: []interface{}{123}}
		assertPredicate(t, Fieldf("my_field").Ge(123), tt)
	})
	t.Run("Lt", func(t *testing.T) {
		tt := TT{wantQuery: "my_field < ?", wantArgs: []interface{}{123}}
		assertPredicate(t, Fieldf("my_field").Lt(123), tt)
	})
	t.Run("Lt", func(t *testing.T) {
		tt := TT{wantQuery: "my_field <= ?", wantArgs: []interface{}{123}}
		assertPredicate(t, Fieldf("my_field").Le(123), tt)
	})
}

func Test_FieldLiteral(t *testing.T) {
	type TT struct {
		excludedTableQualifiers []string
		wantQuery               string
		wantArgs                []interface{}
	}

	assertField := func(t *testing.T, f FieldLiteral, tt TT) {
		is := testutil.New(t)
		var _ Field = f
		buf := &strings.Builder{}
		var args []interface{}
		f.AppendSQLExclude("", buf, &args, make(map[string]int), tt.excludedTableQualifiers)
		is.Equal(tt.wantQuery, buf.String())
		is.Equal(tt.wantArgs, args)
		is.Equal("", f.GetAlias())
		is.Equal(string(f), f.GetName())
	}
	t.Run("FieldLiteral", func(t *testing.T) {
		tt := TT{wantQuery: "lorem ipsum dolor sit amet COUNT(*)"}
		assertField(t, FieldLiteral("lorem ipsum dolor sit amet COUNT(*)"), tt)
	})
}

func Test_Fields(t *testing.T) {
	type TT struct {
		excludedTableQualifiers []string
		wantQuery               string
		wantArgs                []interface{}
	}

	assertFields := func(t *testing.T, fs Fields, tt TT) {
		Is := testutil.New(t)
		buf := &strings.Builder{}
		var args []interface{}
		fs.AppendSQLExclude("", buf, &args, make(map[string]int), tt.excludedTableQualifiers)
		Is.Equal(tt.wantQuery, buf.String())
		Is.Equal(tt.wantArgs, args)
	}
	t.Run("empty", func(t *testing.T) {
		tt := TT{wantQuery: "", wantArgs: nil}
		assertFields(t, Fields{}, tt)
	})
	t.Run("Fields", func(t *testing.T) {
		u := NEW_USERS("u")
		tt := TT{wantQuery: "u.user_id, NULL, ?", wantArgs: []interface{}{456}}
		assertFields(t, Fields{u.USER_ID, nil, FieldValue(456)}, tt)
	})

	assertFieldsWithAlias := func(t *testing.T, fs Fields, tt TT) {
		Is := testutil.New(t)
		buf := &strings.Builder{}
		var args []interface{}
		fs.AppendSQLExcludeWithAlias("", buf, &args, make(map[string]int), tt.excludedTableQualifiers)
		Is.Equal(tt.wantQuery, buf.String())
		Is.Equal(tt.wantArgs, args)
	}
	t.Run("Fields with alias", func(t *testing.T) {
		u := NEW_USERS("u")
		tt := TT{wantQuery: "u.user_id AS uid, NULL, ? AS some_number", wantArgs: []interface{}{456}}
		assertFieldsWithAlias(t, Fields{u.USER_ID.As("uid"), nil, FieldValue(456).As("some_number")}, tt)
	})
}
