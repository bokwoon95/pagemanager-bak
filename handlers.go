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
	"github.com/bokwoon95/pagemanager/hy"
	"github.com/bokwoon95/pagemanager/hyforms"
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

func (d *superadminLoginData) Form(form *hyforms.Form) {
	const acceptTOSMsg = "need to accept our TOS"
	// inputs
	password := form.
		Input("password", "pm-superadmin-password", "").
		Set("#pm-superadmin-password.bg-near-white.pa2.w-100", hy.Attr{"required": hy.Enabled})
	rememberme := form.
		Checkbox("remember-me", "", d.RememberMe).
		Set("#remember-me.pointer", nil)

	// marshal
	form.Set("#loginform.bg-white", hy.Attr{"name": "loginform", "method": "POST", "action": ""})
	form.Append("div.mv2.pt2", nil, hy.H("label.pointer", hy.Attr{"for": password.ID()}, hy.Txt("Password:")))
	form.Append("div", nil, password)
	if hyforms.ErrMsgsMatch(password.ErrMsgs(), hyforms.NoneOfErrMsg) {
		form.Append("div.f7.red", nil, hy.Txt("your password is one of the blacklisted passwords, please try another one"))
	}
	form.Append("div.mv2.pt2", nil, rememberme, hy.H("label.ml1.pointer", hy.Attr{"for": rememberme.ID()}, hy.Txt("Remember Me")))
	if hyforms.ErrMsgsMatch(rememberme.ErrMsgs(), acceptTOSMsg) {
		form.Append("div.f7.red", nil, hy.Txt("You need to accept our terms and conditions"))
	}
	form.Append("div.mv2.pt2", nil, hy.H("button.pointer", hy.Attr{"type": "submit"}, hy.Txt("Log in")))

	// unmarshal
	form.Unmarshal(func() {
		d.Password = password.Validate(hyforms.Required, hyforms.NoneOf("1234")).Value()
		d.RememberMe = rememberme.Checked()
		if !d.RememberMe {
			form.AddInputErrMsgs(rememberme.Name(), acceptTOSMsg)
		}
	})
}

func (pm *PageManager) superadminLogin(w http.ResponseWriter, r *http.Request) {
	type Data struct {
		CSS       template.HTML
		JS        template.HTML
		LoginForm template.HTML
	}
	switch r.Method {
	case "GET":
		data := Data{}
		var err error
		d := &superadminLoginData{}
		_ = hyforms.CookiePop(w, r, "yeetus", d)
		data.LoginForm, err = hyforms.MarshalForm(nil, w, r, d.Form)
		if err != nil {
			http.Error(w, erro.Wrap(err).Error(), http.StatusInternalServerError)
			return
		}
		data.CSS, err = hy.MarshalElement(nil, hy.Elements{
			hy.H("link[rel=stylesheet][type=text/css]", hy.Attr{"href": "/pm-plugins/pagemanager/tachyons.css"}),
			hy.H("link[rel=stylesheet][type=text/css]", hy.Attr{"href": "/pm-plugins/pagemanager/style.css"}),
		})
		if err != nil {
			http.Error(w, erro.Wrap(err).Error(), http.StatusInternalServerError)
			return
		}
		data.JS, err = hy.MarshalElement(nil, hy.Elements{
			hy.JSON("[data-pm-json]", nil, map[string]interface{}{"yeet": 42069}),
			hy.H("script", hy.Attr{"src": "/pm-plugins/pagemanager/pmJSON.js"}),
		})
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
		err := hyforms.UnmarshalForm(w, r, d.Form)
		if err != nil {
			_ = hyforms.CookieSet(w, "yeetus", d, nil)
			http.Redirect(w, r, LocaleURL(r), http.StatusMovedPermanently)
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
