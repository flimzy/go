package json

import (
	"io"
)

type indentWriter struct {
	dst            writer
	prefix, indent string
	err            error
	scan           scanner
	needIndent     bool
	depth          int
}

var _ io.Writer = &indentWriter{}

// IndentWriter wraps w, re-indenting the data written to it, according to
// prefix and indent. If any error occurs, it will be returned on the next
// call to Write() on the returned io.Writer.
func IndentWriter(w io.Writer, prefix, indent string) io.Writer {
	var scan scanner
	scan.reset()
	dst, ok := w.(writer)
	if !ok {
		dst = &convertWriter{w}
	}
	return &indentWriter{
		dst:    dst,
		prefix: prefix,
		indent: indent,
		scan:   scan,
	}
}

// indentStream implements the same logic as Indent, except that the output is
// not rewound in case of an error. src is completely consumed by this function.
func (w *indentWriter) Write(src []byte) (int, error) {
	n := w.scan.bytes
	for _, c := range src {
		w.scan.bytes++
		v := w.scan.step(&w.scan, c)
		if v == scanSkipSpace {
			continue
		}
		if v == scanError {
			break
		}
		if w.needIndent && v != scanEndObject && v != scanEndArray {
			w.needIndent = false
			w.depth++
			// newNewline(w.dst, w.prefix, w.indent, w.depth)
			w.newline()
		}

		// Emit semantically uninteresting bytes
		// (in particular, punctuation in strings) unmodified.
		if v == scanContinue {
			w.dst.WriteByte(c)
			continue
		}

		// Add spacing around real punctuation.
		switch c {
		case '{', '[':
			// delay indent so that empty object and array are formatted as {} and [].
			w.needIndent = true
			w.dst.WriteByte(c)

		case ',':
			w.dst.WriteByte(c)
			// newNewline(w.dst, w.prefix, w.indent, w.depth)
			w.newline()

		case ':':
			w.dst.WriteByte(c)
			w.dst.WriteByte(' ')

		case '}', ']':
			if w.needIndent {
				// suppress indent in empty object/array
				w.needIndent = false
			} else {
				w.depth--
				// newNewline(w.dst, w.prefix, w.indent, w.depth)
				w.newline()
			}
			w.dst.WriteByte(c)

		default:
			w.dst.WriteByte(c)
		}
	}
	if w.scan.eof() == scanError {
		return int(w.scan.bytes - n), w.scan.err
	}
	return int(w.scan.bytes - n), nil
}

func newNewline(dst writer, prefix, indent string, depth int) {
	dst.WriteByte('\n')
	dst.WriteString(prefix)
	for i := 0; i < depth; i++ {
		dst.WriteString(indent)
	}
}

func (w *indentWriter) newline() {
	w.dst.WriteByte('\n')
	w.dst.WriteString(w.prefix)
	for i := 0; i < w.depth; i++ {
		w.dst.WriteString(w.indent)
	}
}

type writer interface {
	Write([]byte) (int, error)
	WriteString(string) (int, error)
	WriteByte(byte) error
}

type convertWriter struct {
	io.Writer
}

func (c convertWriter) WriteString(s string) (int, error) {
	return io.WriteString(c, s)
}
func (c convertWriter) WriteByte(b byte) error {
	_, err := c.Write([]byte{b})
	return err
}
