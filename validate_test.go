package fastjson

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestValidateSimple(t *testing.T) {
	if err := Validate(`123`); err != nil {
		t.Fatalf("cannot validate number: %s", err)
	}
	if err := Validate(`"foobar"`); err != nil {
		t.Fatalf("cannot validate string: %s", err)
	}
	if err := Validate(`null`); err != nil {
		t.Fatalf("cannot validate null: %s", err)
	}
	if err := Validate(`true`); err != nil {
		t.Fatalf("cannot validate true: %s", err)
	}
	if err := Validate(`false`); err != nil {
		t.Fatalf("cannot validate false: %s", err)
	}
	if err := Validate(`foobar`); err == nil {
		t.Fatalf("validation unexpectedly passed")
	}
	if err := Validate(`XDF`); err == nil {
		t.Fatalf("validation unexpectedly passed")
	}

	if err := ValidateBytes([]byte(`{"foo":["bar", 123]}`)); err != nil {
		t.Fatalf("cannot validate valid JSON: %s", err)
	}
	if err := ValidateBytes([]byte(`{"foo": bar`)); err == nil {
		t.Fatalf("validation unexpectedly passed")
	}
}

func TestValidateNumberZeroLen(t *testing.T) {
	tail, err := validateNumber("")
	if err == nil {
		t.Fatalf("expecting non-nil error")
	}
	if tail != "" {
		t.Fatalf("unexpected non-empty tail: %q", tail)
	}
}

func TestValidate(t *testing.T) {
	var tests = []string{
		"",
		"   ",
		" z",
		" 1  1",
		" 1  {}",
		" 1  []",
		" 1  true",
		" 1  null",
		" 1  \"n\"",

		// string
		`"foo"`,
		"\"\xe2\x80\xa8\xe2\x80\xa9\"", // line-sep and paragraph-sep
		` "\uaaaa" `,
		`"\uz"`,
		` "\`,
		` "\z`,
		" \"f\x00o\"",  // control char
		"\"foo\nbar\"", // control char
		`"foo\qw"`,     // unknown escape sequence
		` "foo`,
		` "\uazaa" `,
		`"\"\\\/\b\f\n\r\t"`,

		// number
		"1",
		"  0 ",
		" 0e1 ",
		" 0e+0 ",
		" -0e+0 ",
		"-0",
		"1e6",
		"1e+6",
		"-1e+6",
		"-0e+6",
		" -103e+1 ",
		"-0.01e+006",
		"-z",
		"-",
		"1e",
		"1e+",
		" 03e+1 ",
		" 1e.1 ",
		" 00 ",
		"1.e3",
		"01e+6",
		"-0.01e+0.6",
		"123.",
		"123.345",
		"001 ",
		"001",

		// object
		"{}",
		`{"foo": 3}`,
		"{\"f\x00oo\": 3}",
		`{"foo\WW": 4}`, // unknown escape sequence
		`{"foo": 3 "bar"}`,
		` {}    `,
		strings.Repeat(`{"f":`, 1000) + "{}" + strings.Repeat("}", 1000),
		`{"foo": [{"":3, "4": "3"}, 4, {}], "t_wo": 1}`,
		` {"foo": 2,"fudge}`,
		`{{"foo": }}`,
		`{{"foo": [{"":3, 4: "3"}, 4, "5": {4}]}, "t_wo": 1}`,
		"{",
		`{"foo"`,
		`{"foo",f}`,
		`{"foo",`,
		`{"foo"f`,
		"{}}",
		`{"foo": 234`,
		`{"foo\"bar": 123}`,
		"{\n\t\"foo\"  \n\b\f: \t123}",

		// array
		`[]`,
		`[ 1, {}]`,
		strings.Repeat("[", 1000) + strings.Repeat("]", 1000),
		`[1, 2, 3, 4, {}]`,
		`[`,
		`[1,`,
		`[1a`,
		`[]]`,
		`[1  `,

		// boolean
		"true",
		"   true ",
		"tree",
		"false",
		"  true f",
		"fals",
		"falsee",

		// null
		"null ",
		" null ",
		" nulll ",
		"no",
	}
	for i, test := range tests {
		in := []byte(test)
		got := ValidateBytes(in) == nil
		exp := json.Valid(in)

		if got != exp {
			t.Errorf("#%d: %q got valid? %v, exp? %v", i, in, got, exp)
		}
	}
}
