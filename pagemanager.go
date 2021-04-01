package pagemanager

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/bokwoon95/erro"
	"github.com/bokwoon95/pagemanager/sq"
	"github.com/bokwoon95/pagemanager/tables"
)

var assetsFS fs.FS
var templatesFS fs.FS
var flagDatafolder = flag.String("pm-datafolder", "", "")
var flagSuperadminFolder = flag.String("pm-superadmin", "", "")
var bufpool = sync.Pool{
	New: func() interface{} { return new(strings.Builder) },
}

func init() {
	_, currentFile, _, _ := runtime.Caller(0)
	currentDir := filepath.Dir(currentFile)
	assetsDir := filepath.Join(currentDir, "assets")
	if assetsFS == nil {
		assetsFS = os.DirFS(assetsDir)
	}
	templatesDir := filepath.Join(currentDir, "templates")
	if templatesFS == nil {
		templatesFS = os.DirFS(templatesDir)
	}
}

type PageManager struct {
	themesMutex         *sync.RWMutex
	themes              map[string]theme
	fallbackAssetsIndex map[string]string // asset => theme name
	datafolder          string
	superadminfolder    string
	dataDB              *sql.DB
	superadminDB        *sql.DB
	secretkey           []byte // key-stretched from user's low-entropy password
	localesMutex        *sync.RWMutex
	locales             map[string]string
}

type Route struct {
	LocaleCode  string
	URL         sql.NullString
	Disabled    sql.NullBool
	RedirectURL sql.NullString
	HandlerURL  sql.NullString
	Content     sql.NullString
	ThemePath   sql.NullString
	Template    sql.NullString
}

func New() (*PageManager, error) {
	var err error
	pm := &PageManager{}
	pm.themesMutex = &sync.RWMutex{}
	pm.localesMutex = &sync.RWMutex{}
	pm.themes = make(map[string]theme)
	pm.datafolder, err = LocateDataFolder()
	if err != nil {
		return pm, erro.Wrap(err)
	}
	pm.superadminfolder, err = LocateSuperadminFolder(pm.datafolder)
	if err != nil {
		return pm, erro.Wrap(err)
	}
	pm.dataDB, err = sql.Open("sqlite3", filepath.Join(pm.datafolder, "database.sqlite3"+
		"?_journal_mode=WAL"+
		"&_synchronous=NORMAL"+
		"&_foreign_keys=on",
	))
	if err != nil {
		return pm, erro.Wrap(err)
	}
	pm.superadminDB, err = sql.Open("sqlite3", filepath.Join(pm.superadminfolder, "superadmin.sqlite3"+
		"?_journal_mode=WAL"+
		"&_synchronous=NORMAL"+
		"&_foreign_keys=on",
	))
	if err != nil {
		return pm, erro.Wrap(err)
	}
	ctx := context.Background()
	err = sq.EnsureTables(pm.dataDB, "sqlite3",
		tables.NEW_PAGES(ctx, ""),
		tables.NEW_PAGEDATA(ctx, ""),
		tables.NEW_USERS(ctx, ""),
		tables.NEW_AUTHZ_GROUPS(ctx, ""),
		tables.NEW_SESSIONS(ctx, ""),
		tables.NEW_LOCALES(ctx, ""),
	)
	if err != nil {
		return pm, erro.Wrap(err)
	}
	err = sq.EnsureTables(pm.superadminDB, "sqlite3",
		tables.NEW_SUPERADMIN(ctx, ""),
		tables.NEW_ENCRYPTION_KEYS(ctx, ""),
		tables.NEW_MAC_KEYS(ctx, ""),
	)
	if err != nil {
		return pm, erro.Wrap(err)
	}
	err = seedData(pm.dataDB)
	if err != nil {
		return pm, erro.Wrap(err)
	}
	pm.themes, pm.fallbackAssetsIndex, err = getThemes(pm.datafolder)
	if err != nil {
		return pm, erro.Wrap(err)
	}
	pm.locales, err = getLocales(pm.dataDB)
	if err != nil {
		return pm, erro.Wrap(err)
	}
	return pm, nil
}

