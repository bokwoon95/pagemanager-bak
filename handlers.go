package pagemanager

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/bokwoon95/erro"
	"github.com/bokwoon95/pagemanager/sq"
	"github.com/bokwoon95/pagemanager/tables"
	_ "github.com/mattn/go-sqlite3"
)

func (pm *PageManager) PageManager(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/pm-themes/") ||
			strings.HasPrefix(r.URL.Path, "/pm-images/") ||
			strings.HasPrefix(r.URL.Path, "/pm-plugins/pagemanager/") {
			pm.serveFile(w, r, r.URL.Path)
			return
		}
		route, err := pm.getRoute(r.Context(), r.URL.Path)
		if err != nil {
			http.Error(w, erro.Sdump(err), http.StatusInternalServerError)
			return
		}
		if route.Disabled.Valid && route.Disabled.Bool {
			http.NotFound(w, r)
			return
		}
		if route.HandlerURL.Valid {
			r2 := &http.Request{}
			*r2 = *r
			r2.URL = &url.URL{}
			r2.URL.Path = route.HandlerURL.String
			next.ServeHTTP(w, r2)
			return
		}
		if route.Content.Valid {
			io.WriteString(w, route.Content.String)
			return
		}
		if route.ThemePath.Valid && route.Template.Valid {
			pm.serveTemplate(w, r, route.ThemePath.String, route.Template.String)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (pm *PageManager) getRoute(ctx context.Context, path string) (Route, error) {
	var route Route
	var err error
	elems := strings.SplitN(path, "/", 3)
	if len(elems) >= 2 {
		head := elems[1]
		pm.localesMutex.RLock()
		_, ok := pm.locales[head]
		pm.localesMutex.RUnlock()
		if ok {
			route.LocaleCode = head
			if len(elems) >= 3 {
				path = "/" + elems[2]
			} else {
				path = "/"
			}
		}
	}
	var negapath string
	if strings.HasSuffix(path, "/") {
		negapath = strings.TrimRight(path, "/")
	} else {
		negapath = path + "/"
	}
	p := tables.NEW_PAGES(ctx, "p")
	_, err = sq.Fetch(pm.dataDB, sq.SQLite.
		From(p).
		Where(p.URL.In([]string{path, negapath})).
		OrderBy(sq.Case(p.URL).When(path, 1).Else(2)).
		Limit(1),
		func(row *sq.Row) error {
			route.URL = row.NullString(p.URL)
			route.Disabled = row.NullBool(p.DISABLED)
			route.RedirectURL = row.NullString(p.REDIRECT_URL)
			route.HandlerURL = row.NullString(p.HANDLER_URL)
			route.Content = row.NullString(p.CONTENT)
			route.ThemePath = row.NullString(p.THEME_PATH)
			route.Template = row.NullString(p.TEMPLATE)
			return nil
		},
	)
	if err != nil {
		return route, erro.Wrap(err)
	}
	return route, nil
}

func (pm *PageManager) serveTemplate(w http.ResponseWriter, r *http.Request, Theme, Template string) {
	io.WriteString(w, Theme+" "+Template)
	pm.themesMutex.RLock()
	theme, ok := pm.themes[Theme]
	pm.themesMutex.RUnlock()
	if !ok {
		http.Error(w, erro.Sdump(fmt.Errorf("No such theme called %s", Theme)), http.StatusInternalServerError)
		return
	}
	if theme.err != nil {
		http.Error(w, erro.Sdump(theme.err), http.StatusInternalServerError)
		return
	}
	themeTemplate, ok := theme.themeTemplates[Template]
	if !ok {
		http.Error(w, erro.Sdump(fmt.Errorf("No such template called %s for theme %s", Template, Theme)), http.StatusInternalServerError)
		return
	}
	if len(themeTemplate.HTML) == 0 {
		http.Error(w, erro.Sdump(fmt.Errorf("template has no HTML files")), http.StatusInternalServerError)
		return
	}
	return
}

func (pm *PageManager) serveFile(w http.ResponseWriter, r *http.Request, name string) {
	var f fs.File
	var err error
	if strings.HasPrefix(r.URL.Path, "/pm-plugins/pagemanager/") {
		path := strings.TrimPrefix(filepath.Clean(r.URL.Path), "/pm-plugins/pagemanager/")
		f, err = assetsFS.Open(path)
	}
	if strings.HasPrefix(r.URL.Path, "/pm-themes/") || strings.HasPrefix(r.URL.Path, "/pm-images/") {
		path := strings.TrimPrefix(filepath.Clean(r.URL.Path), "/")
		if strings.HasSuffix(path, "theme-config.js") || strings.HasSuffix(path, ".html") {
			http.NotFound(w, r)
			return
		}
		datafolderFS := os.DirFS(pm.datafolder)
		f, err = datafolderFS.Open(path)
		if errors.Is(err, os.ErrNotExist) {
			func() {
				missingFile := "/" + path
				pm.themesMutex.RLock()
				defer pm.themesMutex.RUnlock()
				themeName, ok := pm.fallbackAssetsIndex[missingFile]
				if !ok {
					return
				}
				theme, ok := pm.themes[themeName]
				if !ok {
					return
				}
				fallbackFile, ok := theme.fallbackAssets[missingFile]
				if !ok {
					return
				}
				f, err = datafolderFS.Open(strings.TrimPrefix(fallbackFile, "/"))
			}()
		}
	}
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.NotFound(w, r)
		} else {
			http.Error(w, erro.Wrap(err).Error(), http.StatusInternalServerError)
		}
		return
	}
	if f == nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		http.Error(w, erro.Wrap(err).Error(), http.StatusInternalServerError)
		return
	}
	if info.IsDir() {
		http.NotFound(w, r)
		return
	}
	fseeker, ok := f.(io.ReadSeeker)
	if !ok {
		http.NotFound(w, r)
		return
	}
	http.ServeContent(w, r, name, info.ModTime(), fseeker)
}
