package sq

import (
	"database/sql"
	"testing"

	"github.com/bokwoon95/pagemanager/testutil"
	_ "github.com/mattn/go-sqlite3"
)

func Test_EnsureTables(t *testing.T) {
	is := testutil.New(t)
	sqldb, err := sql.Open("sqlite3", ":memory:")
	is.NoErr(err)
	is.True(sqldb != nil)
	u, a := NEW_USERS("u"), NEW_APPLICATIONS("a")
	err = WithTx(sqldb, func(tx *sql.Tx) error {
		return EnsureTables(tx, "sqlite3", u, a)
	})
	is.NoErr(err)
	sqldb.Close()
}
