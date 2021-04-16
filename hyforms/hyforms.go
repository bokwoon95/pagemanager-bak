package hyforms

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"html/template"
	"net/http"
	"runtime"
	"time"

	"github.com/bokwoon95/erro"
	"github.com/bokwoon95/pagemanager/encrypthash"
	"github.com/bokwoon95/pagemanager/hy"
	"github.com/microcosm-cc/bluemonday"
)

type ValidationError struct {
	FormErrMsgs  []string
	InputErrMsgs map[string][]string
	Expires      time.Time
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("form errors: %+v, input errors: %+v", e.FormErrMsgs, e.InputErrMsgs)
}

type Hyforms struct {
	sanitizer hy.Sanitizer
	box       *encrypthash.Blackbox
}

var defaultSanitizer = func() hy.Sanitizer {
	p := bluemonday.UGCPolicy()
	p.AllowStyling()
	p.AllowDataAttributes()
	// https://developer.mozilla.org/en-US/docs/Web/HTML/Element/form#attributes
	p.AllowElements("form")
	p.AllowAttrs("accept-charset", "autocomplete", "name", "rel", "action", "enctype", "method", "novalidate", "target").OnElements("form")
	// https://developer.mozilla.org/en-US/docs/Web/HTML/Element/input#attributes
	p.AllowElements("input")
	p.AllowAttrs(
		"accept", "alt", "autocomplete", "autofocus", "capture", "checked",
		"dirname", "disabled", "form", "formaction", "formenctype", "formmethod",
		"formnovalidate", "formtarget", "height", "list", "max", "maxlength", "min",
		"minlength", "multiple", "name", "pattern", "placeholder", "readonly",
		"required", "size", "src", "step", "type", "value", "width",
	).OnElements("input")
	// https://developer.mozilla.org/en-US/docs/Web/HTML/Element/button#attributes
	p.AllowElements("button")
	p.AllowAttrs(
		"autofocus", "disabled", "form", "formaction", "formenctype",
		"formmethod", "formnovalidate", "formtarget", "name", "type", "value",
	).OnElements("button")
	// https://developer.mozilla.org/en-US/docs/Web/HTML/Element/label#attributes
	p.AllowElements("label")
	p.AllowAttrs("for").OnElements("label")
	// https://developer.mozilla.org/en-US/docs/Web/HTML/Element/select#attributes
	p.AllowElements("select")
	p.AllowAttrs("autocomplete", "autofocus", "disabled", "form", "multiple", "name", "required", "size").OnElements("select")
	// https://developer.mozilla.org/en-US/docs/Web/HTML/Element/option#attributes
	p.AllowElements("option")
	p.AllowAttrs("disabled", "label", "selected", "value").OnElements("option")
	// https://developer.mozilla.org/en-US/docs/Web/HTML/Global_attributes/inputmode
	p.AllowAttrs("inputmode").Globally()
	// https://developer.mozilla.org/en-US/docs/Web/HTML/Element/link#attributes
	p.AllowElements("link")
	p.AllowAttrs(
		"as", "crossorigin", "disabled", "href", "hreflang", "imagesizes",
		"imagesrcset", "media", "rel", "sizes", "title", "type",
	).OnElements("link")
	p.AllowStandardURLs()
	// https://developer.mozilla.org/en-US/docs/Web/HTML/Element/script#attributes
	p.AllowElements("script")
	p.AllowAttrs("async", "crossorigin", "defer", "integrity", "nomodule", "nonce", "referrerpolicy", "src", "type").OnElements("script")

	p.AllowElements("svg")

	p.AllowImages()
	p.AllowLists()
	p.AllowTables()

	// settings which bluemonday loves to turn back on, leave it for the last
	p.RequireNoFollowOnLinks(false)
	return p
}()

var defaultHyforms = func() *Hyforms {
	key := make([]byte, 24)
	_, err := rand.Read(key)
	if err != nil {
		panic(err)
	}
	box, err := encrypthash.New(key, nil)
	if err != nil {
		panic(err)
	}
	return &Hyforms{sanitizer: defaultSanitizer, box: box}
}

func (hyf *Hyforms) MarshalForm(w http.ResponseWriter, r *http.Request, fn func(*Form)) (template.HTML, error) {
	form := &Form{
		request:      r,
		inputNames:   make(map[string]struct{}),
		inputErrMsgs: make(map[string][]string),
	}
	func() {
		c, _ := r.Cookie("hyforms.ValidationError")
		if c == nil {
			return
		}
		defer http.SetCookie(w, &http.Cookie{Name: "hyforms.ValidationError", MaxAge: -1})
		b, err := hyf.box.Base64VerifyHash(c.Value)
		if err != nil {
			return
		}
		fmt.Println(c.Value)
		validationErr := &ValidationError{}
		err = gob.NewDecoder(bytes.NewReader(b)).Decode(validationErr)
		if err != nil {
			return
		}
		if time.Now().After(validationErr.Expires) {
			return
		}
		form.formErrMsgs = validationErr.FormErrMsgs
		form.inputErrMsgs = validationErr.InputErrMsgs
	}()
	fn(form)
	if len(form.marshalErrMsgs) > 0 {
		return "", erro.Wrap(fmt.Errorf("marshal errors %v", form.marshalErrMsgs))
	}
	output, err := hy.MarshalElement(hyf.sanitizer, form)
	if err != nil {
		return output, erro.Wrap(err)
	}
	return output, nil
}

