package fastjson

import (
	"bytes"
	"fmt"
	"testing"
)

func TestScanner(t *testing.T) {
	var sc Scanner

	t.Run("success", func(t *testing.T) {
		sc.InitBytes([]byte(`[] {} "" 123`))
		var bb bytes.Buffer
		for sc.Next() {
			v := sc.Value()
			fmt.Fprintf(&bb, "%s", v)
		}
		if err := sc.Error(); err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		s := bb.String()
		if s != `[]{}""123` {
			t.Fatalf("unexpected string obtained; got %q; want %q", s, `[]{}""123`)
		}
	})

	t.Run("error", func(t *testing.T) {
		sc.Init(`[] sdfdsfdf`)
		for sc.Next() {
		}
		if err := sc.Error(); err == nil {
			t.Fatalf("expecting non-nil error")
		}
		if sc.Next() {
			t.Fatalf("Next must return false")
		}
	})
}