func LocateDataFolder() (string, error) {
	const datafoldername = "pagemanager-data"
	cwd, err := os.Getwd()
	if err != nil {
		return "", erro.Wrap(err)
	}
	userhome, err := os.UserHomeDir()
	if err != nil {
		return "", erro.Wrap(err)
	}
	exePath, err := os.Executable()
	if err != nil {
		return "", erro.Wrap(err)
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return "", erro.Wrap(err)
	}
	exeDir := filepath.Dir(exePath)
	paths := []string{
		cwd,                                     // $CWD
		filepath.Join(cwd, datafoldername),      // $CWD/pagemanager-data
		filepath.Join(userhome, datafoldername), // $HOME/pagemanager-data
		exeDir,                                  // $EXE_DIR
		filepath.Join(exeDir, datafoldername),   // $EXE_DIR/pagemanager-data
	}
	if *flagDatafolder != "" {
		if strings.HasPrefix(*flagDatafolder, ".") {
			return cwd + (*flagDatafolder)[1:], nil
		}
		return *flagDatafolder, nil
	}
	for _, path := range paths {
		if filepath.Base(path) != datafoldername {
			continue
		}
		dir, err := os.Open(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return "", erro.Wrap(err)
		}
		defer dir.Close()
		info, err := dir.Stat()
		if err != nil {
			return "", erro.Wrap(err)
		}
		if info.IsDir() {
			return path, nil
		}
	}
	defaultpath := filepath.Join(userhome, datafoldername)
	err = os.MkdirAll(defaultpath, 0775)
	if err != nil {
		return "", erro.Wrap(err)
	}
	err = os.MkdirAll(filepath.Join(defaultpath, "pm-themes"), 0775)
	if err != nil {
		return "", erro.Wrap(err)
	}
	err = os.MkdirAll(filepath.Join(defaultpath, "pm-images"), 0775)
	if err != nil {
		return "", erro.Wrap(err)
	}
	return defaultpath, nil
}

func LocateSuperadminFolder(datafolder string) (string, error) {
	const superadminfoldername = "pagemanager-superadmin"
	cwd, err := os.Getwd()
	if err != nil {
		return "", erro.Wrap(err)
	}
	userhome, err := os.UserHomeDir()
	if err != nil {
		return "", erro.Wrap(err)
	}
	exePath, err := os.Executable()
	if err != nil {
		return "", erro.Wrap(err)
	}
	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return "", erro.Wrap(err)
	}
	exeDir := filepath.Dir(exePath)
	paths := []string{
		cwd,                                      // $CWD
		filepath.Join(cwd, superadminfoldername), // $CWD/pagemanager-superadmin
		filepath.Join(userhome, superadminfoldername), // $HOME/pagemanager-superadmin
		exeDir, // $EXE_DIR
		filepath.Join(exeDir, superadminfoldername), // $EXE_DIR/pagemanager-superadmin
	}
	if *flagSuperadminFolder != "" {
		if strings.HasPrefix(*flagSuperadminFolder, ".") {
			return cwd + (*flagSuperadminFolder)[1:], nil
		}
		return *flagSuperadminFolder, nil
	}
	if !strings.HasSuffix(datafolder, string(os.PathSeparator)) {
		datafolder += string(os.PathSeparator)
	}
	for _, path := range paths {
		// superadminfolder must not be located inside the datafolder
		if strings.HasPrefix(path, datafolder) {
			continue
		}
		if filepath.Base(path) != superadminfoldername {
			continue
		}
		f, err := os.Open(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			return "", erro.Wrap(err)
		}
		f.Close()
		return path, nil
	}
	defaultpath := filepath.Join(userhome, superadminfoldername)
	if strings.HasPrefix(defaultpath, datafolder) {
		return "", erro.Wrap(fmt.Errorf("superadminfolder defaultpath resides in the datafolder"))
	}
	err = os.MkdirAll(defaultpath, 0775)
	if err != nil {
		return "", erro.Wrap(err)
	}
	return defaultpath, nil
}

