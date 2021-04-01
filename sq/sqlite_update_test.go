package sq

import (
	"errors"
	"testing"

	"github.com/bokwoon95/pagemanager/testutil"
)

func TestUpdateQuery_ToSQL(t *testing.T) {
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

	assert := func(t *testing.T, q SQLiteUpdateQuery, wantQuery string, wantArgs []interface{}) {
		is := testutil.New(t, testutil.Parallel, testutil.FailFast)
		var _ Query = q
		gotQuery, gotArgs, _, err := q.ToSQL()
		is.NoErr(err)
		is.Equal("sqlite3", q.Dialect())
		is.Equal(wantQuery, gotQuery)
		is.Equal(wantArgs, gotArgs)
	}
	t.Run("empty", func(t *testing.T) {
		q := SQLiteUpdateQuery{}
		wantQuery := "UPDATE NULL"
		assert(t, q, wantQuery, nil)
	})
	t.Run("Update", func(t *testing.T) {
		u := USERS{TableInfo: TableInfo{Schema: "db1"}}
		ReflectTable(&u)
		q := SQLite.Update(u).Set(u.USER_ID.SetInt(1))
		wantQuery := "UPDATE db1.users SET user_id = ?"
		wantArgs := []interface{}{1}
		assert(t, q, wantQuery, wantArgs)
	})
	t.Run("Joins", func(t *testing.T) {
		q := SQLite.Update(u).
			From(SQLite.From(u).SelectAll().Subquery("subquery")).
			Join(u, u.USER_ID.Eq(u.AGE)).
			LeftJoin(u, u.USER_ID.Eq(u.AGE)).
			RightJoin(u, u.USER_ID.Eq(u.AGE)).
			FullJoin(u, u.USER_ID.Eq(u.AGE)).
			CustomJoin("CROSS JOIN", u)
		wantQuery := "UPDATE db1.users AS u" +
			" FROM (SELECT * FROM db1.users AS u) AS subquery" +
			" JOIN db1.users AS u ON u.user_id = u.age" +
			" LEFT JOIN db1.users AS u ON u.user_id = u.age" +
			" RIGHT JOIN db1.users AS u ON u.user_id = u.age" +
			" FULL JOIN db1.users AS u ON u.user_id = u.age" +
			" CROSS JOIN db1.users AS u"
		assert(t, q, wantQuery, nil)
	})
	t.Run("Setx", func(t *testing.T) {
		user := User{UserID: 1, Name: "Bob", Email: "bob@email.com", Age: 24}
		q := SQLite.Update(u).
			Setx(func(col *Column) error {
				col.SetString(u.NAME, user.Name)
				col.SetString(u.EMAIL, user.Email)
				col.SetInt(u.AGE, user.Age)
				return nil
			}).
			Where(u.USER_ID.EqInt64(user.UserID))
		wantQuery := "UPDATE db1.users AS u" +
			" SET name = ?, email = ?, age = ?" +
			" WHERE u.user_id = ?"
		wantArgs := []interface{}{user.Name, user.Email, user.Age, user.UserID}
		assert(t, q, wantQuery, wantArgs)
	})
	t.Run("Setx RowValue", func(t *testing.T) {
		user := User{UserID: 1, Name: "Bob", Email: "bob@email.com", Age: 24}
		q := SQLite.Update(u).
			Setx(func(col *Column) error {
				col.Set(
					RowValue{u.NAME, u.EMAIL, u.AGE},
					RowValue{user.Name, user.Email, user.Age},
				)
				return nil
			}).
			Where(u.USER_ID.EqInt64(user.UserID))
		wantQuery := "UPDATE db1.users AS u" +
			" SET (name, email, age) = (?, ?, ?)" +
			" WHERE u.user_id = ?"
		wantArgs := []interface{}{user.Name, user.Email, user.Age, user.UserID}
		assert(t, q, wantQuery, wantArgs)
	})
	t.Run("Setx RowValue Select", func(t *testing.T) {
		user := User{UserID: 1, Name: "Bob", Email: "bob@email.com", Age: 24}
		uu := USERS{TableInfo: TableInfo{Schema: "db1", Alias: "uu"}}
		ReflectTable(&uu)
		q := SQLite.Update(u).
			Setx(func(col *Column) error {
				col.Set(
					RowValue{u.NAME, u.EMAIL, u.AGE},
					SQLite.Select(uu.NAME, uu.EMAIL, uu.AGE).From(uu).Where(uu.USER_ID.EqInt(99)).Limit(1),
				)
				return nil
			}).
			Where(u.USER_ID.EqInt64(user.UserID))
		wantQuery := "UPDATE db1.users AS u" +
			" SET (name, email, age) =" +
			" (SELECT uu.name, uu.email, uu.age FROM db1.users AS uu WHERE uu.user_id = ? LIMIT ?)" +
			" WHERE u.user_id = ?"
		wantArgs := []interface{}{99, int64(1), user.UserID}
		assert(t, q, wantQuery, wantArgs)
	})
	t.Run("CTE v1", func(t *testing.T) {
		cte1 := SQLite.Select(u.USER_ID, u.AGE).From(u).Where(u.USER_ID.EqInt(3)).CTE("cte1")
		cte2 := SQLite.Select(u.USER_ID.As("uid2"), u.AGE).From(u).Where(u.AGE.EqInt(5)).CTE("cte2")
		cte3 := SQLite.Select(u.NAME).From(u).Where(u.NAME.LikeString("bob%")).CTE("cte3")
		q := SQLite.WithUpdate(cte1, cte2, cte3).
			Update(u).
			Set(u.USER_ID.SetInt(1)).
			From(cte1).Join(cte2, cte2["uid2"].Eq(cte1["user_id"])).
			Where(u.USER_ID.EqInt(1))
		wantQuery := "WITH cte1 AS (SELECT u.user_id, u.age FROM db1.users AS u WHERE u.user_id = ?)," +
			" cte2 AS (SELECT u.user_id AS uid2, u.age FROM db1.users AS u WHERE u.age = ?)," +
			" cte3 AS (SELECT u.name FROM db1.users AS u WHERE u.name LIKE ?)" +
			" UPDATE db1.users AS u" +
			" SET user_id = ?" +
			" FROM cte1 JOIN cte2 ON cte2.uid2 = cte1.user_id" +
			" WHERE u.user_id = ?"
		wantArgs := []interface{}{3, 5, "bob%", 1, 1}
		assert(t, q, wantQuery, wantArgs)
	})
	t.Run("CTE v2", func(t *testing.T) {
		cte1 := SQLite.Select(u.USER_ID, u.AGE).From(u).Where(u.USER_ID.EqInt(3)).CTE("cte1")
		cte2 := SQLite.Select(u.USER_ID.As("uid2"), u.AGE).From(u).Where(u.AGE.EqInt(5)).CTE("cte2")
		cte3 := SQLite.Select(u.NAME).From(u).Where(u.NAME.LikeString("bob%")).CTE("cte3")
		q := SQLite.Update(u).
			Set(u.USER_ID.SetInt(1)).
			From(cte1).Join(cte2, cte2["uid2"].Eq(cte1["user_id"])).
			Where(u.USER_ID.EqInt(1)).
			With(cte1, cte2, cte3)
		wantQuery := "WITH cte1 AS (SELECT u.user_id, u.age FROM db1.users AS u WHERE u.user_id = ?)," +
			" cte2 AS (SELECT u.user_id AS uid2, u.age FROM db1.users AS u WHERE u.age = ?)," +
			" cte3 AS (SELECT u.name FROM db1.users AS u WHERE u.name LIKE ?)" +
			" UPDATE db1.users AS u" +
			" SET user_id = ?" +
			" FROM cte1 JOIN cte2 ON cte2.uid2 = cte1.user_id" +
			" WHERE u.user_id = ?"
		wantArgs := []interface{}{3, 5, "bob%", 1, 1}
		assert(t, q, wantQuery, wantArgs)
	})
	t.Run("dual writing columns", func(t *testing.T) {
		user := User{UserID: 1, Name: "Bob"}
		mapper := func(userID int64, name string) func(*Column) error {
			return func(col *Column) error {
				col.SetInt64(u.USER_ID, userID)
				col.SetInt(u.AGE, int(userID)) // dual write u.user_id to u.age
				col.SetString(u.NAME, name)
				col.SetString(u.EMAIL, name) // dual write u.name into u.email
				return nil
			}
		}
		q := SQLite.Update(u).Setx(mapper(user.UserID, user.Name)).Where(u.USER_ID.EqInt64(user.UserID))
		wantQuery := "UPDATE db1.users AS u" +
			" SET user_id = ?, age = ?, name = ?, email = ?" +
			" WHERE u.user_id = ?"
		wantArgs := []interface{}{user.UserID, int(user.UserID), user.Name, user.Name, user.UserID}
		assert(t, q, wantQuery, wantArgs)
	})

	t.Run("returning domain error in columnMapper", func(t *testing.T) {
		var ErrUnderage = errors.New("too young to drive")
		isUnderage := func(age int) bool { return age < 18 }
		user := User{Name: "Bob", Email: "bob@email.com", Age: 17}
		q := SQLite.Update(u).Setx(func(col *Column) error {
			if isUnderage(user.Age) {
				return ErrUnderage
			}
			col.SetString(u.NAME, user.Name)
			col.SetString(u.EMAIL, user.Email)
			col.SetInt(u.AGE, user.Age)
			return nil
		})
		is := testutil.New(t, testutil.Parallel, testutil.FailFast)
		_, _, _, err := q.ToSQL()
		is.True(errors.Is(err, ErrUnderage))
	})
	t.Run("FetchableFields", func(t *testing.T) {
		// Get
		is := testutil.New(t, testutil.Parallel, testutil.FailFast)
		q := SQLite.Update(u)
		_, err := q.GetFetchableFields()
		is.True(errors.Is(err, ErrUnsupported))
		// Set
		_, err = q.SetFetchableFields(Fields{})
		is.True(errors.Is(err, ErrUnsupported))
	})
}
