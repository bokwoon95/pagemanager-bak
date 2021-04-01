package sq

import "strings"

type Window struct {
	WindowName        string
	RenderName        bool
	PartitionByFields Fields
	OrderByFields     Fields
	FrameDefinition   string
}

func (w Window) AppendSQL(dialect string, buf *strings.Builder, args *[]interface{}, params map[string]int) error {
	if w.RenderName {
		buf.WriteString(w.WindowName)
		return nil
	}
	buf.WriteString("(")
	var written bool
	if len(w.PartitionByFields) > 0 {
		buf.WriteString("PARTITION BY ")
		w.PartitionByFields.AppendSQLExclude(dialect, buf, args, nil, nil)
		written = true
	}
	if len(w.OrderByFields) > 0 {
		if written {
			buf.WriteString(" ")
		}
		buf.WriteString("ORDER BY ")
		w.OrderByFields.AppendSQLExclude(dialect, buf, args, nil, nil)
		written = true
	}
	if w.FrameDefinition != "" {
		if written {
			buf.WriteString(" ")
		}
		buf.WriteString(w.FrameDefinition)
	}
	buf.WriteString(")")
	return nil
}

func (w Window) As(name string) Window {
	w.WindowName = name
	return w
}

func (w Window) Name() Window {
	if w.WindowName == "" {
		w.WindowName = randomString(8)
	}
	w.RenderName = true
	return w
}

func PartitionBy(fields ...Field) Window {
	return Window{PartitionByFields: fields}
}

func OrderBy(fields ...Field) Window {
	return Window{OrderByFields: fields}
}

func (w Window) PartitionBy(fields ...Field) Window {
	w.PartitionByFields = fields
	return w
}

func (w Window) OrderBy(fields ...Field) Window {
	w.OrderByFields = fields
	return w
}

func (w Window) Frame(frameDefinition string) Window {
	w.FrameDefinition = frameDefinition
	return w
}

type Windows []Window

func (ws Windows) AppendSQL(dialect string, buf *strings.Builder, args *[]interface{}, params map[string]int) error {
	for i, window := range ws {
		if i > 0 {
			buf.WriteString(", ")
		}
		if window.WindowName != "" {
			buf.WriteString(window.WindowName)
		} else {
			buf.WriteString(randomString(8))
		}
		buf.WriteString(" AS ")
		window.AppendSQL(dialect, buf, args, nil)
	}
	return nil
}