func seedData(db sq.Queryer) error {
	ctx := context.Background()
	p := tables.NEW_PAGES(ctx, "p")
	db = sq.NewDB(db, nil, sq.Linterpolate|sq.Lcaller)
	_, _, err := sq.Exec(db, sq.SQLite.DeleteFrom(p), sq.ErowsAffected)
	if err != nil {
		return erro.Wrap(err)
	}
	// pm_pages.content
	_, _, err = sq.Exec(db, sq.SQLite.
		InsertInto(p).
		Valuesx(func(col *sq.Column) error {
			col.SetString(p.URL, `/hello/`)
			col.SetString(p.CONTENT, `<h1>This is hello</h1>`)
			return nil
		}).
		OnConflict(p.URL).
		DoUpdateSet(sq.SetExcluded(p.CONTENT)),
		sq.ErowsAffected,
	)
	if err != nil {
		return erro.Wrap(err)
	}
	// pm_pages.handler_url
	_, _, err = sq.Exec(db, sq.SQLite.
		InsertInto(p).
		Valuesx(func(col *sq.Column) error {
			col.SetString(p.URL, `/goodbye`)
			col.SetString(p.HANDLER_URL, `/`)
			return nil
		}).
		OnConflict(p.URL).
		DoUpdateSet(sq.SetExcluded(p.HANDLER_URL)),
		sq.ErowsAffected,
	)
	if err != nil {
		return erro.Wrap(err)
	}
	// pm_pages.theme_path, pm_pages.template
	var templates = []struct {
		url, theme_path, template string
	}{
		{"/posts", "plainsimple", "PostsIndex"},
	}
	_, _, err = sq.Exec(db, sq.SQLite.
		InsertInto(p).
		Valuesx(func(col *sq.Column) error {
			for _, t := range templates {
				col.SetString(p.URL, t.url)
				col.SetString(p.THEME_PATH, t.theme_path)
				col.SetString(p.TEMPLATE, t.template)
			}
			return nil
		}).
		OnConflict(p.URL).
		DoUpdateSet(sq.SetExcluded(p.THEME_PATH), sq.SetExcluded(p.TEMPLATE)),
		sq.ErowsAffected,
	)
	if err != nil {
		return erro.Wrap(err)
	}
	// pm_sessions
	s := tables.NEW_SESSIONS(ctx, "s")
	var sessions = []struct {
		sessionhash string
		userid      int64
		sessiondata map[string]interface{}
	}{
		{"1234", 0, map[string]interface{}{"yeet": 1}},
	}
	_, _, err = sq.Exec(db, sq.SQLite.
		InsertInto(s).
		Valuesx(func(col *sq.Column) error {
			var b []byte
			for _, sess := range sessions {
				col.SetString(s.SESSION_HASH, sess.sessionhash)
				col.SetInt64(s.USER_ID, sess.userid)
				b, err = json.Marshal(sess.sessiondata)
				if err != nil {
					return erro.Wrap(err)
				}
				col.Set(s.SESSION_DATA, string(b))
			}
			return nil
		}).
		OnConflict().DoNothing(),
		sq.ErowsAffected,
	)
	if err != nil {
		return erro.Wrap(err)
	}
	// pm_users, pm_authz_groups
	u, ag := tables.NEW_USERS(ctx, "u"), tables.NEW_AUTHZ_GROUPS(ctx, "ag")
	_, _, err = sq.Exec(db, sq.SQLite.DeleteFrom(u), sq.ErowsAffected)
	if err != nil {
		return erro.Wrap(err)
	}
	_, _, err = sq.Exec(db, sq.SQLite.DeleteFrom(ag), sq.ErowsAffected)
	if err != nil {
		return erro.Wrap(err)
	}
	var users = []struct {
		userid      int64
		publicid    string
		username    string
		authzgroups []string
	}{
		{0, "", "", []string{"pm-pagemanager"}},
	}
	_, _, err = sq.Exec(db, sq.SQLite.
		InsertInto(u).
		Valuesx(func(col *sq.Column) error {
			var b []byte
			for _, user := range users {
				col.SetInt64(u.USER_ID, user.userid)
				col.SetString(u.PUBLIC_USER_ID, user.publicid)
				col.SetString(u.USERNAME, user.username)
				b, err = json.Marshal(user.authzgroups)
				if err != nil {
					return erro.Wrap(err)
				}
				col.Set(u.AUTHZ_GROUPS, string(b))
			}
			return nil
		}),
		sq.ErowsAffected,
	)
	if err != nil {
		return erro.Wrap(err)
	}
	var groups = []struct {
		name string
		data map[string]interface{}
	}{
		{"pm-pagemanager", map[string]interface{}{"pm-page-perms": PageCreate | PageRead | PageUpdate | PageDelete}},
	}
	_, _, err = sq.Exec(db, sq.SQLite.
		InsertInto(ag).
		Valuesx(func(col *sq.Column) error {
			var b []byte
			for _, group := range groups {
				col.SetString(ag.NAME, group.name)
				b, err = json.Marshal(group.data)
				if err != nil {
					return erro.Wrap(err)
				}
				col.Set(ag.AUTHZ_DATA, string(b))
			}
			return nil
		}),
		sq.ErowsAffected,
	)
	if err != nil {
		return erro.Wrap(err)
	}
	// pm_locales
	l := tables.NEW_LOCALES(ctx, "l")
	_, _, err = sq.Exec(db, sq.SQLite.DeleteFrom(l), sq.ErowsAffected)
	if err != nil {
		return erro.Wrap(err)
	}
	var locales = []struct {
		code        string
		description string
	}{
		{"en", "English"},
	}
	_, _, err = sq.Exec(db, sq.SQLite.
		InsertInto(l).
		Valuesx(func(col *sq.Column) error {
			for _, locale := range locales {
				col.SetString(l.LOCALE_CODE, locale.code)
				col.SetString(l.DESCRIPTION, locale.description)
			}
			return nil
		}),
		sq.ErowsAffected,
	)
	if err != nil {
		return erro.Wrap(err)
	}
	return nil
}

const (
	PageCreate = 1 << iota
	PageRead
	PageUpdate
	PageDelete
)

func getLocales(db sq.Queryer) (map[string]string, error) {
	ctx := context.Background()
	l := tables.NEW_LOCALES(ctx, "l")
	db = sq.NewDB(db, nil, sq.Linterpolate|sq.Lcaller)
	locales := make(map[string]string)
	_, err := sq.Fetch(db, sq.SQLite.From(l), func(row *sq.Row) error {
		localeCode := row.String(l.LOCALE_CODE)
		description := row.String(l.DESCRIPTION)
		if row.Count() == 0 {
			return nil
		}
		locales[localeCode] = description
		return nil
	})
	if err != nil {
		return locales, erro.Wrap(err)
	}
	return locales, nil
}
