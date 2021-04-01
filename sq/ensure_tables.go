package sq

import (
	"database/sql"
	"fmt"
	"net/url"
	"reflect"
	"strings"

	"github.com/bokwoon95/erro"
)

func EnsureTables(db Queryer, dialect string, tables ...Table) error {
	var err error
	var tx *sql.Tx
	if txer, ok := db.(Transactor); ok {
		tx, err = txer.Begin()
		if err == nil {
			db = tx
			defer tx.Rollback()
		}
	}
	var tbls []htable
	for _, table := range tables {
		var T reflect.Value  // the non-pointer reflected value of the table struct
		var typ reflect.Type // reflect.Type of T
		T = reflect.ValueOf(table)
		typ = T.Type()
		if typ.Kind() == reflect.Ptr {
			ReflectTable(table)
		}
		T = reflect.Indirect(T)
		typ = T.Type()
		if typ.Kind() != reflect.Struct {
			return fmt.Errorf("not a struct")
		}
		tbl := htable{}
		for i := 0; i < T.NumField(); i++ {
			v := T.Field(i)
			t, ok := v.Interface().(TableInfo)
			if !ok {
				continue
			}
			field := typ.Field(i)
			if field.Name != "TableInfo" {
				continue
			}
			tbl.name = strings.ToLower(t.Name)
			fieldTag := field.Tag.Get("sq")
			m := parseFieldTag(fieldTag)
			if m.Get("name") != "" {
				tbl.name = m.Get("name")
			}
			if tbl.name == "" {
				tbl.name = strings.ToLower(typ.Name())
			}
			break
		}
		for i := 0; i < T.NumField(); i++ {
			col := hcolumn{}
			v := T.Field(i)
			fieldValue := v.Interface()
			_, ok := fieldValue.(Field)
			if !ok {
				continue
			}
			field := typ.Field(i)
			col.name = strings.ToLower(field.Name)
			fieldTag := field.Tag.Get("sq")
			m := parseFieldTag(fieldTag)
			if m.Get("name") != "" {
				col.name = m.Get("name")
			}
			switch fieldValue.(type) {
			case BlobField:
				col.typ = "BLOB"
			case BooleanField:
				col.typ = "BOOLEAN"
			case JSONField:
				col.typ = "JSON"
			case NumberField:
				col.typ = "INTEGER"
			case StringField:
				col.typ = "TEXT"
			case TimeField:
				col.typ = "DATETIME"
			}
			if m.Get("type") != "" {
				col.typ = m.Get("type")
			}
			for _, constraint := range m["misc"] {
				col.constraints = append(col.constraints, strings.ReplaceAll(constraint, "_", " "))
			}
			tbl.columns = append(tbl.columns, col)
		}
		tbls = append(tbls, tbl)
	}
	err = loadtables(db, tbls)
	if err != nil {
		return erro.Wrap(err)
	}
	if tx != nil {
		return tx.Commit()
	}
	return nil
}

func loadtables(db Queryer, tables []htable) error {
	var rows *sql.Rows
	var err error
	for _, table := range tables {
		// does table exist?
		var exists sql.NullBool
		rows, err = db.Query("SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE name = ?)", table.name)
		if err != nil {
			return erro.Wrap(err)
		}
		defer rows.Close()
		for rows.Next() {
			err = rows.Scan(&exists)
			if err != nil {
				return erro.Wrap(err)
			}
			break
		}
		// if not exists, create table from scratch and continue
		if !exists.Valid || !exists.Bool {
			_, err = db.Exec(table.ddl())
			if err != nil {
				return erro.Wrap(err)
			}
			continue
		}
		// do columns exist?
		columnset := make(map[string]struct{})
		rows, err = db.Query("SELECT name FROM pragma_table_info(?)", table.name)
		if err != nil {
			return erro.Wrap(err)
		}
		defer rows.Close()
		var name sql.NullString
		for rows.Next() {
			err = rows.Scan(&name)
			if err != nil {
				return erro.Wrap(err)
			}
			if name.Valid {
				columnset[name.String] = struct{}{}
			}
		}
		for _, column := range table.columns {
			if _, ok := columnset[column.name]; ok {
				continue
			}
			query := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table.name, column.name, column.typ)
			if len(column.constraints) > 0 {
				query = query + strings.Join(column.constraints, " ")
			}
			_, err = db.Exec(query)
			if err != nil {
				return erro.Wrap(err)
			}
		}
	}
	return nil
}

