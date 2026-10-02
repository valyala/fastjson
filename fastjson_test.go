package fastjson

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestCompact(t *testing.T) {
	fixtures := []struct {
		name string
		json string
	}{
		{"small", smallFixture},
		{"medium", mediumFixture},
		{"large", largeFixture},
		{"canada", canadaFixture},
		{"citm", citmFixture},
		{"twitter", twitterFixture},
		{"empty-object", "{}"},
		{"empty-array", "[]"},
		{"whitespace-empty-object", "  { \n\t }  "},
		{"whitespace-empty-array", "  [ \n\t ]  "},
		{"number-int", "  123456  "},
		{"number-zero", "  0  "},
		{"number-negative", "  -42  "},
		{"number-float", "  -0.01e+006  "},
		{"string-simple", `  "hello world"  `},
		{"string-empty", `  ""  `},
		{"string-escapes", `  "\" \\ \/ \b \f \n \r \t \u0020"  `},
		{"string-escaped-backslash-quote", `  "\\\""  `},
		{"string-escaped-backslashes", `  "\\\\"  `},
		{"string-spaces", `  "   spaced   content   "  `},
		{"boolean-true", "  true  "},
		{"boolean-false", "  false  "},
		{"null", "  null  "},
		{"nested-structures", ` { "a" : [ 1 , { "b" : "c" , "d" : [ true , false , null ] } ] } `},
		{"deeply-nested-array", strings.Repeat("[", 50) + "1" + strings.Repeat("]", 50)},
		{"deeply-nested-object", strings.Repeat(`{"a":`, 50) + `1` + strings.Repeat("}", 50)},
	}

	for _, tc := range fixtures {
		t.Run(tc.name, func(t *testing.T) {
			src := []byte(tc.json)

			var stdBuf bytes.Buffer
			if err := json.Compact(&stdBuf, src); err != nil {
				t.Fatalf("json.Compact failed: %s", err)
			}
			expected := stdBuf.Bytes()

			// Test fresh destination
			got, err := Compact(nil, src)
			if err != nil {
				t.Fatalf("Compact failed: %s", err)
			}
			if !bytes.Equal(got, expected) {
				t.Fatalf("output mismatch:\ngot:      %q\nexpected: %q", got, expected)
			}

			// Test pre-allocated destination with existing prefix
			prefix := []byte("prefix:")
			dstPrefixed := append([]byte(nil), prefix...)
			gotPrefixed, err := Compact(dstPrefixed, src)
			if err != nil {
				t.Fatalf("Compact with prefix failed: %s", err)
			}
			if !bytes.HasPrefix(gotPrefixed, prefix) {
				t.Fatalf("prefix missing in result")
			}
			if !bytes.Equal(gotPrefixed[len(prefix):], expected) {
				t.Fatalf("prefixed output mismatch:\ngot:      %q\nexpected: %q", gotPrefixed[len(prefix):], expected)
			}

			// Test in-place compaction
			srcCopy := append([]byte(nil), src...)
			gotInPlace, err := Compact(srcCopy[:0], srcCopy)
			if err != nil {
				t.Fatalf("Compact in-place failed: %s", err)
			}
			if !bytes.Equal(gotInPlace, expected) {
				t.Fatalf("in-place output mismatch:\ngot:      %q\nexpected: %q", gotInPlace, expected)
			}
		})
	}
}

func TestCompactMalformed(t *testing.T) {
	malformedCases := []string{
		"",
		"   ",
		"\t\r\n",
		"{",
		"}",
		"[",
		"]",
		`{"foo"`,
		`{"foo":`,
		`{"foo": 1,}`,
		`{"foo": 1 "bar": 2}`,
		`[1, 2,]`,
		`[1, 2`,
		`"unclosed string`,
		`"control char` + "\x00" + `"`,
		`"control char` + "\x1f" + `"`,
		`"invalid escape \q"`,
		"1.",
		"01",
		"-",
		"1e",
		"1e+",
		"--1",
		"+1",
		"truee",
		"fals",
		"nul",
		"undefined",
		"NaN",
		`{"a": 1} trailing`,
		`[1, 2] 3`,
		`1 2`,
		`"foo" "bar"`,
	}

	for i, tc := range malformedCases {
		src := []byte(tc)

		// Ensure dst is preserved untouched on error
		prefix := []byte("keep-this-prefix")
		dst := append([]byte(nil), prefix...)
		res, err := Compact(dst, src)
		if err == nil {
			t.Errorf("#%d %q: expecting non-nil error, got compacted: %q", i, tc, res)
		}
		if !bytes.Equal(res, prefix) {
			t.Errorf("#%d %q: dst corrupted on error: got %q, expected %q", i, tc, res, prefix)
		}
	}
}
