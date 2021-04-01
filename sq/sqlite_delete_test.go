package sq

import (
	"errors"
	"testing"

	"github.com/bokwoon95/pagemanager/testutil"
)

func TestDeleteQuery_ToSQL(t *testing.T) {
	type USERS struct {
		TableInfo
		USER_ID NumberField
		NAME    StringField
		EMAIL   StringField
		AGE     NumberField
	}
	type User struct {
		UserID int64
		Name   string
		Email  string
		Age    int
	}
	u := USERS{TableInfo: TableInfo{Schema: "db1", Alias: "u"}}
	ReflectTable(&u)

	assert := func(t *testing.T, q SQLiteDeleteQuery, wantQuery string, wantArgs []interface{}) {
		is := testutil.New(t, testutil.Parallel, testutil.FailFast)
		var _ Query = q
		gotQuery, gotArgs, _, err := q.ToSQL()
		is.NoErr(err)
		is.Equal("sqlite3", q.Dialect())
		is.Equal(wantQuery, gotQuery)
		is.Equal(wantArgs, gotArgs)
	}
	t.Run("empty", func(t *testing.T) {
		q := SQLiteDeleteQuery{}
		wantQuery := "DELETE FROM NULL"
		assert(t, q, wantQuery, nil)
	})
	t.Run("From", func(t *testing.T) {
		u := USERS{TableInfo: TableInfo{Schema: "db1"}}
		ReflectTable(&u)
		q := SQLite.DeleteFrom(u).Where(u.USER_ID.EqInt(1))
		wantQuery := "DELETE FROM db1.users WHERE users.user_id = ?"
		wantArgs := []interface{}{1}
		assert(t, q, wantQuery, wantArgs)
	})
	t.Run("misc", func(t *testing.T) {
		q := SQLite.DeleteFrom(u).Where(u.USER_ID.EqInt(1)).OrderBy(u.NAME, u.EMAIL.Desc()).Limit(-10).Offset(-5)
		wantQuery := "DELETE FROM db1.users AS u WHERE u.user_id = ? ORDER BY u.name, u.email DESC LIMIT ? OFFSET ?"
		wantArgs := []interface{}{1, int64(10), int64(5)}
		assert(t, q, wantQuery, wantArgs)
	})
	t.Run("CTE v1", func(t *testing.T) {
		cte1 := SQLite.Select(u.USER_ID, u.AGE).From(u).Where(u.USER_ID.EqInt(3)).CTE("cte1")
		cte2 := SQLite.Select(u.USER_ID.As("uid2"), u.AGE).From(u).Where(u.AGE.EqInt(5)).CTE("cte2")
		cte3 := SQLite.Select(u.NAME).From(u).Where(u.NAME.LikeString("bob%")).CTE("cte3")
		q := SQLite.WithDelete(cte1, cte2, cte3).
			DeleteFrom(u).
			Where(u.USER_ID.EqInt(1))
		wantQuery := "WITH cte1 AS (SELECT u.user_id, u.age FROM db1.users AS u WHERE u.user_id = ?)," +
			" cte2 AS (SELECT u.user_id AS uid2, u.age FROM db1.users AS u WHERE u.age = ?)," +
			" cte3 AS (SELECT u.name FROM db1.users AS u WHERE u.name LIKE ?)" +
			" DELETE FROM db1.users AS u" +
			" WHERE u.user_id = ?"
		wantArgs := []interface{}{3, 5, "bob%", 1}
		assert(t, q, wantQuery, wantArgs)
	})
	t.Run("CTE v2", func(t *testing.T) {
		cte1 := SQLite.Select(u.USER_ID, u.AGE).From(u).Where(u.USER_ID.EqInt(3)).CTE("cte1")
		cte2 := SQLite.Select(u.USER_ID.As("uid2"), u.AGE).From(u).Where(u.AGE.EqInt(5)).CTE("cte2")
		cte3 := SQLite.Select(u.NAME).From(u).Where(u.NAME.LikeString("bob%")).CTE("cte3")
		q := SQLite.DeleteFrom(u).
			Where(u.USER_ID.EqInt(1)).
			With(cte1, cte2, cte3)
		wantQuery := "WITH cte1 AS (SELECT u.user_id, u.age FROM db1.users AS u WHERE u.user_id = ?)," +
			" cte2 AS (SELECT u.user_id AS uid2, u.age FROM db1.users AS u WHERE u.age = ?)," +
			" cte3 AS (SELECT u.name FROM db1.users AS u WHERE u.name LIKE ?)" +
			" DELETE FROM db1.users AS u" +
			" WHERE u.user_id = ?"
		wantArgs := []interface{}{3, 5, "bob%", 1}
		assert(t, q, wantQuery, wantArgs)
	})

	t.Run("FetchableFields", func(t *testing.T) {
		// Get
		is := testutil.New(t, testutil.Parallel, testutil.FailFast)
		q := SQLite.DeleteFrom(u)
		_, err := q.GetFetchableFields()
		is.True(errors.Is(err, ErrUnsupported))
		// Set
		_, err = q.SetFetchableFields(Fields{})
		is.True(errors.Is(err, ErrUnsupported))
	})
}
