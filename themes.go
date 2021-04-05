package pagemanager

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/bokwoon95/erro"
	"github.com/davecgh/go-spew/spew"
	"github.com/dop251/goja"
	"github.com/mitchellh/mapstructure"
)

// /pm-themes/plainsimple/index.css
// /pm-themes/plainsimple/index.pm-sha256-RFWPLDbv2BY+rCkDzsE+0fr8ylGr2R2faWMhq4lfEQc=.css
// /pm-themes/plainsimple/data
// /pm-themes/plainsimple/data.pm-sha256-RFWPLDbv2BY+rCkDzsE+0fr8ylGr2R2faWMhq4lfEQc=
// /pm-themes/plainsimple/haha.meh
// /pm-themes/plainsimple/haha.pm-sha256-RFWPLDbv2BY+rCkDzsE+0fr8ylGr2R2faWMhq4lfEQc=.meh

type themeTemplate struct {
	HTML                  []string
	CSS                   []string
	JS                    []string
	TemplateVariables     map[string]interface{}
	ContentSecurityPolicy map[string][]string
}

type theme struct {
	err            error  // any error encountered when parsing theme-config.js
	path           string // path to the theme folder in the "themes" folder
	name           string
	description    string
	fallbackAssets map[string]string
	themeTemplates map[string]themeTemplate
}

type theme2 struct {
	Name           string
	Description    string
	FallbackAssets map[string]string
	Templates      map[string]themeTemplate
}

