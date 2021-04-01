package testutil

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type I struct {
	t            *testing.T
	out          io.Writer
	failFast     bool
	noDefaults   bool
	showFunction bool
	parallel     bool
}

func New(t *testing.T, opts ...Option) I {
	is := I{t: t, out: os.Stdout}
	for _, opt := range opts {
		opt(&is)
	}
	if is.parallel {
		t.Parallel()
	}
	return is
}

func (is I) New(t *testing.T, opts ...Option) I {
	is.t = t
	for _, opt := range opts {
		opt(&is)
	}
	if is.parallel {
		t.Parallel()
	}
	return is
}

type Option func(*I)

func FailFast(is *I)   { is.failFast = true }
func NoDefaults(is *I) { is.noDefaults = true }
func Parallel(is *I)   { is.parallel = true }
func SetOutput(w io.Writer) Option {
	return func(is *I) { is.out = w }
}

func (is I) fail() {
	if is.failFast {
		is.t.FailNow()
	} else {
		is.t.Fail()
	}
}

func (is I) True(expression bool) {
	if expression {
		return
	}
	io.WriteString(is.out, "["+callers(1).stringify(is.showFunction)+"] not true\n")
	is.fail()
}

func (is I) Equal(x, y interface{}, opts ...cmp.Option) {
	if !is.noDefaults {
		opts = append([]cmp.Option{
			cmp.Exporter(func(typ reflect.Type) bool { return true }),
		}, opts...)
	}
	diff := cmp.Diff(x, y, opts...)
	if diff == "" {
		return
	}
	io.WriteString(is.out, "["+callers(1).stringify(is.showFunction)+"] diff -x +y\n"+diff)
	is.fail()
}

func (is I) NoErr(err error) {
	if err == nil {
		return
	}
	io.WriteString(is.out, "["+callers(1).stringify(is.showFunction)+"] err: "+err.Error()+"\n")
	is.fail()
}

func (is I) Fail() {
	io.WriteString(is.out, "["+callers(1).stringify(is.showFunction)+"] failed")
	is.fail()
}

func callers(skip int) callsites {
	var pc [50]uintptr
	// Skip two extra frames to account for this function
	// and runtime.Callers itself.
	n := runtime.Callers(skip+2, pc[:])
	if n == 0 {
		panic("is: zero callers found")
	}
	var sites []callsite
	frames := runtime.CallersFrames(pc[:n])
	for frame, more := frames.Next(); more; frame, more = frames.Next() {
		sites = append(sites, callsite{function: frame.Function, file: frame.File, line: frame.Line})
	}
	for i := len(sites)/2 - 1; i >= 0; i-- {
		opp := len(sites) - 1 - i
		sites[i], sites[opp] = sites[opp], sites[i]
	}
	return sites
}

type callsite struct {
	function string
	file     string
	line     int
}

type callsites []callsite

func (sites callsites) stringify(showFunction bool) string {
	buf := &strings.Builder{}
	for i, site := range sites {
		if i == 0 {
			continue
		}
		if i > 1 {
			buf.WriteString(" -> ")
		}
		buf.WriteString(filepath.Base(site.file) + ":" + strconv.Itoa(site.line))
		if showFunction {
			buf.WriteString(":" + filepath.Base(site.function))
		}
	}
	return buf.String()
}
