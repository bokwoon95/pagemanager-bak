package pagemanager

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"html/template"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/bokwoon95/erro"
	"github.com/bokwoon95/pagemanager/sq"
	"github.com/bokwoon95/pagemanager/tables"
)

type PageData struct {
	Ctx        context.Context
	URL        string
	DataID     string
	LocaleCode string
	EditMode   int
	CSSList    []string
	JSList     []string
	CSPList    map[string][]string
}

func (pg PageData) CSS() template.HTML {
	buf := bufpool.Get().(*strings.Builder)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	for _, name := range pg.CSSList {
		if buf.Len() > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString(`<link rel="stylesheet" type="text/css" href="`)
		buf.WriteString(name)
		buf.WriteString(`">`)
	}
	return template.HTML(buf.String())
}

func (pg PageData) JS() template.HTML {
	buf := bufpool.Get().(*strings.Builder)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	for _, name := range pg.JSList {
		if buf.Len() > 0 {
			buf.WriteString("\n")
		}
		buf.WriteString(`<script src="`)
		buf.WriteString(name)
		buf.WriteString(`"></script>`)
	}
	return template.HTML(buf.String())
}

func (pg PageData) ContentSecurityPolicy() template.HTML {
	return template.HTML(fmt.Sprint(pg.CSPList))
}

type PageDataOption func(*PageData)

func pmLocale(localeCode string) PageDataOption {
	return func(pg *PageData) { pg.LocaleCode = localeCode }
}

func pmDataID(dataID string) PageDataOption {
	return func(pg *PageData) { pg.DataID = dataID }
}

type NullString struct {
	Valid bool
	Str   string
}

// Scan implements the Scanner interface.
func (ns *NullString) Scan(value interface{}) error {
	if value == nil {
		ns.Str, ns.Valid = "", false
		return nil
	}
	ns.Str = asString(value)
	ns.Valid = true
	return nil
}

func asString(src interface{}) string {
	switch v := src.(type) {
	case fmt.Stringer:
		return v.String()
	case string:
		return v
	case []byte:
		return string(v)
	case time.Time:
		return v.Format(time.RFC3339Nano)
	}
	rv := reflect.ValueOf(src)
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10)
	case reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 64)
	case reflect.Float32:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 32)
	case reflect.Bool:
		return strconv.FormatBool(rv.Bool())
	}
	return fmt.Sprintf("%v", src)
}

// Value implements the driver Valuer interface.
func (ns NullString) Value() (driver.Value, error) {
	if !ns.Valid {
		return nil, nil
	}
	return ns.String, nil
}

func (ns NullString) String() string {
	return ns.Str
}

func safeHTML(v interface{}) template.HTML {
	return template.HTML(asString(v))
}

func (pm *PageManager) pmGetValue(pg PageData, key string, opts ...PageDataOption) (NullString, error) {
	for _, opt := range opts {
		opt(&pg)
	}
	var ns NullString
	PAGEDATA := tables.NEW_PAGEDATA(pg.Ctx, "p")
	_, err := sq.FetchContext(pg.Ctx, pm.dataDB, sq.SQLite.
		From(PAGEDATA).
		Where(
			PAGEDATA.LOCALE_CODE.In([]string{pg.LocaleCode, ""}),
			PAGEDATA.DATA_ID.EqString(pg.DataID),
			PAGEDATA.KEY.EqString(key),
			PAGEDATA.ARRAY_INDEX.IsNull(),
		).
		OrderBy(sq.
			Case(PAGEDATA.LOCALE_CODE).
			When(pg.LocaleCode, 1).
			When("", 2),
		).
		Limit(1),
		func(row *sq.Row) error {
			row.ScanInto(&ns, PAGEDATA.VALUE)
			return nil
		},
	)
	if err != nil {
		return ns, erro.Wrap(err)
	}
	return ns, nil
}

func (pm *PageManager) pmGetRows(pg PageData, key string, opts ...PageDataOption) ([]interface{}, error) {
	for _, opt := range opts {
		opt(&pg)
	}
	PAGEDATA := tables.NEW_PAGEDATA(pg.Ctx, "p")
	exists, err := sq.ExistsContext(pg.Ctx, pm.dataDB, sq.SQLite.
		From(PAGEDATA).
		Where(
			PAGEDATA.LOCALE_CODE.EqString(pg.LocaleCode),
			PAGEDATA.DATA_ID.EqString(pg.DataID),
			PAGEDATA.KEY.EqString(key),
			PAGEDATA.ARRAY_INDEX.IsNotNull(),
		),
	)
	localeCode := pg.LocaleCode
	if !exists {
		localeCode = "" // default locale code
	}
	var values []interface{}
	var b []byte
	_, err = sq.FetchContext(pg.Ctx, pm.dataDB, sq.SQLite.
		From(PAGEDATA).
		Where(
			PAGEDATA.LOCALE_CODE.EqString(localeCode),
			PAGEDATA.DATA_ID.EqString(pg.DataID),
			PAGEDATA.KEY.EqString(key),
			PAGEDATA.ARRAY_INDEX.IsNotNull(),
		).
		OrderBy(PAGEDATA.ARRAY_INDEX),
		func(row *sq.Row) error {
			b = row.Bytes(PAGEDATA.VALUE)
			if row.Count() == 0 {
				return nil
			}
			value := make(map[string]interface{})
			err := json.Unmarshal(b, &value)
			if err != nil {
				values = append(values, string(b)) // couldn't unmarshal json, switching to string
			} else {
				values = append(values, value)
			}
			return nil
		},
	)
	if err != nil {
		return values, erro.Wrap(err)
	}
	return values, nil
}

func (pm *PageManager) funcmap() map[string]interface{} {
	return map[string]interface{}{
		"safeHTML":   safeHTML,
		"pmGetValue": pm.pmGetValue,
		"pmGetRows":  pm.pmGetRows,
		"pmLocale":   pmLocale,
		"pmDataID":   pmDataID,
	}
}