func getThemes(datafolder string) (themes map[string]theme, fallbackAssetsIndex map[string]string, err error) {
	themes, fallbackAssetsIndex = make(map[string]theme), make(map[string]string)
	if datafolder == "" {
		return themes, fallbackAssetsIndex, erro.Wrap(fmt.Errorf("pm.datafolder is empty"))
	}
	err = filepath.WalkDir(filepath.Join(datafolder, "pm-themes"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		b, err := os.ReadFile(filepath.Join(path, "theme-config.js"))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil // if theme-config.js doesn't exist in current dir, keep looking
			}
			return erro.Wrap(err)
		}
		cwd := strings.TrimPrefix(path, datafolder)
		if runtime.GOOS == "windows" {
			cwd = strings.ReplaceAll(cwd, `\`, `/`) // theme_path is always stored with unix-style forward slashes
		}
		t := theme{path: strings.TrimPrefix(cwd, "/pm-themes/")}
		defer func() {
			themes[t.path] = t
		}()
		vm := goja.New()
		vm.Set("$THEME_PATH", cwd+"/")
		res, err := vm.RunString("(function(){" + string(b) + "})()")
		if err != nil {
			t.err = err
			return fs.SkipDir
		}
		var t2 theme2
		err = mapstructure.Decode(res.Export(), &t2)
		if err != nil {
			t.err = err
			return fs.SkipDir
		}
		t3 := theme3{
			path:           strings.TrimPrefix(cwd, "/pm-themes/"),
			fallbackAssets: make(map[string]string),
			themeTemplates: make(map[string]themeTemplate2),
		}
		t3.Unmarshal(res.Export())
		spew.Dump(t3)
		t.name = t2.Name
		t.description = t2.Description
		t.fallbackAssets = make(map[string]string)
		for name, fallback := range t2.FallbackAssets {
			if !strings.HasPrefix(fallback, "/") {
				fallback = cwd + "/" + fallback
			}
			t.fallbackAssets[name] = fallback
			fallbackAssetsIndex[name] = t.path
		}
		t.themeTemplates = make(map[string]themeTemplate)
		for name, tt2 := range t2.Templates {
			tt := themeTemplate{
				HTML:                  make([]string, len(tt2.HTML)),
				CSS:                   make([]string, len(tt2.CSS)),
				JS:                    make([]string, len(tt2.JS)),
				TemplateVariables:     tt2.TemplateVariables,
				ContentSecurityPolicy: tt2.ContentSecurityPolicy,
			}
			for i, path := range tt2.HTML {
				if !strings.HasPrefix(path, "/") {
					path = cwd + "/" + path
				}
				tt.HTML[i] = path
			}
			for i, path := range tt2.CSS {
				if !strings.HasPrefix(path, "/") {
					path = cwd + "/" + path
				}
				tt.CSS[i] = path
			}
			for i, path := range tt2.JS {
				if !strings.HasPrefix(path, "/") {
					path = cwd + "/" + path
				}
				tt.JS[i] = path
			}
			t.themeTemplates[name] = tt
		}
		return fs.SkipDir
	})
	if err != nil {
		return themes, fallbackAssetsIndex, erro.Wrap(err)
	}
	return themes, fallbackAssetsIndex, nil
}

type asset struct {
	path   string
	data   []byte
	hash   [32]byte
	inline bool
}

type themeTemplate2 struct {
	HTML                  []string
	CSS                   []asset
	JS                    []asset
	TemplateVariables     map[string]interface{}
	ContentSecurityPolicy map[string][]string
}

type theme3 struct {
	err            error  // any error encountered when parsing theme-config.js
	path           string // path to the theme folder in the "pm-themes" folder
	name           string
	description    string
	fallbackAssets map[string]string
	themeTemplates map[string]themeTemplate2
}

func (t *theme3) Unmarshal(data interface{}) {
	data2, ok := data.(map[string]interface{})
	if !ok {
		return
	}
	themePath := "/pm-themes/" + t.path
	t.name, _ = data2["Name"].(string)
	t.description, _ = data2["Description"].(string)
	fallbackAssets, _ := data2["FallbackAssets"].(map[string]interface{})
	for asset, __fallback__ := range fallbackAssets {
		fallback, ok := __fallback__.(string)
		if !ok {
			continue
		}
		if strings.HasPrefix(fallback, "/") {
			t.fallbackAssets[asset] = fallback
		} else {
			t.fallbackAssets[asset] = themePath + "/" + fallback
		}
	}
	templates, _ := data2["Templates"].(map[string]interface{})
	for templateName, __template__ := range templates {
		tt := themeTemplate2{
			TemplateVariables:     make(map[string]interface{}),
			ContentSecurityPolicy: make(map[string][]string),
		}
		template, _ := __template__.(map[string]interface{})
		HTMLs, _ := template["HTML"].([]interface{})
		for _, __html__ := range HTMLs {
			html, ok := __html__.(string)
			if !ok {
				continue
			}
			if strings.HasPrefix(html, "/") {
				tt.HTML = append(tt.HTML, html)
			} else {
				tt.HTML = append(tt.HTML, themePath+"/"+html)
			}
		}
		CSSs, _ := template["CSS"].([]interface{})
		for _, __css__ := range CSSs {
			var a asset
			switch css := __css__.(type) {
			case string:
				if strings.HasPrefix(css, "/") {
					a.path = css
				} else {
					a.path = themePath + "/" + css
				}
				tt.CSS = append(tt.CSS, a)
			case map[string]interface{}:
				a.path, _ = css["Path"].(string)
				a.inline, _ = css["Inline"].(bool)
				tt.CSS = append(tt.CSS, a)
			default:
				continue
			}
		}
		JSs, _ := template["JS"].([]interface{})
		for _, __js__ := range JSs {
			var a asset
			switch js := __js__.(type) {
			case string:
				if strings.HasPrefix(js, "/") {
					a.path = js
				} else {
					a.path = themePath + "/" + js
				}
				tt.JS = append(tt.JS, a)
			case map[string]interface{}:
				a.path, _ = js["Path"].(string)
				a.inline, _ = js["Inline"].(bool)
				tt.JS = append(tt.JS, a)
			default:
				continue
			}
		}
		tt.TemplateVariables, _ = template["TemplateVariables"].(map[string]interface{})
		contentSecurityPolicy, _ := template["ContentSecurityPolicy"].(map[string]interface{})
		for name, __policies__ := range contentSecurityPolicy {
			policies, _ := __policies__.([]interface{})
			for _, __policy__ := range policies {
				policy, ok := __policy__.(string)
				if !ok {
					continue
				}
				tt.ContentSecurityPolicy[name] = append(tt.ContentSecurityPolicy[name], policy)
			}
		}
		t.themeTemplates[templateName] = tt
	}
}
