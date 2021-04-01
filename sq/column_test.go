package sq

import (
	"strings"
	"testing"
	"time"

	"github.com/bokwoon95/pagemanager/testutil"
)

func TestColumnInsert(t *testing.T) {
	is := testutil.New(t)
	type User struct {
		UserID      int
		DisplayName string
		Email       string
		Password    string
	}
	users := []User{
		{
			UserID:      1,
			DisplayName: "one",
			Email:       "one",
			Password:    "one",
		},
		{
			UserID:      2,
			DisplayName: "two",
			Email:       "two",
			Password:    "two",
		},
		{
			UserID:      3,
			DisplayName: "three",
			Email:       "three",
			Password:    "three",
		},
	}
	col := &Column{mode: ColumnModeInsert}
	u := NEW_USERS("u")
	for _, user := range users {
		col.Set(u.USER_ID, user.UserID)
		col.Set(u.DISPLAYNAME, user.DisplayName)
		col.Set(u.EMAIL, user.Email)
		col.Set(u.PASSWORD, user.Password)
		col.Set(nil, 999) // this should have no effect on the result
	}
	is.Equal(Fields{u.USER_ID, u.DISPLAYNAME, u.EMAIL, u.PASSWORD}, col.insertColumns)
	is.Equal(
		RowValues{
			{users[0].UserID, users[0].DisplayName, users[0].Email, users[0].Password},
			{users[1].UserID, users[1].DisplayName, users[1].Email, users[1].Password},
			{users[2].UserID, users[2].DisplayName, users[2].Email, users[2].Password},
		},
		col.rowValues,
	)
}

func TestColumnUpdate(t *testing.T) {
	is := testutil.New(t)
	type User struct {
		UserID      int
		DisplayName string
		Email       string
		Password    string
	}
	col := &Column{mode: ColumnModeUpdate}
	a := NEW_APPLICATIONS("a")
	now := time.Now()
	col.SetBool(a.SUBMITTED, true)
	col.SetFloat64(a.APPLICATION_ID, 1.0)
	col.SetInt(a.APPLICATION_ID, 1)
	col.SetInt64(a.APPLICATION_ID, 1)
	col.SetString(a.TEAM_NAME, "lorem ipsum")
	col.SetTime(a.CREATED_AT, now)
	buf1, buf2 := &strings.Builder{}, &strings.Builder{}
	var args1, args2 []interface{}
	_ = Assignments{
		a.SUBMITTED.SetBool(true),
		a.APPLICATION_ID.SetFloat64(1.0),
		a.APPLICATION_ID.SetInt(1),
		a.APPLICATION_ID.SetInt64(1),
		a.TEAM_NAME.SetString("lorem ipsum"),
		a.CREATED_AT.SetTime(now),
	}.AppendSQLExclude("", buf1, &args1, make(map[string]int), nil)
	_ = col.assignments.AppendSQLExclude("", buf2, &args2, make(map[string]int), nil)
	is.Equal(buf1.String(), buf2.String())
	is.Equal(args1, args2)
}
