// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package json

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func TestStreamIndent(t *testing.T) {
	var buf bytes.Buffer
	for _, tt := range examples {
		buf.Reset()
		w := IndentWriter(&buf, "", "\t")
		_, err := io.Copy(w, strings.NewReader(tt.indent))
		if err != nil {
			t.Errorf("Indent(%#q): %v", tt.indent, err)
		} else if s := buf.String(); s != tt.indent {
			t.Errorf("Indent(%#q) = %#q, want original", tt.indent, s)
		}

		buf.Reset()
		w = IndentWriter(&buf, "", "\t")
		_, err = io.Copy(w, strings.NewReader(tt.compact))
		if err != nil {
			t.Errorf("Indent(%#q): %v", tt.compact, err)
			continue
		} else if s := buf.String(); s != tt.indent {
			t.Errorf("Indent(%#q) = %#q, want %#q", tt.compact, s, tt.indent)
		}
	}
}
