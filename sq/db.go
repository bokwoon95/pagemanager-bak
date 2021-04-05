package sq

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/bokwoon95/erro"
)

type DB struct {
	queryer
	logger  Logger
	logFlag LogFlag
}

func NewDB(db Queryer, logger Logger, logflag LogFlag) DB {
	return DB{queryer: db, logger: logger, logFlag: logflag}
}

func NewDefaultDB(db Queryer) DB {
	return DB{queryer: db, logger: defaultlogger, logFlag: Lcompact}
}

func NewTx(tx *sql.Tx, src DB) Queryer {
	return DB{queryer: tx, logger: src.logger, logFlag: src.logFlag}
}

func (db DB) GetLogger(context.Context) (Logger, LogFlag, error) { return db.logger, db.logFlag, nil }

func (db DB) Begin() (*sql.Tx, error) {
	txer, ok := db.queryer.(Transactor)
	if !ok {
		return nil, ErrUnsupported
	}
	tx, err := txer.Begin()
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func (db DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error) {
	txer, ok := db.queryer.(Transactor)
	if !ok {
		return nil, ErrUnsupported
	}
	tx, err := txer.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return tx, nil
}

func WithTxContext(ctx context.Context, txer Transactor, opts *sql.TxOptions, fn func(*sql.Tx) error) error {
	tx, err := txer.BeginTx(ctx, opts)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	err = fn(tx)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func WithTx(txer Transactor, fn func(*sql.Tx) error) error {
	return WithTxContext(context.Background(), txer, nil, fn)
}

func Fetch(db Queryer, q Query, rowmapper func(*Row) error) (rowCount int64, err error) {
	return fetchContext(context.Background(), db, q, rowmapper, 1)
}

func FetchContext(ctx context.Context, db Queryer, q Query, rowmapper func(*Row) error) (rowCount int64, err error) {
	return fetchContext(ctx, db, q, rowmapper, 1)
}

func fetchContext(ctx context.Context, db Queryer, q Query, rowmapper func(*Row) error, skip int) (rowCount int64, err error) {
	if db == nil {
		return 0, errors.New("db is nil")
	}
	if q == nil {
		return 0, errors.New("query is nil")
	}
	if rowmapper == nil {
		return 0, errors.New("cannot call Fetch/FetchContext without a rowmapper")
	}
	var logger Logger
	var logflag LogFlag
	if db, ok := db.(QueryerLogger); ok {
		logger, logflag, err = db.GetLogger(ctx)
		if err != nil {
			logger = nil
			logflag = 0
		}
	}
	stats := QueryStats{
		Dialect: q.Dialect(),
		LogFlag: logflag,
	}
	if Lcaller&logflag != 0 {
		stats.Filename, stats.LineNumber, stats.FunctionName = caller(skip + 1)
	}
	r := &Row{}
	start := time.Now()
	defer func() {
		if logger == nil {
			return
		}
		stats.Error = err
		stats.TimeTaken = time.Since(start)
		stats.RowCount = r.count
		logger.LogQueryStats(ctx, stats)
	}()
	err = rowmapper(r)
	if err != nil {
		return 0, err
	}
	q, err = q.SetFetchableFields(r.fields) // Queries must handle the case when len(r.fields) == 0. For example, SelectQuery must default to SELECT 1 in case the rowmapper does nothing
	if err != nil {
		return 0, err
	}
	buf := bufpool.Get().(*strings.Builder)
	resultsBuf := bufpool.Get().(*strings.Builder)
	tmpbuf := bufpool.Get().(*strings.Builder)
	tmpargs := argspool.Get().([]interface{})
	defer func() {
		if resultsBuf.Len() > 0 {
			stats.ResultsPreview = resultsBuf.String()
		}
		buf.Reset()
		resultsBuf.Reset()
		tmpbuf.Reset()
		tmpargs = tmpargs[:0]
		bufpool.Put(buf)
		bufpool.Put(resultsBuf)
		bufpool.Put(tmpbuf)
		argspool.Put(tmpargs)
	}()
	_ = q.AppendSQL("", buf, &stats.Args, make(map[string]int))
	stats.Query = buf.String()
	if stats.Dialect == "postgres" {
		stats.Query = QuestionToDollarPlaceholders(stats.Query)
	}
	r.rows, err = db.QueryContext(ctx, stats.Query, stats.Args...)
	if err != nil {
		return 0, err
	}
	if len(r.dest) == 0 {
		return 0, nil
	}
	for r.rows.Next() {
		r.count++
		err = r.rows.Scan(r.dest...)
		if err != nil {
			err = wrapScanError(err, r)
			return r.count, err
		}
		if logger != nil && Lresults&logflag != 0 && r.count <= 5 {
			resultsBuf.WriteString("\n----[ Row ")
			resultsBuf.WriteString(strconv.FormatInt(r.count, 10))
			resultsBuf.WriteString(" ]----")
			for i := range r.dest {
				tmpbuf.Reset()
				tmpargs = tmpargs[:0]
				_ = r.fields[i].AppendSQLExclude("", tmpbuf, &tmpargs, nil, nil)
				resultsBuf.WriteString("\n")
				resultsBuf.WriteString(QuestionInterpolate(tmpbuf.String(), tmpargs...))
				resultsBuf.WriteString(": ")
				appendSQLDisplay(resultsBuf, r.dest[i])
			}
		}
		r.index = 0
		err = rowmapper(r)
		if err != nil {
			if errors.Is(err, SkipRows) {
				break
			}
			return r.count, err
		}
	}
	err = r.rows.Close()
	if err != nil {
		return r.count, err
	}
	err = r.rows.Err()
	if err != nil {
		return r.count, err
	}
	if r.count > 5 {
		resultsBuf.WriteString("\n...\n(" + strconv.FormatInt(r.count-5, 10) + " more rows)")
	}
	return r.count, nil
}

func wrapScanError(err error, row *Row) error {
	buf := bufpool.Get().(*strings.Builder)
	tmpbuf := bufpool.Get().(*strings.Builder)
	tmpargs := argspool.Get().([]interface{})
	defer func() {
		buf.Reset()
		tmpbuf.Reset()
		tmpargs = tmpargs[:0]
		bufpool.Put(buf)
		bufpool.Put(tmpbuf)
		argspool.Put(tmpargs)
	}()
	for i := range row.dest {
		tmpbuf.Reset()
		tmpargs = tmpargs[:0]
		_ = row.fields[i].AppendSQLExclude("", tmpbuf, &tmpargs, make(map[string]int), nil)
		buf.WriteString("\n" + strconv.Itoa(i) + ") ")
		buf.WriteString(QuestionInterpolate(tmpbuf.String(), tmpargs...))
		buf.WriteString(" => ")
		buf.WriteString(reflect.TypeOf(row.dest[i]).String())
	}
	return fmt.Errorf("Please check if your mapper function is correct:%s\n%w", buf.String(), err)
}

func Exec(db Queryer, q Query, execFlag ExecFlag) (rowsAffected, lastInsertID int64, err error) {
	return execContext(context.Background(), db, q, execFlag, 1)
}

func ExecContext(ctx context.Context, db Queryer, q Query, execFlag ExecFlag) (rowsAffected, lastInsertID int64, err error) {
	return execContext(ctx, db, q, execFlag, 1)
}

func execContext(ctx context.Context, db Queryer, q Query, execFlag ExecFlag, skip int) (rowsAffected, lastInsertID int64, err error) {
	if db == nil {
		return 0, 0, errors.New("db is nil")
	}
	if q == nil {
		return 0, 0, errors.New("query is nil")
	}
	var logger Logger
	var logflag LogFlag
	if db, ok := db.(QueryerLogger); ok {
		logger, logflag, err = db.GetLogger(ctx)
		if err != nil {
			logger = nil
			logflag = 0
		}
	}
	stats := QueryStats{
		Dialect:  q.Dialect(),
		LogFlag:  logflag,
		ExecFlag: execFlag | ExecActive,
	}
	if Lcaller&logflag != 0 {
		stats.Filename, stats.LineNumber, stats.FunctionName = caller(skip + 1)
	}
	start := time.Now()
	defer func() {
		if logger == nil {
			return
		}
		stats.Error = err
		stats.TimeTaken = time.Since(start)
		stats.RowsAffected = rowsAffected
		stats.LastInsertID = lastInsertID
		logger.LogQueryStats(ctx, stats)
	}()
	buf := bufpool.Get().(*strings.Builder)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	_ = q.AppendSQL("", buf, &stats.Args, make(map[string]int))
	stats.Query = buf.String()
	if stats.Dialect == "postgres" {
		stats.Query = QuestionToDollarPlaceholders(stats.Query)
	}
	res, err := db.ExecContext(ctx, stats.Query, stats.Args...)
	if err != nil {
		return 0, 0, erro.Wrap(err)
	}
	if res != nil && ErowsAffected&execFlag != 0 {
		rowsAffected, err = res.RowsAffected()
		if err != nil {
			return 0, 0, erro.Wrap(err)
		}
	}
	if res != nil && ElastInsertID&execFlag != 0 {
		lastInsertID, err = res.LastInsertId()
		if err != nil {
			return 0, 0, erro.Wrap(err)
		}
	}
	return rowsAffected, lastInsertID, nil
}

func Exists(db DB, q Query) (exists bool, err error) {
	return existsContext(context.Background(), db, q, 1)
}

func ExistsContext(ctx context.Context, db Queryer, q Query) (exists bool, err error) {
	return existsContext(context.Background(), db, q, 1)
}

func existsContext(ctx context.Context, db Queryer, q Query, skip int) (exists bool, err error) {
	if db == nil {
		return false, errors.New("db is nil")
	}
	if q == nil {
		return false, errors.New("query is nil")
	}
	var logger Logger
	var logflag LogFlag
	if db, ok := db.(QueryerLogger); ok {
		logger, logflag, err = db.GetLogger(ctx)
		if err != nil {
			logger = nil
			logflag = 0
		}
	}
	stats := QueryStats{
		Dialect: q.Dialect(),
		LogFlag: logflag,
	}
	if Lcaller&logflag != 0 {
		stats.Filename, stats.LineNumber, stats.FunctionName = caller(skip + 1)
	}
	start := time.Now()
	defer func() {
		if logger == nil {
			return
		}
		stats.Error = err
		stats.TimeTaken = time.Since(start)
		if exists {
			stats.RowCount = 1
		}
		logger.LogQueryStats(ctx, stats)
	}()
	buf := bufpool.Get().(*strings.Builder)
	defer func() {
		buf.Reset()
		bufpool.Put(buf)
	}()
	q, err = q.SetFetchableFields([]Field{FieldLiteral("1")})
	if err != nil {
		return false, erro.Wrap(err)
	}
	buf.WriteString("SELECT EXISTS(")
	err = q.AppendSQL("", buf, &stats.Args, make(map[string]int))
	if err != nil {
		return false, erro.Wrap(err)
	}
	buf.WriteString(")")
	stats.Query = buf.String()
	if stats.Dialect == "postgres" {
		stats.Query = QuestionToDollarPlaceholders(stats.Query)
	}
	rows, err := db.Query(stats.Query, stats.Args...)
	if err != nil {
		return false, erro.Wrap(err)
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&exists)
		if err != nil {
			return false, erro.Wrap(err)
		}
		break
	}
	return exists, nil
}
