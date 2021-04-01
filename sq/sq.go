package sq

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"sync"
)

var ErrUnsupported = errors.New("unsupported operation")
var SkipRows = errors.New("skip subsequent rows")

type Table interface {
	AppendSQL(dialect string, buf *strings.Builder, args *[]interface{}, params map[string]int) error
	GetAlias() string
	GetName() string // Table name must exclude the schema (if any)
}

func getAliasOrName(val interface {
	GetAlias() string
	GetName() string
}) string {
	s := val.GetAlias()
	if s == "" {
		s = val.GetName()
	}
	return s
}

type BaseTable interface {
	Table
	GetSchema() string
}

type Query interface {
	AppendSQL(dialect string, buf *strings.Builder, args *[]interface{}, params map[string]int) error
	ToSQL() (query string, args []interface{}, params map[string]int, err error)
	SetFetchableFields([]Field) (Query, error)
	GetFetchableFields() ([]Field, error)
	Dialect() string
}

type Field interface {
	// Fields should respect the excludedTableQualifiers argument in ToSQL().
	// E.g. if the field 'name' belongs to a table called 'users' and the
	// excludedTableQualifiers contains 'users', the field should present itself
	// as 'name' and not 'users.name'. i.e. any table qualifiers in the list
	// must be excluded.
	//
	// This is to play nice with certain clauses in the INSERT and UPDATE
	// queries that expressly forbid table qualified columns.
	AppendSQLExclude(dialect string, buf *strings.Builder, args *[]interface{}, params map[string]int, excludedTableQualifiers []string) error
	GetAlias() string
	GetName() string // Field name must exclude the table name
}

type Predicate interface {
	Field
	Not() Predicate
}

type namedparam struct {
	name  string
	value interface{}
}

type Params map[string]interface{}

func Param(name string, value interface{}) Field {
	return namedparam{name: name, value: value}
}
func (param namedparam) GetAlias() string { return "" }
func (param namedparam) GetName() string  { return "" }
func (param namedparam) AppendSQLExclude(dialect string, buf *strings.Builder, args *[]interface{}, params map[string]int, excludedTableQualifiers []string) error {
	buf.WriteString("?")
	*args = append(*args, param.value)
	params[param.name] = len(*args) - 1
	return nil
}

var _ Field = namedparam{}

var bufpool = sync.Pool{
	New: func() interface{} { return new(strings.Builder) },
}

var argspool = sync.Pool{
	New: func() interface{} { return make([]interface{}, 0) },
}

type Queryer interface {
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	Exec(query string, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type queryer Queryer

type QueryerLogger interface {
	Queryer
	GetLogger(ctx context.Context) (Logger, LogFlag, error)
}

type Transactor interface {
	Begin() (*sql.Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
}

type SelectType string

const (
	SelectTypeDefault    SelectType = "SELECT"
	SelectTypeDistinct   SelectType = "SELECT DISTINCT"
	SelectTypeDistinctOn SelectType = "SELECT DISTINCT ON"
)

type SQLiteDialect struct{}
type PostgresDialect struct{}

var SQLite = SQLiteDialect{}
var Postgres = PostgresDialect{}
