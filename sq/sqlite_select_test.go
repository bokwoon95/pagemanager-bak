package sq

import (
	"testing"

	"github.com/bokwoon95/pagemanager/testutil"
)

func TestSelectQuery_ToSQL(t *testing.T) {
	type USERS struct {
		TableInfo
		USER_ID NumberField
		NAME    StringField
		EMAIL   StringField
		AGE     NumberField
	}
	u := USERS{TableInfo: TableInfo{Schema: "db1", Alias: "u"}}
	ReflectTable(&u)

	assert := func(t *testing.T, q SQLiteSelectQuery, wantQuery string, wantArgs []interface{}) {
		is := testutil.New(t, testutil.Parallel, testutil.FailFast)
		var _ Query = q
		gotQuery, gotArgs, _, err := q.ToSQL()
		is.NoErr(err)
		is.Equal("sqlite3", q.Dialect())
		is.Equal(wantQuery, gotQuery)
		is.Equal(wantArgs, gotArgs)
	}
	t.Run("empty", func(t *testing.T) {
		q := SQLiteSelectQuery{}
		wantQuery := "SELECT 1"
		assert(t, q, wantQuery, nil)
	})
	t.Run("FROM", func(t *testing.T) {
		q := SQLite.From(u)
		wantQuery := "SELECT 1 FROM db1.users AS u"
		assert(t, q, wantQuery, nil)
	})
	t.Run("SelectOne", func(t *testing.T) {
		q := SQLite.SelectOne().From(u)
		wantQuery := "SELECT 1 FROM db1.users AS u"
		assert(t, q, wantQuery, nil)
	})
	t.Run("SelectDistinct", func(t *testing.T) {
		q := SQLite.SelectDistinct(u.USER_ID).From(u)
		wantQuery := "SELECT DISTINCT u.user_id FROM db1.users AS u"
		assert(t, q, wantQuery, nil)
	})
	t.Run("Joins", func(t *testing.T) {
		q := SQLite.Select().SelectDistinct().SelectOne().SelectAll().SelectCount().Select().
			From(SQLite.From(u).SelectAll().Subquery("subquery")).
			Join(u, u.USER_ID.Eq(u.AGE)).
			LeftJoin(u, u.USER_ID.Eq(u.AGE)).
			RightJoin(u, u.USER_ID.Eq(u.AGE)).
			FullJoin(u, u.USER_ID.Eq(u.AGE)).
			CustomJoin("CROSS JOIN", u)
		wantQuery := "SELECT DISTINCT COUNT(*)" +
			" FROM (SELECT * FROM db1.users AS u) AS subquery" +
			" JOIN db1.users AS u ON u.user_id = u.age" +
			" LEFT JOIN db1.users AS u ON u.user_id = u.age" +
			" RIGHT JOIN db1.users AS u ON u.user_id = u.age" +
			" FULL JOIN db1.users AS u ON u.user_id = u.age" +
			" CROSS JOIN db1.users AS u"
		assert(t, q, wantQuery, nil)
	})
	t.Run("Misc", func(t *testing.T) {
		w := PartitionBy(u.EMAIL).OrderBy(u.USER_ID).As("w")
		q := SQLite.Select(SumOver(u.NAME, w.Name())).
			From(u).
			GroupBy(u.NAME).
			Having(u.NAME.NeString("alice")).
			Window(w).
			OrderBy(u.AGE, u.USER_ID.Desc()).
			Limit(-10).
			Offset(-5)
		wantQuery := "SELECT SUM(u.name) OVER w" +
			" FROM db1.users AS u" +
			" GROUP BY u.name" +
			" HAVING u.name <> ?" +
			" WINDOW w AS (PARTITION BY u.email ORDER BY u.user_id)" +
			" ORDER BY u.age, u.user_id DESC" +
			" LIMIT ?" +
			" OFFSET ?"
		wantArgs := []interface{}{"alice", int64(10), int64(5)}
		assert(t, q, wantQuery, wantArgs)
	})
	t.Run("CTE v1", func(t *testing.T) {
		cte1 := SQLite.Select(u.USER_ID, u.AGE).From(u).Where(u.USER_ID.EqInt(3)).CTE("cte1")
		cte2 := SQLite.Select(u.USER_ID.As("uid2"), u.AGE).From(u).Where(u.AGE.EqInt(5)).CTE("cte2")
		cte3 := SQLite.Select(u.NAME).From(u).Where(u.NAME.LikeString("bob%")).CTE("cte3")
		q := SQLite.WithSelect(cte1, cte2, cte3).Select(cte2["uid2"]).From(cte1).Join(cte2, cte2["age"].Eq(cte1["age"]))
		wantQuery := "WITH cte1 AS (SELECT u.user_id, u.age FROM db1.users AS u WHERE u.user_id = ?)," +
			" cte2 AS (SELECT u.user_id AS uid2, u.age FROM db1.users AS u WHERE u.age = ?)," +
			" cte3 AS (SELECT u.name FROM db1.users AS u WHERE u.name LIKE ?)" +
			" SELECT cte2.uid2 FROM cte1 JOIN cte2 ON cte2.age = cte1.age"
		wantArgs := []interface{}{3, 5, "bob%"}
		assert(t, q, wantQuery, wantArgs)
	})
	t.Run("CTE v2", func(t *testing.T) {
		cte1 := SQLite.Select(u.USER_ID, u.AGE).From(u).Where(u.USER_ID.EqInt(3)).CTE("cte1")
		cte2 := SQLite.Select(u.USER_ID.As("uid2"), u.AGE).From(u).Where(u.AGE.EqInt(5)).CTE("cte2")
		cte3 := SQLite.Select(u.NAME).From(u).Where(u.NAME.LikeString("bob%")).CTE("cte3")
		q := SQLite.Select(cte2["uid2"]).From(cte1).Join(cte2, cte2["age"].Eq(cte1["age"])).With(cte1, cte2, cte3)
		wantQuery := "WITH cte1 AS (SELECT u.user_id, u.age FROM db1.users AS u WHERE u.user_id = ?)," +
			" cte2 AS (SELECT u.user_id AS uid2, u.age FROM db1.users AS u WHERE u.age = ?)," +
			" cte3 AS (SELECT u.name FROM db1.users AS u WHERE u.name LIKE ?)" +
			" SELECT cte2.uid2 FROM cte1 JOIN cte2 ON cte2.age = cte1.age"
		wantArgs := []interface{}{3, 5, "bob%"}
		assert(t, q, wantQuery, wantArgs)
	})

	t.Run("FetchableFields", func(t *testing.T) {
		// Get
		is := testutil.New(t, testutil.Parallel, testutil.FailFast)
		q := SQLite.Select(u.USER_ID).From(u)
		fields, err := q.GetFetchableFields()
		is.NoErr(err)
		is.Equal([]Field{u.USER_ID}, fields)
		// Set
		fields = []Field{u.AGE, u.NAME, u.EMAIL}
		qq, err := SQLite.Select(u.USER_ID).From(u).SetFetchableFields(fields)
		is.NoErr(err)
		query, args, _, err := qq.ToSQL()
		is.NoErr(err)
		is.Equal("SELECT u.age, u.name, u.email FROM db1.users AS u", query)
		is.Equal(([]interface{})(nil), args)
	})
}
