package sq

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/bokwoon95/erro"
	"github.com/bokwoon95/pagemanager/testutil"
)

type TBL struct {
	TableInfo
	ID         NumberField
	AGE        NumberField
	NAME       StringField
	EMAIL      StringField
	INFO       JSONField
	DATA       BlobField
	IS_ADMIN   BooleanField
	CREATED_AT TimeField
}

type Tb struct {
	ID        int64
	Age       int
	Name      string
	Email     string
	Info      json.RawMessage
	Data      []byte
	IsAdmin   bool
	CreatedAt time.Time
}

func Test_Fetch2(t *testing.T) {
	is := testutil.New(t)
	sqldb, err := sql.Open("sqlite3", ":memory:")
	is.NoErr(err)
	is.True(sqldb != nil)
	db := NewDB(sqldb, DefaultLogger(), Lverbose)

	// create table
	tbl := TBL{}
	err = WithTx(db, func(tx *sql.Tx) (err error) {
		err = EnsureTables(tx, "sqlite3", &tbl)
		if err != nil {
			return erro.Wrap(err)
		}
		return nil
	})
	is.NoErr(err)

	// insert data
	now := time.Now()
	data := []Tb{
		{
			ID: 1, Age: 15, Name: "a", Email: "a@email.com", Info: []byte(`{"id":1,"name":"a"}`),
			Data: []byte("\x00\x01\x02"), IsAdmin: true, CreatedAt: now.Add(1 * time.Hour),
		},
		{
			ID: 2, Age: 16, Name: "b", Email: "b@email.com", Info: []byte(`{"id":2,"name":"b"}`),
			Data: []byte("\x00\x01\x02"), IsAdmin: true, CreatedAt: now.Add(2 * time.Hour),
		},
		{
			ID: 3, Age: 17, Name: "c", Email: "c@email.com", Info: []byte(`{"id":3,"name":"c"}`),
			Data: []byte("\x00\x01\x02"), IsAdmin: true, CreatedAt: now.Add(3 * time.Hour),
		},
		{
			ID: 4, Age: 18, Name: "d", Email: "d@email.com", Info: []byte(`{"id":4,"name":"d"}`),
			Data: []byte("\x00\x01\x02"), IsAdmin: true, CreatedAt: now.Add(4 * time.Hour),
		},
		{
			ID: 5, Age: 19, Name: "e", Email: "e@email.com", Info: []byte(`{"id":5,"name":"e"}`),
			Data: []byte("\x00\x01\x02"), IsAdmin: false, CreatedAt: now.Add(5 * time.Hour),
		},
		{
			ID: 6, Age: 20, Name: "f", Email: "f@email.com", Info: []byte(`{"id":6,"name":"f"}`),
			Data: []byte("\x00\x01\x02"), IsAdmin: false, CreatedAt: now.Add(6 * time.Hour),
		},
		{
			ID: 7, Age: 21, Name: "g", Email: "g@email.com", Info: []byte(`{"id":7,"name":"g"}`),
			Data: []byte("\x00\x01\x02"), IsAdmin: false, CreatedAt: now.Add(7 * time.Hour),
		},
		{
			ID: 8, Age: 22, Name: "h", Email: "h@email.com", Info: []byte(`{"id":8,"name":"h"}`),
			Data: []byte("\x00\x01\x02"), IsAdmin: false, CreatedAt: now.Add(8 * time.Hour),
		},
	}
	rowsAffected, _, err := Exec(db, SQLite.InsertInto(tbl).Valuesx(func(col *Column) error {
		for _, tb := range data {
			col.SetInt64(tbl.ID, tb.ID)
			col.SetInt(tbl.AGE, tb.Age)
			col.SetString(tbl.NAME, tb.Name)
			col.Set(tbl.INFO, tb.Info)
			col.Set(tbl.DATA, tb.Data)
			col.SetBool(tbl.IS_ADMIN, tb.IsAdmin)
			col.SetTime(tbl.CREATED_AT, tb.CreatedAt)
		}
		return nil
	}), ErowsAffected)
	is.NoErr(err)
	is.Equal(int64(8), rowsAffected)

	t.Run("empty", func(t *testing.T) {
		var err error
		_, err = FetchContext(context.Background(), DB{}, nil, nil)
		is.True(err != nil)
		_, err = FetchContext(context.Background(), db, nil, nil)
		is.True(err != nil)
		_, _, err = ExecContext(context.Background(), DB{}, nil, 0)
		is.True(err != nil)
	})

	t.Run("basic select", func(t *testing.T) {
		is := testutil.New(t)
		var tbs []Tb
		rowCount, err := Fetch(db, SQLite.From(tbl).Where(tbl.ID.GtInt(2)), func(row *Row) error {
			tb := Tb{
				ID:        row.Int64(tbl.ID),
				Age:       row.Int(tbl.AGE),
				Name:      row.String(tbl.NAME),
				Email:     row.String(tbl.EMAIL),
				Data:      row.Bytes(tbl.DATA),
				IsAdmin:   row.Bool(tbl.IS_ADMIN),
				CreatedAt: row.Time(tbl.CREATED_AT),
			}
			var b []byte
			row.ScanInto(&b, tbl.INFO)
			tb.Info = b
			return row.Accumulate(func() error {
				tbs = append(tbs, tb)
				return nil
			})
		})
		is.NoErr(err)
		is.Equal(int64(6), rowCount)
	})

	t.Run("wrapScanError", func(t *testing.T) {
		is := testutil.New(t)
		rowCount, err := Fetch(sqldb, SQLite.From(tbl).Where(tbl.ID.GtInt(2)), func(row *Row) error {
			var id time.Time
			row.ScanInto(&id, tbl.ID)
			return nil
		})
		is.True(err != nil)
		is.Equal(int64(1), rowCount)
	})
}