func (hyf *Hyforms) UnmarshalForm(w http.ResponseWriter, r *http.Request, fn func(*Form)) error {
	r.ParseForm()
	form := &Form{
		mode:         FormModeUnmarshal,
		request:      r,
		inputNames:   make(map[string]struct{}),
		inputErrMsgs: make(map[string][]string),
	}
	fn(form)
	if len(form.formErrMsgs) > 0 || len(form.inputErrMsgs) > 0 {
		validationErr := ValidationError{
			FormErrMsgs:  form.formErrMsgs,
			InputErrMsgs: form.inputErrMsgs,
			Expires:      time.Now().Add(5 * time.Second),
		}
		buf := &bytes.Buffer{}
		err := gob.NewEncoder(buf).Encode(validationErr)
		if err != nil {
			return fmt.Errorf("%w: failed gob encoding %s", &validationErr, err.Error())
		}
		value, err := hyf.box.Base64Hash(buf.Bytes())
		if err != nil {
			return erro.Wrap(err)
		}
		http.SetCookie(w, &http.Cookie{
			Name:   "hyforms.ValidationError",
			Value:  value,
			MaxAge: 5,
		})
		return &validationErr
	}
	return nil
}

func MarshalForm(s hy.Sanitizer, w http.ResponseWriter, r *http.Request, fn func(*Form)) (template.HTML, error) {
	form := &Form{
		request:      r,
		inputNames:   make(map[string]struct{}),
		inputErrMsgs: make(map[string][]string),
	}
	func() {
		c, _ := r.Cookie("hyforms.ValidationError")
		if c == nil {
			return
		}
		defer http.SetCookie(w, &http.Cookie{Name: "hyforms.ValidationError", MaxAge: -1})
		b, err := base64.RawURLEncoding.DecodeString(c.Value)
		if err != nil {
			return
		}
		fmt.Println(c.Value)
		validationErr := &ValidationError{}
		err = gob.NewDecoder(bytes.NewReader(b)).Decode(validationErr)
		if err != nil {
			return
		}
		if time.Now().After(validationErr.Expires) {
			return
		}
		form.formErrMsgs = validationErr.FormErrMsgs
		form.inputErrMsgs = validationErr.InputErrMsgs
	}()
	fn(form)
	if len(form.marshalErrMsgs) > 0 {
		return "", erro.Wrap(fmt.Errorf("marshal errors %v", form.marshalErrMsgs))
	}
	output, err := hy.MarshalElement(s, form)
	if err != nil {
		return output, erro.Wrap(err)
	}
	return output, nil
}

func UnmarshalForm(w http.ResponseWriter, r *http.Request, fn func(*Form)) error {
	r.ParseForm()
	form := &Form{
		mode:         FormModeUnmarshal,
		request:      r,
		inputNames:   make(map[string]struct{}),
		inputErrMsgs: make(map[string][]string),
	}
	fn(form)
	if len(form.formErrMsgs) > 0 || len(form.inputErrMsgs) > 0 {
		validationErr := ValidationError{
			FormErrMsgs:  form.formErrMsgs,
			InputErrMsgs: form.inputErrMsgs,
			Expires:      time.Now().Add(5 * time.Second),
		}
		buf := &bytes.Buffer{}
		err := gob.NewEncoder(buf).Encode(validationErr)
		if err != nil {
			return fmt.Errorf("%w: failed gob encoding %s", &validationErr, err.Error())
		}
		http.SetCookie(w, &http.Cookie{
			Name:   "hyforms.ValidationError",
			Value:  base64.RawURLEncoding.EncodeToString(buf.Bytes()),
			MaxAge: 5,
		})
		return &validationErr
	}
	return nil
}

func caller(skip int) (file string, line int, function string) {
	var pc [1]uintptr
	n := runtime.Callers(skip+2, pc[:])
	if n == 0 {
		return "???", 1, "???"
	}
	frame, _ := runtime.CallersFrames(pc[:n]).Next()
	return frame.File, frame.Line, frame.Function
}

func validateInput(f *Form, inputName string, value interface{}, validators []Validator) {
	if len(validators) == 0 {
		return
	}
	var stop bool
	var errMsg string
	ctx := f.request.Context()
	ctx = context.WithValue(ctx, ctxKeyName, inputName)
	for _, validator := range validators {
		stop, errMsg = validator(ctx, value)
		if errMsg != "" {
			f.inputErrMsgs[inputName] = append(f.inputErrMsgs[inputName], errMsg)
		}
		if stop {
			return
		}
	}
}
