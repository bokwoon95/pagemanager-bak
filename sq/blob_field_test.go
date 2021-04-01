package sq

import (
	"strings"
	"testing"

	"github.com/bokwoon95/pagemanager/testutil"
)

func Test_BlobField(t *testing.T) {
	type TT struct {
		excludedTableQualifiers []string
		wantQuery               string
		wantArgs                []interface{}
	}

	assertField := func(t *testing.T, f BlobField, tt TT) {
		Is := testutil.New(t)
		var _ Field = f
		buf := &strings.Builder{}
		var args []interface{}
		_ = f.AppendSQLExclude("", buf, &args, make(map[string]int), tt.excludedTableQualifiers)
		Is.Equal(tt.wantQuery, buf.String())
		Is.Equal(f.alias, f.GetAlias())
		Is.Equal(f.name, f.GetName())
		if len(tt.excludedTableQualifiers) == 0 {
			Is.Equal(f.String(), buf.String())
		}
	}
	t.Run("BlobField", func(t *testing.T) {
		f := NewBlobField("my_field", TableInfo{Name: "my_table", Alias: "tbl"})
		tt := TT{wantQuery: "tbl.my_field"}
		assertField(t, f, tt)
	})
	t.Run("BlobField with alias", func(t *testing.T) {
		f := NewBlobField("my_field", TableInfo{Name: "my_table", Alias: "tbl"})
		tt := TT{wantQuery: "tbl.my_field"}
		assertField(t, f.As("dt"), tt)
	})
	t.Run("ASC", func(t *testing.T) {
		f := NewBlobField("my_field", TableInfo{Name: "my_table", Alias: "tbl"})
		tt := TT{wantQuery: "my_field ASC", excludedTableQualifiers: []string{"tbl"}}
		assertField(t, f.Asc(), tt)
	})
	t.Run("DESC", func(t *testing.T) {
		f := NewBlobField("my_field", TableInfo{Name: "my_table", Alias: "tbl"})
		tt := TT{wantQuery: "my_field DESC", excludedTableQualifiers: []string{"tbl"}}
		assertField(t, f.Desc(), tt)
	})
	t.Run("NULLS FIRST", func(t *testing.T) {
		f := NewBlobField("my_field", TableInfo{Name: "my_table", Alias: "tbl"})
		tt := TT{wantQuery: "my_field NULLS FIRST", excludedTableQualifiers: []string{"tbl"}}
		assertField(t, f.NullsFirst(), tt)
	})
	t.Run("NULLS LAST", func(t *testing.T) {
		f := NewBlobField("my_field", TableInfo{Name: "my_table", Alias: "tbl"})
		tt := TT{wantQuery: "my_field NULLS LAST", excludedTableQualifiers: []string{"tbl"}}
		assertField(t, f.NullsLast(), tt)
	})

	assertAssignment := func(t *testing.T, a Assignment, tt TT) {
		Is := testutil.New(t)
		buf := &strings.Builder{}
		var args []interface{}
		_ = a.AppendSQLExclude("", buf, &args, make(map[string]int), tt.excludedTableQualifiers)
		Is.Equal(tt.wantQuery, buf.String())
		Is.Equal(tt.wantArgs, args)
	}
	t.Run("SetBlob", func(t *testing.T) {
		data := []byte{'a', 'b', 'c', 'd'}
		f := NewBlobField("my_field", TableInfo{Name: "my_table", Alias: "tbl"})
		tt := TT{wantQuery: "my_field = ?", wantArgs: []interface{}{data}, excludedTableQualifiers: []string{"tbl"}}
		assertAssignment(t, f.SetBlob(data), tt)
	})
}
