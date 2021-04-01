package sq

import (
	"errors"
	"testing"

	"github.com/bokwoon95/pagemanager/testutil"
)

func TestInsertQuery_ToSQL(t *testing.T) {
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

	assert := func(t *testing.T, q SQLiteInsertQuery, wantQuery string, wantArgs []interface{}) {
		is := testutil.New(t, testutil.Parallel, testutil.FailFast)
		var _ Query = q
		gotQuery, gotArgs, _, err := q.ToSQL()
		is.NoErr(err)
		is.Equal("sqlite3", q.Dialect())
		is.Equal(wantQuery, gotQuery)
		is.Equal(wantArgs, gotArgs)
	}
	t.Run("empty", func(t *testing.T) {
		q := SQLiteInsertQuery{}
		wantQuery := "INSERT INTO NULL"
		assert(t, q, wantQuery, nil)
	})
	t.Run("INSERT INTO", func(t *testing.T) {
		q := SQLite.InsertInto(u)
		wantQuery := "INSERT INTO db1.users AS u"
		assert(t, q, wantQuery, nil)
	})
	t.Run("VALUES", func(t *testing.T) {
		q := SQLite.InsertInto(u).
			Columns(u.USER_ID, u.NAME, u.EMAIL, u.AGE).
			Values(1, "a", "a@email.com", 11).
			Values(2, "b", "b@email.com", 22).
			Values(3, "c", "c@email.com", 33).
			OnConflict().DoNothing()
		wantQuery := "INSERT INTO db1.users AS u" +
			" (user_id, name, email, age)" +
			" VALUES" +
			" (?, ?, ?, ?)," +
			" (?, ?, ?, ?)," +
			" (?, ?, ?, ?)" +
			" ON CONFLICT DO NOTHING"
		wantArgs := []interface{}{
			1, "a", "a@email.com", 11,
			2, "b", "b@email.com", 22,
			3, "c", "c@email.com", 33,
		}
		assert(t, q, wantQuery, wantArgs)
	})
	t.Run("INSERT SELECT", func(t *testing.T) {
		q := SQLite.InsertInto(u).
			Columns(u.USER_ID, u.NAME, u.EMAIL).
			Select(SQLite.Select(u.USER_ID, u.NAME, u.EMAIL).
				From(u).
				Where(u.AGE.GtInt(30)),
			).
			OnConflict(u.USER_ID, u.NAME).
			Where(u.NAME.IsNotNull()).
			DoUpdateSet(
				SetExcluded(u.USER_ID),
				SetExcluded(u.NAME),
				SetExcluded(u.EMAIL),
			).
			Where(u.EMAIL.IsNotNull())
		wantQuery := "INSERT INTO db1.users AS u" +
			" (user_id, name, email)" +
			" SELECT u.user_id, u.name, u.email FROM db1.users AS u WHERE u.age > ?" +
			" ON CONFLICT (user_id, name)" +
			" WHERE name IS NOT NULL" +
			" DO UPDATE SET user_id = EXCLUDED.user_id, name = EXCLUDED.name, email = EXCLUDED.email" +
			" WHERE u.email IS NOT NULL"
		wantArgs := []interface{}{30}
		assert(t, q, wantQuery, wantArgs)
	})
	t.Run("aliasless table", func(t *testing.T) {
		u := USERS{TableInfo: TableInfo{Schema: "db1"}}
		ReflectTable(&u)
		q := SQLite.InsertInto(u).Columns(u.USER_ID, u.NAME, u.EMAIL)
		wantQuery := "INSERT INTO db1.users (user_id, name, email)"
		assert(t, q, wantQuery, nil)
	})
	t.Run("Valuesx, one entry", func(t *testing.T) {
		user := User{Name: "Bob", Email: "bob@email.com", Age: 22}
		q := SQLite.InsertInto(u).Valuesx(func(col *Column) error {
			col.SetString(u.NAME, user.Name)
			col.SetString(u.EMAIL, user.Email)
			col.SetInt(u.AGE, user.Age)
			return nil
		})
		wantQuery := "INSERT INTO db1.users AS u (name, email, age) VALUES (?, ?, ?)"
		wantArgs := []interface{}{user.Name, user.Email, user.Age}
		assert(t, q, wantQuery, wantArgs)
	})
	t.Run("Valuesx, multiple entries", func(t *testing.T) {
		users := []User{
			{Name: "Bob", Email: "bob@email.com", Age: 22},
			{Name: "Alice", Email: "alice@email.com", Age: 23},
			{Name: "Tom", Email: "tom@email.com", Age: 24},
			{Name: "Jerry", Email: "jerry@email.com", Age: 25},
		}
		q := SQLite.InsertInto(u).Valuesx(func(col *Column) error {
			for _, user := range users {
				col.SetString(u.NAME, user.Name)
				col.SetString(u.EMAIL, user.Email)
				col.SetInt(u.AGE, user.Age)
			}
			return nil
		})
		wantQuery := "INSERT INTO db1.users AS u (name, email, age)" +
			" VALUES (?, ?, ?), (?, ?, ?), (?, ?, ?), (?, ?, ?)"
		wantArgs := []interface{}{
			users[0].Name, users[0].Email, users[0].Age,
			users[1].Name, users[1].Email, users[1].Age,
			users[2].Name, users[2].Email, users[2].Age,
			users[3].Name, users[3].Email, users[3].Age,
		}
		assert(t, q, wantQuery, wantArgs)
	})
	t.Run("CTE v1", func(t *testing.T) {
		cte1 := SQLite.Select(u.USER_ID, u.AGE).From(u).Where(u.USER_ID.EqInt(3)).CTE("cte1")
		cte2 := SQLite.Select(u.USER_ID.As("uid2"), u.AGE).From(u).Where(u.AGE.EqInt(5)).CTE("cte2")
		cte3 := SQLite.Select(u.NAME).From(u).Where(u.NAME.LikeString("bob%")).CTE("cte3")
		q := SQLite.WithInsert(cte1, cte2, cte3).
			InsertInto(u).
			Columns(u.USER_ID, u.AGE).
			Select(SQLite.Select(cte1["user_id"], cte1["age"]).From(cte1))
		wantQuery := "WITH cte1 AS (SELECT u.user_id, u.age FROM db1.users AS u WHERE u.user_id = ?)," +
			" cte2 AS (SELECT u.user_id AS uid2, u.age FROM db1.users AS u WHERE u.age = ?)," +
			" cte3 AS (SELECT u.name FROM db1.users AS u WHERE u.name LIKE ?)" +
			" INSERT INTO db1.users AS u (user_id, age)" +
			" SELECT cte1.user_id, cte1.age FROM cte1"
		wantArgs := []interface{}{3, 5, "bob%"}
		assert(t, q, wantQuery, wantArgs)
	})
	t.Run("CTE v2", func(t *testing.T) {
		cte1 := SQLite.Select(u.USER_ID, u.AGE).From(u).Where(u.USER_ID.EqInt(3)).CTE("cte1")
		cte2 := SQLite.Select(u.USER_ID.As("uid2"), u.AGE).From(u).Where(u.AGE.EqInt(5)).CTE("cte2")
		cte3 := SQLite.Select(u.NAME).From(u).Where(u.NAME.LikeString("bob%")).CTE("cte3")
		q := SQLite.InsertInto(u).
			Columns(u.USER_ID, u.AGE).
			Select(SQLite.Select(cte1["user_id"], cte1["age"]).From(cte1)).
			With(cte1, cte2, cte3)
		wantQuery := "WITH cte1 AS (SELECT u.user_id, u.age FROM db1.users AS u WHERE u.user_id = ?)," +
			" cte2 AS (SELECT u.user_id AS uid2, u.age FROM db1.users AS u WHERE u.age = ?)," +
			" cte3 AS (SELECT u.name FROM db1.users AS u WHERE u.name LIKE ?)" +
			" INSERT INTO db1.users AS u (user_id, age)" +
			" SELECT cte1.user_id, cte1.age FROM cte1"
		wantArgs := []interface{}{3, 5, "bob%"}
		assert(t, q, wantQuery, wantArgs)
	})

	t.Run("returning domain error in columnMapper", func(t *testing.T) {
		var ErrUnderage = errors.New("too young to drive")
		isUnderage := func(age int) bool { return age < 18 }
		users := []User{
			{Name: "Bob", Email: "bob@email.com", Age: 20},
			{Name: "Alice", Email: "alice@email.com", Age: 19},
			{Name: "Tom", Email: "tom@email.com", Age: 18},
			{Name: "Jerry", Email: "jerry@email.com", Age: 17},
		}
		q := SQLite.InsertInto(u).Valuesx(func(col *Column) error {
			for _, user := range users {
				if isUnderage(user.Age) {
					return ErrUnderage
				}
				col.SetString(u.NAME, user.Name)
				col.SetString(u.EMAIL, user.Email)
				col.SetInt(u.AGE, user.Age)
			}
			return nil
		})
		is := testutil.New(t, testutil.Parallel, testutil.FailFast)
		_, _, _, err := q.ToSQL()
		is.True(errors.Is(err, ErrUnderage))
	})
	t.Run("FetchableFields", func(t *testing.T) {
		// Get
		is := testutil.New(t, testutil.Parallel, testutil.FailFast)
		q := SQLite.InsertInto(u)
		_, err := q.GetFetchableFields()
		is.True(errors.Is(err, ErrUnsupported))
		// Set
		_, err = q.SetFetchableFields(Fields{})
		is.True(errors.Is(err, ErrUnsupported))
	})
}
