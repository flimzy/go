package json

import (
	"io"
)

type indentWriter struct {
	dst            io.Writer
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
	return &indentWriter{
		dst:    w,
		prefix: prefix,
		indent: indent,
		scan:   scan,
	}
}

// indentStream implements the same logic as Indent, except that the output is
// not rewound in case of an error. src is completely consumed by this function.
func (w *indentWriter) Write(src []byte) (int, error) {
	var n int
	for _, c := range src {
		w.scan.bytes++
		n++
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
			w.newline()
		}

		// Emit semantically uninteresting bytes
		// (in particular, punctuation in strings) unmodified.
		if v == scanContinue {
			w.dst.Write([]byte{c})
			continue
		}

		// Add spacing around real punctuation.
		switch c {
		case '{', '[':
			// delay indent so that empty object and array are formatted as {} and [].
			w.needIndent = true
			w.dst.Write([]byte{c})

		case ',':
			w.dst.Write([]byte{c})
			w.newline()

		case ':':
			w.dst.Write([]byte{c})
			w.dst.Write([]byte{' '})

		case '}', ']':
			if w.needIndent {
				// suppress indent in empty object/array
				w.needIndent = false
			} else {
				w.depth--
				w.newline()
			}
			w.dst.Write([]byte{c})

		default:
			w.dst.Write([]byte{c})
		}
	}
	if w.scan.eof() == scanError {
		return n, w.scan.err
	}
	return n, nil
}

func (w *indentWriter) newline() {
	w.dst.Write([]byte{'\n'})
	w.dst.Write([]byte(w.prefix))
	for i := 0; i < w.depth; i++ {
		w.dst.Write([]byte(w.indent))
	}
}
