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
