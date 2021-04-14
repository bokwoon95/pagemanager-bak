package pagemanager

import (
	"context"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/bokwoon95/erro"
	"github.com/bokwoon95/pagemanager/hyp"
	"github.com/bokwoon95/pagemanager/hypforms"
	"github.com/bokwoon95/pagemanager/sq"
	"github.com/bokwoon95/pagemanager/tables"
	_ "github.com/mattn/go-sqlite3"
)

func (pm *PageManager) PageManager(next http.Handler) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", next)
	mux.HandleFunc("/pm-superadmin", pm.superadminLogin)
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
		r2 := &http.Request{} // r2 is like r, but with the localeCode stripped from the URL and injected into the request context
		*r2 = *r
		r2 = r2.WithContext(context.WithValue(r2.Context(), LocaleCodeKey{}, route.LocaleCode))
		r2.URL = &url.URL{}
		*r2.URL = *r.URL
		r2.URL.Path = route.URL.String
		if route.Disabled.Valid && route.Disabled.Bool {
			http.NotFound(w, r)
			return
		}
		if route.HandlerURL.Valid {
			r2.URL.Path = route.HandlerURL.String
			next.ServeHTTP(w, r2)
			return
		}
		if route.Content.Valid {
			io.WriteString(w, route.Content.String)
			return
		}
		if route.ThemePath.Valid && route.Template.Valid {
			pm.serveTemplate(w, r2, route)
			return
		}
		mux.ServeHTTP(w, r2)
	})
}

const (
	EditModeOff      = 0
	EditModeBasic    = 1
	EditModeAdvanced = 2
)

type LocaleCodeKey struct{}

func (pm *PageManager) getRoute(ctx context.Context, path string) (Route, error) {
	var route Route
	var err error
	elems := strings.SplitN(path, "/", 3) // because first character of path is always '/', we ignore the first element
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
	if !route.URL.Valid {
		route.URL.String = path
		route.URL.Valid = true
	}
	return route, nil
}

