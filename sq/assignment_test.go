package sq

import (
	"strings"
	"testing"

	"github.com/bokwoon95/pagemanager/testutil"
)

func Test_Assignment(t *testing.T) {
	assert := func(t *testing.T, a Assignment, wantQuery string, wantArgs []interface{}) {
		Is := testutil.New(t)
		buf := &strings.Builder{}
		var args []interface{}
		_ = a.AppendSQLExclude("", buf, &args, make(map[string]int), nil)
		Is.Equal(wantQuery, buf.String())
		Is.Equal(wantArgs, args)
	}

	t.Run("field assign field", func(t *testing.T) {
		u := NEW_USERS("u")
		a := Assign(u.USER_ID, u.DISPLAYNAME)
		wantQuery := "u.user_id = u.displayname"
		assert(t, a, wantQuery, nil)
	})

	t.Run("field assign value", func(t *testing.T) {
		u := NEW_USERS("u")
		a := Assign(u.USER_ID, 5)
		wantQuery := "u.user_id = ?"
		wantArgs := []interface{}{5}
		assert(t, a, wantQuery, wantArgs)
	})
}

func Test_Assignments(t *testing.T) {
	assert := func(t *testing.T, as Assignments, wantQuery string, wantArgs []interface{}) {
		Is := testutil.New(t)
		buf := &strings.Builder{}
		var args []interface{}
		_ = as.AppendSQLExclude("", buf, &args, make(map[string]int), nil)
		Is.Equal(wantQuery, buf.String())
		Is.Equal(wantArgs, args)
	}

	t.Run("multiple assignments", func(t *testing.T) {
		u := NEW_USERS("u")
		as := Assignments{
			Assign(u.USER_ID, u.DISPLAYNAME),
			Assign(u.PASSWORD, "123456"),
			Assign(u.EMAIL, "bob@email.com"),
		}
		wantQuery := "u.user_id = u.displayname, u.password = ?, u.email = ?"
		wantArgs := []interface{}{"123456", "bob@email.com"}
		assert(t, as, wantQuery, wantArgs)
	})
}