func ReflectTable(table Table) error {
	ptrvalue := reflect.ValueOf(table)
	typ := ptrvalue.Type()
	if typ.Kind() != reflect.Ptr {
		return fmt.Errorf("not a pointer")
	}
	value := reflect.Indirect(ptrvalue)
	typ = value.Type()
	if typ.Kind() != reflect.Struct {
		return fmt.Errorf("not a struct pointer")
	}
	var tableinfo TableInfo
	for i := 0; i < value.NumField(); i++ {
		v := value.Field(i)
		if !v.CanSet() {
			continue
		}
		t, ok := v.Interface().(TableInfo)
		if !ok {
			continue
		}
		field := typ.Field(i)
		if !field.Anonymous {
			continue
		}
		tableinfo.Schema = strings.ToLower(t.Schema)
		tableinfo.Name = strings.ToLower(t.Name)
		tableinfo.Alias = strings.ToLower(t.Alias)
		fieldTag := field.Tag.Get("sq")
		m := parseFieldTag(fieldTag)
		if m.Get("name") != "" {
			tableinfo.Name = m.Get("name")
		}
		if tableinfo.Name == "" {
			tableinfo.Name = strings.ToLower(typ.Name())
		}
		value.Field(i).Set(reflect.ValueOf(tableinfo))
		break
	}
	for i := 0; i < value.NumField(); i++ {
		v := value.Field(i)
		if !v.CanSet() {
			continue
		}
		fieldValue := v.Interface()
		_, ok := fieldValue.(Field)
		if !ok {
			continue
		}
		field := typ.Field(i)
		fieldName := strings.ToLower(field.Name)
		// fieldType := reflect.ValueOf(fieldValue).Type()
		fieldTag := field.Tag.Get("sq")
		// fmt.Printf("Name: %s,\t Value: %v,\t Type: %s,\t TagName: %s\n", fieldName, fieldValue, fieldType, fieldTag)
		m := parseFieldTag(fieldTag)
		if m.Get("name") != "" {
			fieldName = m.Get("name")
		}
		switch fieldValue.(type) {
		case BlobField:
			v.Set(reflect.ValueOf(NewBlobField(fieldName, tableinfo)))
		case BooleanField:
			v.Set(reflect.ValueOf(NewBooleanField(fieldName, tableinfo)))
		case JSONField:
			v.Set(reflect.ValueOf(NewJSONField(fieldName, tableinfo)))
		case NumberField:
			v.Set(reflect.ValueOf(NewNumberField(fieldName, tableinfo)))
		case StringField:
			v.Set(reflect.ValueOf(NewStringField(fieldName, tableinfo)))
		case TimeField:
			v.Set(reflect.ValueOf(NewTimeField(fieldName, tableinfo)))
		}
	}
	return nil
}

func parseFieldTag(fieldTag string) url.Values {
	fieldTag = strings.ReplaceAll(fieldTag, " ", "&")
	m, _ := url.ParseQuery(fieldTag)
	for k, v := range m {
		if len(v) != 1 {
			continue
		}
		if !strings.Contains(v[0], ",") {
			continue
		}
		m[k] = strings.Split(v[0], ",")
	}
	return m
}

type htable struct {
	name        string
	columns     []hcolumn
	constraints []string
}

type hcolumn struct {
	name        string
	typ         string
	constraints []string
}

func (t htable) ddl() string {
	buf := &strings.Builder{}
	buf.WriteString("CREATE TABLE ")
	buf.WriteString(t.name)
	buf.WriteString(" (")
	for i, c := range t.columns {
		buf.WriteString("\n    ")
		if i > 0 {
			buf.WriteString(",")
		}
		buf.WriteString(c.name)
		buf.WriteString(" ")
		buf.WriteString(c.typ)
		if len(c.constraints) > 0 {
			buf.WriteString(" ")
			buf.WriteString(strings.Join(c.constraints, " "))
		}
	}
	if len(t.constraints) > 0 {
		buf.WriteString("\n    ,")
		buf.WriteString(strings.Join(t.constraints, "\n    ,"))
	}
	buf.WriteString("\n)")
	return buf.String()
}

type typechangecallback func(db Queryer, table BaseTable, field Field, wantType, gotType string) error