func (pm *PageManager) serveTemplate(w http.ResponseWriter, r *http.Request, route Route) {
	pm.themesMutex.RLock()
	theme, ok := pm.themes[route.ThemePath.String]
	pm.themesMutex.RUnlock()
	if !ok {
		http.Error(w, erro.Sdump(fmt.Errorf("No such theme called %s", route.ThemePath.String)), http.StatusInternalServerError)
		return
	}
	if theme.err != nil {
		http.Error(w, erro.Sdump(theme.err), http.StatusInternalServerError)
		return
	}
	themeTemplate, ok := theme.themeTemplates[route.Template.String]
	if !ok {
		http.Error(w, erro.Sdump(fmt.Errorf("No such template called %s for theme %s", route.Template.String, route.ThemePath.String)), http.StatusInternalServerError)
		return
	}
	if len(themeTemplate.HTML) == 0 {
		http.Error(w, erro.Sdump(fmt.Errorf("template has no HTML files")), http.StatusInternalServerError)
		return
	}
	type Data struct {
		Page              PageData
		TemplateVariables map[string]interface{}
	}
	t := template.New("").Funcs(pm.funcmap())
	datafolderFS := os.DirFS(pm.datafolder)
	for _, filename := range themeTemplate.HTML {
		filename = strings.TrimPrefix(filename, "/")
		b, err := fs.ReadFile(datafolderFS, filename)
		if err != nil {
			http.Error(w, erro.Sdump(err), http.StatusInternalServerError)
			return
		}
		_, err = t.New(filename).Parse(string(b))
		if err != nil {
			http.Error(w, erro.Sdump(err), http.StatusInternalServerError)
			return
		}
	}
	t = t.Lookup(strings.TrimPrefix(themeTemplate.HTML[0], "/"))
	data := Data{
		Page: PageData{
			Ctx:        r.Context(),
			URL:        route.URL.String,
			DataID:     route.URL.String,
			LocaleCode: route.LocaleCode,
			CSSAssets:  themeTemplate.CSS,
			JSAssets:   themeTemplate.JS,
			CSP:        themeTemplate.ContentSecurityPolicy,
		},
		TemplateVariables: themeTemplate.TemplateVariables,
	}
	switch r.FormValue("pm-edit") {
	case "basic":
		data.Page.EditMode = EditModeBasic
	case "advanced":
		data.Page.EditMode = EditModeAdvanced
	}
	if data.Page.EditMode == EditModeBasic {
		data.Page.CSSAssets = append(data.Page.CSSAssets, Asset{Path: "/pm-plugins/pagemanager/editmode.css"})
		data.Page.JSAssets = append(data.Page.JSAssets, Asset{Path: "/pm-plugins/pagemanager/editmode.js"})
	}
	err := t.Execute(w, data)
	if err != nil {
		http.Error(w, erro.Sdump(err), http.StatusInternalServerError)
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

// sha256-JTm1pbrQf0HAbR27OuEt4ctbhU5wBu8sx03KF+37i5Y=

type superadminLoginData struct {
	Password   string
	RememberMe bool
}

func (d *superadminLoginData) Form(form *hypforms.Form) {
	type attr = hyp.Attr
	var h, txt = hyp.H, hyp.Txt
	password := form.Input("password", "pm-superadmin-password", "").Set("#pm-superadmin-password.bg-near-white.pa2.w-100", attr{
		"required": hyp.Enabled,
	})
	rememberme := form.Checkbox("remember-me", "").Set("#remember-me.pointer", nil)
	form.Set("#loginform.bg-white", attr{"name": "loginform", "method": "POST", "action": ""})
	form.Append("div.mv2.pt2", nil, h("label.pointer", attr{"for": "pm-superadmin-password"}, txt("Password:")))
	form.Append("div", nil, password)
	if errs := password.Errors(); len(errs) > 0 {
		div := h("div", nil)
		for _, err := range password.Errors() {
			div.Append("div.f6.gray", nil, txt(err.Error()))
		}
		form.AppendElements(div)
	}
	form.Append("div.mv2.pt2", nil, rememberme, h("label.ml1.pointer", attr{"for": "remember-me"}, txt("Remember Me")))
	form.Append("div.mv2.pt2", nil, h("button.pointer", attr{"type": "submit"}, txt("Log in")))
	form.Unmarshal(func() {
		d.Password = password.Validate().Value()
		d.RememberMe = rememberme.Checked()
	})
}

func (pm *PageManager) superadminLogin(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		Page      PageData
		CSS       template.HTML
		JS        template.HTML
		LoginForm template.HTML
	}
	switch r.Method {
	case "GET":
		data := Data{Page: NewPage()}
		data.Page.JSON["yeet"] = 42069
		data.Page.CSSAssets = []Asset{
			{Path: "/pm-plugins/pagemanager/tachyons.css"},
			{Path: "/pm-plugins/pagemanager/style.css"},
		}
		data.Page.JSAssets = []Asset{
			{Path: "/pm-plugins/pagemanager/pmJSON.js"},
		}
		var err error
		d := &superadminLoginData{}
		data.LoginForm, err = hypforms.MarshalForm(nil, r, d.Form)
		if err != nil {
			http.Error(w, erro.Wrap(err).Error(), http.StatusInternalServerError)
			return
		}
		data.CSS, err = hyp.Marshal(nil, h("", nil,
			h("link[rel=stylesheet][type=text/css]", attr{"href": "/pm-plugins/pagemanager/tachyons.css"}),
			h("link[rel=stylesheet][type=text/css]", attr{"href": "/pm-plugins/pagemanager/style.css"}),
		))
		if err != nil {
			http.Error(w, erro.Wrap(err).Error(), http.StatusInternalServerError)
			return
		}
		data.JS, err = hyp.Marshal(nil, h("", nil,
			hyp.JSON("[data-pm-json]", nil, map[string]interface{}{"yeet": 42069}),
			h("script", attr{"src": "/pm-plugins/pagemanager/pmJSON.js"}),
		))
		if err != nil {
			http.Error(w, erro.Wrap(err).Error(), http.StatusInternalServerError)
			return
		}
		t, err := pm.parseTemplates(templatesFS, "superadmin-login.html")
		if err != nil {
			http.Error(w, erro.Wrap(err).Error(), http.StatusInternalServerError)
			return
		}
		err = executeTemplate(t, w, data)
		if err != nil {
			http.Error(w, erro.Wrap(err).Error(), http.StatusInternalServerError)
			return
		}
	case "POST":
		var d superadminLoginData
		err := hypforms.UnmarshalForm(r, d.Form)
		fmt.Println(d)
		if err != nil {
			hypforms.Redirect(w, r, LocaleURL(r), err)
			return
		}
		http.Redirect(w, r, LocaleURL(r), http.StatusMovedPermanently)
	}
}

func LocaleURL(r *http.Request) string {
	localeCode, _ := r.Context().Value(LocaleCodeKey{}).(string)
	if localeCode == "" {
		return r.URL.Path
	}
	return "/" + localeCode + r.URL.Path
}
