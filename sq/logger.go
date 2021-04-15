package sq

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// LogFlag is a flag that affects the verbosity of the Logger output.
type LogFlag int

// LogFlags
const (
	Linterpolate LogFlag = 0b1     // Interpolate the args into the query
	Lmultiline   LogFlag = 0b10    // Show the query before and after interpolation
	Lcaller      LogFlag = 0b100   // Show caller information i.e. filename, line number, function name
	Lresults     LogFlag = 0b1000  // Show the first 5 results if applicable. Lmultiline must be enabled.
	Lcolor       LogFlag = 0b10000 // Colorize log output
	Lverbose     LogFlag = Lmultiline | Lcaller | Lresults | Lcolor
	Lcompact     LogFlag = Linterpolate | Lcaller | Lcolor
)

// ExecFlag is a flag that affects the behavior of Exec.
type ExecFlag int

// ExecFlags
const (
	ExecActive    ExecFlag = 0b1   // Used by Logger to discern between Fetch and Exec queries
	ElastInsertID ExecFlag = 0b10  // Get last inserted ID
	ErowsAffected ExecFlag = 0b100 // Get number of rows affected
)

var (
	colorReset  = "\x1b[0m"
	colorRed    = "\x1b[31m"
	colorGreen  = "\x1b[32m"
	colorYellow = "\x1b[33m"
	colorBlue   = "\x1b[34m"
	colorPurple = "\x1b[35m"
	colorCyan   = "\x1b[36m"
	colorGray   = "\x1b[37m"
	colorWhite  = "\x1b[97m"
)

func init() {
	if runtime.GOOS == "windows" {
		colorReset = ""
		colorRed = ""
		colorGreen = ""
		colorYellow = ""
		colorBlue = ""
		colorPurple = ""
		colorCyan = ""
		colorGray = ""
		colorWhite = ""
	}
}

type QueryStats struct {
	Dialect        string
	Query          string
	Args           []interface{}
	Error          error
	FunctionName   string
	Filename       string
	LineNumber     int
	RowCount       int64
	RowsAffected   int64
	ResultsPreview string
	LastInsertID   int64
	ExecFlag       ExecFlag
	LogFlag        LogFlag
	TimeTaken      time.Duration
}

type Logger interface {
	LogQueryStats(context.Context, QueryStats)
}

type logger struct {
	*log.Logger
}

var defaultlogger = logger{log.New(os.Stdout, "", log.LstdFlags)}

func DefaultLogger() Logger { return defaultlogger }

func (lg logger) LogQueryStats(ctx context.Context, stats QueryStats) {
	select {
	case <-ctx.Done():
		return
	default:
	}
	buf := bufpool.Get().(*strings.Builder)
	defer func() {
		if buf.Len() > 0 {
			lg.Printf(buf.String())
		}
		buf.Reset()
		bufpool.Put(buf)
	}()
	if stats.Error == nil {
		buf.WriteString(colorGreen + "[OK]" + colorReset)
	} else {
		buf.WriteString(colorRed + "[FAIL]" + colorReset)
	}
	if Lmultiline&stats.LogFlag == 0 {
		// Log one-liner
		if Linterpolate&stats.LogFlag != 0 {
			if stats.Dialect == "postgres" {
				buf.WriteString(" " + DollarInterpolate(stats.Query, stats.Args...))
			} else {
				buf.WriteString(" " + QuestionInterpolate(stats.Query, stats.Args...))
			}
		} else {
			buf.WriteString(" " + stats.Query + " " + fmt.Sprint(stats.Args))
		}
		buf.WriteString(" |")
	}
	buf.WriteString(colorBlue + " timeTaken" + colorReset + "=" + stats.TimeTaken.String())
	if stats.ExecFlag == 0 {
		buf.WriteString(colorBlue + " rowCount" + colorReset + "=" + strconv.FormatInt(stats.RowCount, 10))
	} else {
		if ErowsAffected&stats.ExecFlag != 0 {
			buf.WriteString(colorBlue + " rowsAffected" + colorReset + "=" + strconv.FormatInt(stats.RowsAffected, 10))
		}
		if ElastInsertID&stats.ExecFlag != 0 {
			buf.WriteString(colorBlue + " lastInsertID" + colorReset + "=" + strconv.FormatInt(stats.LastInsertID, 10))
		}
	}
	if Lcaller&stats.LogFlag != 0 {
		buf.WriteString(colorBlue + " caller" + colorReset + "=" + stats.Filename + ":" + strconv.Itoa(stats.LineNumber) + ":" + filepath.Base(stats.FunctionName))
	}
	if Lmultiline&stats.LogFlag != 0 {
		// Log multiline
		buf.WriteString("\n" + colorPurple + "----[ Executing query ]----" + colorReset)
		buf.WriteString("\n" + stats.Query + " " + fmt.Sprint(stats.Args))
		buf.WriteString("\n" + colorPurple + "----[ with bind values ]----" + colorReset)
		if stats.Dialect == "postgres" {
			buf.WriteString("\n" + DollarInterpolate(stats.Query, stats.Args...))
		} else {
			buf.WriteString("\n" + QuestionInterpolate(stats.Query, stats.Args...))
		}
	}
	if Lresults&stats.LogFlag != 0 && stats.ResultsPreview != "" {
		buf.WriteString("\n" + colorPurple + "----[ Fetched result ]----" + colorReset)
		buf.WriteString(stats.ResultsPreview)
	}
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
