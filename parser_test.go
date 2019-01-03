package fastjson

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestParseRawNumber(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		f := func(s, expectedRN, expectedTail string) {
			t.Helper()

			rn, tail, err := parseRawNumber(s)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}
			if rn != expectedRN {
				t.Fatalf("unexpected raw number; got %q; want %q", rn, expectedRN)
			}
			if tail != expectedTail {
				t.Fatalf("unexpected tail; got %q; want %q", tail, expectedTail)
			}
		}

		f("0", "0", "")
		f("0tail", "0", "tail")
		f("123", "123", "")
		f("123tail", "123", "tail")
		f("-123tail", "-123", "tail")
		f("-12.345tail", "-12.345", "tail")
		f("-12.345e67tail", "-12.345e67", "tail")
		f("-12.345E+67 tail", "-12.345E+67", " tail")
		f("-12.345E-67,tail", "-12.345E-67", ",tail")
		f("-1234567.8e+90tail", "-1234567.8e+90", "tail")
	})

	t.Run("error", func(t *testing.T) {
		f := func(s, expectedTail string) {
			t.Helper()

			_, tail, err := parseRawNumber(s)
			if err == nil {
				t.Fatalf("expecting non-nil error")
			}
			if tail != expectedTail {
				t.Fatalf("unexpected tail; got %q; want %q", tail, expectedTail)
			}
		}

		f("xyz", "xyz")
		f(" ", " ")
		f("[", "[")
		f(",", ",")
		f("{", "{")
		f("\"", "\"")
	})
}

func TestUnescapeStringBestEffort(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		testUnescapeStringBestEffort(t, ``, ``)
		testUnescapeStringBestEffort(t, `\"`, `"`)
		testUnescapeStringBestEffort(t, `\\`, `\`)
		testUnescapeStringBestEffort(t, `\\\"`, `\"`)
		testUnescapeStringBestEffort(t, `\\\"абв`, `\"абв`)
		testUnescapeStringBestEffort(t, `йцук\n\"\\Y`, "йцук\n\"\\Y")
		testUnescapeStringBestEffort(t, `q\u1234we`, "q\u1234we")
		testUnescapeStringBestEffort(t, `п\ud83e\udd2dи`, "п🤭и")
	})

	t.Run("error", func(t *testing.T) {
		testUnescapeStringBestEffort(t, `\`, ``)
		testUnescapeStringBestEffort(t, `foo\qwe`, `foo\qwe`)
		testUnescapeStringBestEffort(t, `\"x\uyz\"`, `"x\uyz"`)
		testUnescapeStringBestEffort(t, `\u12\"пролw`, `\u12"пролw`)
		testUnescapeStringBestEffort(t, `п\ud83eи`, "п\\ud83eи")
	})
}

func testUnescapeStringBestEffort(t *testing.T, s, expectedS string) {
	t.Helper()

	// unescapeString modifies the original s, so call it
	// on a byte slice copy.
	b := append([]byte{}, s...)
	us := unescapeStringBestEffort(b2s(b))
	if us != expectedS {
		t.Fatalf("unexpected unescaped string; got %q; want %q", us, expectedS)
	}
}

func TestParseRawString(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		f := func(s, expectedRS, expectedTail string) {
			t.Helper()

			rs, tail, err := parseRawString(s[1:])
			if err != nil {
				t.Fatalf("unexpected error on parseRawString: %s", err)
			}
			if rs != expectedRS {
				t.Fatalf("unexpected string on parseRawString; got %q; want %q", rs, expectedRS)
			}
			if tail != expectedTail {
				t.Fatalf("unexpected tail on parseRawString; got %q; want %q", tail, expectedTail)
			}

			// parseRawKey results must be identical to parseRawString.
			rs, tail, err = parseRawKey(s[1:])
			if err != nil {
				t.Fatalf("unexpected error on parseRawKey: %s", err)
			}
			if rs != expectedRS {
				t.Fatalf("unexpected string on parseRawKey; got %q; want %q", rs, expectedRS)
			}
			if tail != expectedTail {
				t.Fatalf("unexpected tail on parseRawKey; got %q; want %q", tail, expectedTail)
			}
		}

		f(`""`, "", "")
		f(`""xx`, "", "xx")
		f(`"foobar"`, "foobar", "")
		f(`"foobar"baz`, "foobar", "baz")
		f(`"\""`, `\"`, "")
		f(`"\""tail`, `\"`, "tail")
		f(`"\\"`, `\\`, "")
		f(`"\\"tail`, `\\`, "tail")
		f(`"x\\"`, `x\\`, "")
		f(`"x\\"tail`, `x\\`, "tail")
		f(`"x\\y"`, `x\\y`, "")
		f(`"x\\y"tail`, `x\\y`, "tail")
		f(`"\\\"й\n\"я"tail`, `\\\"й\n\"я`, "tail")
		f(`"\\\\\\\\"tail`, `\\\\\\\\`, "tail")
	})

	t.Run("error", func(t *testing.T) {
		f := func(s, expectedTail string) {
			t.Helper()

			_, tail, err := parseRawString(s[1:])
			if err == nil {
				t.Fatalf("expecting non-nil error on parseRawString")
			}
			if tail != expectedTail {
				t.Fatalf("unexpected tail on parseRawString; got %q; want %q", tail, expectedTail)
			}

			// parseRawKey results must be identical to parseRawString.
			_, tail, err = parseRawKey(s[1:])
			if err == nil {
				t.Fatalf("expecting non-nil error on parseRawKey")
			}
			if tail != expectedTail {
				t.Fatalf("unexpected tail on parseRawKey; got %q; want %q", tail, expectedTail)
			}
		}

		f(`"`, "")
		f(`"unclosed string`, "")
		f(`"\"`, "")
		f(`"\"unclosed`, "")
		f(`"foo\\\\\"тест\n\r\t`, "")
	})
}

func TestParserPool(t *testing.T) {
	var pp ParserPool
	for i := 0; i < 10; i++ {
		p := pp.Get()
		if _, err := p.Parse("null"); err != nil {
			t.Fatalf("cannot parse null: %s", err)
		}
		pp.Put(p)
	}
}

func TestValueInvalidTypeConversion(t *testing.T) {
	var p Parser

	v, err := p.Parse(`[{},[],"",123.45,true,null]`)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	a := v.GetArray()

	// object
	_, err = a[0].Object()
	if err != nil {
		t.Fatalf("unexpected error when obtaining object: %s", err)
	}
	_, err = a[0].Array()
	if err == nil {
		t.Fatalf("expecting non-nil error when trying to obtain array from object")
	}

	// array
	_, err = a[1].Array()
	if err != nil {
		t.Fatalf("unexpected error when obtaining array: %s", err)
	}
	_, err = a[1].Object()
	if err == nil {
		t.Fatalf("expecting non-nil error when trying to obtain object from array")
	}

	// string
	_, err = a[2].StringBytes()
	if err != nil {
		t.Fatalf("unexpected error when obtaining string: %s", err)
	}
	_, err = a[2].Int()
	if err == nil {
		t.Fatalf("expecting non-nil error when trying to obtain int from string")
	}

	// number
	_, err = a[3].Float64()
	if err != nil {
		t.Fatalf("unexpected error when obtaining float64: %s", err)
	}
	_, err = a[3].StringBytes()
	if err == nil {
		t.Fatalf("expecting non-nil error when trying to obtain string from number")
	}

	// true
	_, err = a[4].Bool()
	if err != nil {
		t.Fatalf("unexpected error when obtaining bool: %s", err)
	}
	_, err = a[4].StringBytes()
	if err == nil {
		t.Fatalf("expecting non-nil error when trying to obtain string from bool")
	}

	// null
	_, err = a[5].Bool()
	if err == nil {
		t.Fatalf("expecting non-nil error when trying to obtain bool from null")
	}
}

func TestValueGetTyped(t *testing.T) {
	var p Parser

	v, err := p.Parse(`{"foo": 123, "bar": "433", "baz": true, "obj":{}, "arr":[1,2,3],
		"zero_float1": 0.00,
		"zero_float2": -0e123
	}`)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if !v.Exists("foo") {
		t.Fatalf("foo must exist in the v")
	}
	if v.Exists("foo", "bar") {
		t.Fatalf("foo.bar mustn't exist in the v")
	}
	if v.Exists("foobar") {
		t.Fatalf("foobar mustn't exist in the v")
	}

	o := v.GetObject("obj")
	os := o.String()
	if os != "{}" {
		t.Fatalf("unexpected object; got %s; want %s", os, "{}")
	}
	o = v.GetObject("arr")
	if o != nil {
		t.Fatalf("unexpected non-nil object: %s", o)
	}
	o = v.GetObject("foo", "bar")
	if o != nil {
		t.Fatalf("unexpected non-nil object: %s", o)
	}
	a := v.GetArray("arr")
	if len(a) != 3 {
		t.Fatalf("unexpected array len; got %d; want %d", len(a), 3)
	}
	a = v.GetArray("obj")
	if a != nil {
		t.Fatalf("unexpected non-nil array: %s", a)
	}
	a = v.GetArray("foo", "bar")
	if a != nil {
		t.Fatalf("unexpected non-nil array: %s", a)
	}
	n := v.GetInt("foo")
	if n != 123 {
		t.Fatalf("unexpected value; got %d; want %d", n, 123)
	}
	un := v.GetUint("foo")
	if un != 123 {
		t.Fatalf("unexpected value; got %d; want %d", un, 123)
	}
	n = v.GetInt("bar")
	if n != 0 {
		t.Fatalf("unexpected non-zero value; got %d", n)
	}
	f := v.GetFloat64("foo")
	if f != 123.0 {
		t.Fatalf("unexpected value; got %f; want %f", f, 123.0)
	}
	f = v.GetFloat64("bar")
	if f != 0 {
		t.Fatalf("unexpected value; got %f; want %f", f, 0.0)
	}
	f = v.GetFloat64("foooo", "bar")
	if f != 0 {
		t.Fatalf("unexpected value; got %f; want %f", f, 0.0)
	}
	f = v.GetFloat64()
	if f != 0 {
		t.Fatalf("unexpected value; got %f; want %f", f, 0.0)
	}
	sb := v.GetStringBytes("bar")
	if string(sb) != "433" {
		t.Fatalf("unexpected value; got %q; want %q", sb, "443")
	}
	sb = v.GetStringBytes("foo")
	if sb != nil {
		t.Fatalf("unexpected value; got %q; want %q", sb, []byte(nil))
	}
	bv := v.GetBool("baz")
	if !bv {
		t.Fatalf("unexpected value; got %v; want %v", bv, true)
	}
	bv = v.GetBool("bar")
	if bv {
		t.Fatalf("unexpected value; got %v; want %v", bv, false)
	}

	zv := v.Get("zero_float1")
	zf, err := zv.Float64()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if zf != 0 {
		t.Fatalf("unexpected zero_float1 value: %f. Expecting 0", zf)
	}

	zv = v.Get("zero_float2")
	zf, err = zv.Float64()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if zf != 0 {
		t.Fatalf("unexpected zero_float1 value: %f. Expecting 0", zf)
	}
}

func TestVisitNil(t *testing.T) {
	var p Parser
	v, err := p.Parse(`{}`)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	o := v.GetObject("non-existing-key")
	if o != nil {
		t.Fatalf("obtained an object for non-existing key: %#v", o)
	}
	o.Visit(func(k []byte, v *Value) {
		t.Fatalf("unexpected visit call; k=%q; v=%s", k, v)
	})
}

func TestValueGet(t *testing.T) {
	var pp ParserPool

	p := pp.Get()
	v, err := p.ParseBytes([]byte(`{"xx":33.33,"foo":[123,{"bar":["baz"],"x":"y"}], "": "empty-key", "empty-value": ""}`))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	t.Run("positive", func(t *testing.T) {
		sb := v.GetStringBytes("")
		if string(sb) != "empty-key" {
			t.Fatalf("unexpected value for empty key; got %q; want %q", sb, "empty-key")
		}
		sb = v.GetStringBytes("empty-value")
		if string(sb) != "" {
			t.Fatalf("unexpected non-empty value: %q", sb)
		}

		vv := v.Get("foo", "1")
		if vv == nil {
			t.Fatalf("cannot find the required value")
		}
		o, err := vv.Object()
		if err != nil {
			t.Fatalf("cannot obtain object: %s", err)
		}

		n := 0
		o.Visit(func(k []byte, v *Value) {
			n++
			switch string(k) {
			case "bar":
				if v.Type() != TypeArray {
					t.Fatalf("unexpected value type; got %d; want %d", v.Type(), TypeArray)
				}
				s := v.String()
				if s != `["baz"]` {
					t.Fatalf("unexpected array; got %q; want %q", s, `["baz"]`)
				}
			case "x":
				sb, err := v.StringBytes()
				if err != nil {
					t.Fatalf("cannot obtain string: %s", err)
				}
				if string(sb) != "y" {
					t.Fatalf("unexpected string; got %q; want %q", sb, "y")
				}
			default:
				t.Fatalf("unknown key: %s", k)
			}
		})
		if n != 2 {
			t.Fatalf("unexpected number of items visited in the array; got %d; want %d", n, 2)
		}
	})

	t.Run("negative", func(t *testing.T) {
		vv := v.Get("nonexisting", "path")
		if vv != nil {
			t.Fatalf("expecting nil value for nonexisting path. Got %#v", vv)
		}
		vv = v.Get("foo", "bar", "baz")
		if vv != nil {
			t.Fatalf("expecting nil value for nonexisting path. Got %#v", vv)
		}
		vv = v.Get("foo", "-123")
		if vv != nil {
			t.Fatalf("expecting nil value for nonexisting path. Got %#v", vv)
		}
		vv = v.Get("foo", "234")
		if vv != nil {
			t.Fatalf("expecting nil value for nonexisting path. Got %#v", vv)
		}
		vv = v.Get("xx", "yy")
		if vv != nil {
			t.Fatalf("expecting nil value for nonexisting path. Got %#v", vv)
		}
	})

	pp.Put(p)
}

func TestParserParse(t *testing.T) {
	var p Parser

	t.Run("complex-string", func(t *testing.T) {
		v, err := p.Parse(`{"тест":1, "\\\"фыва\"":2, "\\\"\u1234x":"\\fЗУ\\\\"}`)
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		n := v.GetInt("тест")
		if n != 1 {
			t.Fatalf("unexpected int; got %d; want %d", n, 1)
		}
		n = v.GetInt(`\"фыва"`)
		if n != 2 {
			t.Fatalf("unexpected int; got %d; want %d", n, 2)
		}
		sb := v.GetStringBytes("\\\"\u1234x")
		if string(sb) != `\fЗУ\\` {
			t.Fatalf("unexpected string; got %q; want %q", sb, `\fЗУ\\`)
		}
	})

	t.Run("invalid-string-escape", func(t *testing.T) {
		v, err := p.Parse(`"fo\u"`)
		if err != nil {
			t.Fatalf("unexpected error when parsing string")
		}
		// Make sure only valid string part remains
		sb, err := v.StringBytes()
		if err != nil {
			t.Fatalf("cannot obtain string: %s", err)
		}
		if string(sb) != "fo\\u" {
			t.Fatalf("unexpected string; got %q; want %q", sb, "fo\\u")
		}

		v, err = p.Parse(`"foo\ubarz2134"`)
		if err != nil {
			t.Fatalf("unexpected error when parsing string")
		}
		sb, err = v.StringBytes()
		if err != nil {
			t.Fatalf("cannot obtain string: %s", err)
		}
		if string(sb) != "foo\\ubarz2134" {
			t.Fatalf("unexpected string; got %q; want %q", sb, "foo")
		}
	})

	t.Run("invalid-number", func(t *testing.T) {
		v, err := p.Parse("123+456")
		if err != nil {
			t.Fatalf("unexpected error when parsing int")
		}

		// Make sure invalid int isn't parsed.
		n, err := v.Int()
		if err == nil {
			t.Fatalf("expecting non-nil error")
		}
		if n != 0 {
			t.Fatalf("unexpected int; got %d; want %d", n, 0)
		}
	})

	t.Run("empty-json", func(t *testing.T) {
		_, err := p.Parse("")
		if err == nil {
			t.Fatalf("expecting non-nil error when parsing empty json")
		}
		_, err = p.Parse("\n\t    \n")
		if err == nil {
			t.Fatalf("expecting non-nil error when parsing empty json")
		}
	})

	t.Run("invalid-tail", func(t *testing.T) {
		_, err := p.Parse("123 456")
		if err == nil {
			t.Fatalf("expecting non-nil error when parsing invalid tail")
		}
		_, err = p.Parse("[] 1223")
		if err == nil {
			t.Fatalf("expecting non-nil error when parsing invalid tail")
		}
	})

	t.Run("invalid-json", func(t *testing.T) {
		f := func(s string) {
			t.Helper()
			if _, err := p.Parse(s); err == nil {
				t.Fatalf("expecting non-nil error when parsing invalid json %q", s)
			}
		}

		f("free")
		f("tree")
		f("\x00\x10123")
		f("1 \n\x01")
		f("{\x00}")
		f("[\x00]")
		f("\"foo\"\x00")
		f("{\"foo\"\x00:123}")
		f("nil")
		f("[foo]")
		f("{foo}")
		f("[123 34]")
		f(`{"foo" "bar"}`)
		f(`{"foo":123 "bar":"baz"}`)
		f("-2134.453eec+43")

		if _, err := p.Parse("-2134.453E+43"); err != nil {
			t.Fatalf("unexpected error when parsing number: %s", err)
		}

		// Incomplete object key key.
		f(`{"foo: 123}`)

		// Incomplete string.
		f(`"{\"foo\": 123}`)

		v, err := p.Parse(`"{\"foo\": 123}"`)
		if err != nil {
			t.Fatalf("unexpected error when parsing json string: %s", err)
		}
		sb := v.GetStringBytes()
		if string(sb) != `{"foo": 123}` {
			t.Fatalf("unexpected string value; got %q; want %q", sb, `{"foo": 123}`)
		}
	})

	t.Run("incomplete-object", func(t *testing.T) {
		f := func(s string) {
			t.Helper()
			if _, err := p.Parse(s); err == nil {
				t.Fatalf("expecting non-nil error when parsing incomplete object %q", s)
			}
		}

		f(" {  ")
		f(`{"foo"`)
		f(`{"foo":`)
		f(`{"foo":null`)
		f(`{"foo":null,`)
		f(`{"foo":null,}`)
		f(`{"foo":null,"bar"}`)

		if _, err := p.Parse(`{"foo":null,"bar":"baz"}`); err != nil {
			t.Fatalf("unexpected error when parsing object: %s", err)
		}
	})

	t.Run("incomplete-array", func(t *testing.T) {
		f := func(s string) {
			t.Helper()
			if _, err := p.Parse(s); err == nil {
				t.Fatalf("expecting non-nil error when parsing incomplete array %q", s)
			}
		}

		f("  [ ")
		f("[123")
		f("[123,")
		f("[123,]")
		f("[123,{}")
		f("[123,{},]")

		if _, err := p.Parse("[123,{},[]]"); err != nil {
			t.Fatalf("unexpected error when parsing array: %s", err)
		}
	})

	t.Run("incomplete-string", func(t *testing.T) {
		f := func(s string) {
			t.Helper()
			if _, err := p.Parse(s); err == nil {
				t.Fatalf("expecting non-nil error when parsing incomplete string %q", s)
			}
		}

		f(`  "foo`)
		f(`"foo\`)
		f(`"foo\"`)
		f(`"foo\\\"`)
		f(`"foo'`)
		f(`"foo'bar'`)

		if _, err := p.Parse(`"foo\\\""`); err != nil {
			t.Fatalf("unexpected error when parsing string: %s", err)
		}
	})

	t.Run("empty-object", func(t *testing.T) {
		v, err := p.Parse("{}")
		if err != nil {
			t.Fatalf("cannot parse empty object: %s", err)
		}
		tp := v.Type()
		if tp != TypeObject || tp.String() != "object" {
			t.Fatalf("unexpected value obtained for empty object: %#v", v)
		}
		o, err := v.Object()
		if err != nil {
			t.Fatalf("cannot obtain object: %s", err)
		}
		n := o.Len()
		if n != 0 {
			t.Fatalf("unexpected number of items in empty object: %d; want 0", n)
		}
		s := v.String()
		if s != "{}" {
			t.Fatalf("unexpected string representation of empty object: got %q; want %q", s, "{}")
		}
	})

	t.Run("empty-array", func(t *testing.T) {
		v, err := p.Parse("[]")
		if err != nil {
			t.Fatalf("cannot parse empty array: %s", err)
		}
		tp := v.Type()
		if tp != TypeArray || tp.String() != "array" {
			t.Fatalf("unexpected value obtained for empty array: %#v", v)
		}
		a, err := v.Array()
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		n := len(a)
		if n != 0 {
			t.Fatalf("unexpected number of items in empty array: %d; want 0", n)
		}
		s := v.String()
		if s != "[]" {
			t.Fatalf("unexpected string representation of empty array: got %q; want %q", s, "[]")
		}
	})

	t.Run("null", func(t *testing.T) {
		v, err := p.Parse("null")
		if err != nil {
			t.Fatalf("cannot parse null: %s", err)
		}
		tp := v.Type()
		if tp != TypeNull || tp.String() != "null" {
			t.Fatalf("unexpected value obtained for null: %#v", v)
		}
		s := v.String()
		if s != "null" {
			t.Fatalf("unexpected string representation of null; got %q; want %q", s, "null")
		}
	})

	t.Run("true", func(t *testing.T) {
		v, err := p.Parse("true")
		if err != nil {
			t.Fatalf("cannot parse true: %s", err)
		}
		tp := v.Type()
		if tp != TypeTrue || tp.String() != "true" {
			t.Fatalf("unexpected value obtained for true: %#v", v)
		}
		b, err := v.Bool()
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if !b {
			t.Fatalf("expecting true; got false")
		}
		s := v.String()
		if s != "true" {
			t.Fatalf("unexpected string representation of true; got %q; want %q", s, "true")
		}
	})

	t.Run("false", func(t *testing.T) {
		v, err := p.Parse("false")
		if err != nil {
			t.Fatalf("cannot parse false: %s", err)
		}
		tp := v.Type()
		if tp != TypeFalse || tp.String() != "false" {
			t.Fatalf("unexpected value obtained for false: %#v", v)
		}
		b, err := v.Bool()
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if b {
			t.Fatalf("expecting false; got true")
		}
		s := v.String()
		if s != "false" {
			t.Fatalf("unexpected string representation of false; got %q; want %q", s, "false")
		}
	})

	t.Run("integer", func(t *testing.T) {
		v, err := p.Parse("12345")
		if err != nil {
			t.Fatalf("cannot parse integer: %s", err)
		}
		tp := v.Type()
		if tp != TypeNumber || tp.String() != "number" {
			t.Fatalf("unexpected type obtained for integer: %#v", v)
		}
		n, err := v.Int()
		if err != nil {
			t.Fatalf("cannot obtain int: %s", err)
		}
		if n != 12345 {
			t.Fatalf("unexpected value obtained for integer; got %d; want %d", n, 12345)
		}
		s := v.String()
		if s != "12345" {
			t.Fatalf("unexpected string representation of integer; got %q; want %q", s, "12345")
		}
	})

	t.Run("int64", func(t *testing.T) {
		v, err := p.Parse("-8838840643388017390")
		if err != nil {
			t.Fatalf("cannot parse int64: %s", err)
		}
		tp := v.Type()
		if tp != TypeNumber || tp.String() != "number" {
			t.Fatalf("unexpected type obtained for int64: %#v", v)
		}
		n, err := v.Int64()
		if err != nil {
			t.Fatalf("cannot obtain int64: %s", err)
		}
		if n != -8838840643388017390 {
			t.Fatalf("unexpected value obtained for int64; got %d; want %d", n, -8838840643388017390)
		}
		s := v.String()
		if s != "-8838840643388017390" {
			t.Fatalf("unexpected string representation of int64; got %q; want %q", s, "-8838840643388017390")
		}
	})

	t.Run("uint64", func(t *testing.T) {
		v, err := p.Parse("18446744073709551615")
		if err != nil {
			t.Fatalf("cannot parse uint64: %s", err)
		}
		tp := v.Type()
		if tp != TypeNumber || tp.String() != "number" {
			t.Fatalf("unexpected type obtained for uint64: %#v", v)
		}
		n, err := v.Uint64()
		if err != nil {
			t.Fatalf("cannot obtain uint64: %s", err)
		}
		if n != 18446744073709551615 {
			t.Fatalf("unexpected value obtained for uint64; got %d; want %d", n, uint64(18446744073709551615))
		}
		s := v.String()
		if s != "18446744073709551615" {
			t.Fatalf("unexpected string representation of uint64; got %q; want %q", s, "18446744073709551615")
		}
	})

	t.Run("float", func(t *testing.T) {
		v, err := p.Parse("-12.345")
		if err != nil {
			t.Fatalf("cannot parse integer: %s", err)
		}
		n, err := v.Float64()
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		tp := v.Type()
		if tp != TypeNumber || tp.String() != "number" {
			t.Fatalf("unexpected type obtained for integer: %#v", v)
		}
		if n != -12.345 {
			t.Fatalf("unexpected value obtained for integer; got %f; want %f", n, -12.345)
		}
		s := v.String()
		if s != "-12.345" {
			t.Fatalf("unexpected string representation of integer; got %q; want %q", s, "-12.345")
		}
	})

	t.Run("string", func(t *testing.T) {
		v, err := p.Parse(`"foo bar"`)
		if err != nil {
			t.Fatalf("cannot parse string: %s", err)
		}
		tp := v.Type()
		if tp != TypeString || tp.String() != "string" {
			t.Fatalf("unexpected type obtained for string: %#v", v)
		}
		sb, err := v.StringBytes()
		if err != nil {
			t.Fatalf("cannot obtain string: %s", err)
		}
		if string(sb) != "foo bar" {
			t.Fatalf("unexpected value obtained for string; got %q; want %q", sb, "foo bar")
		}
		ss := v.String()
		if ss != `"foo bar"` {
			t.Fatalf("unexpected string representation of string; got %q; want %q", ss, `"foo bar"`)
		}
	})

	t.Run("string-escaped", func(t *testing.T) {
		v, err := p.Parse(`"\n\t\\foo\"bar\u3423x\/\b\f\r\\"`)
		if err != nil {
			t.Fatalf("cannot parse string: %s", err)
		}
		tp := v.Type()
		if tp != TypeString {
			t.Fatalf("unexpected type obtained for string: %#v", v)
		}
		sb, err := v.StringBytes()
		if err != nil {
			t.Fatalf("cannot obtain string: %s", err)
		}
		if string(sb) != "\n\t\\foo\"bar\u3423x/\b\f\r\\" {
			t.Fatalf("unexpected value obtained for string; got %q; want %q", sb, "\n\t\\foo\"bar\u3423x/\b\f\r\\")
		}
		ss := v.String()
		if ss != `"\n\t\\foo\"bar㐣x/\b\f\r\\"` {
			t.Fatalf("unexpected string representation of string; got %q; want %q", ss, `"\n\t\\foo\"bar㐣x/\b\f\r\\"`)
		}
	})

	t.Run("object-one-element", func(t *testing.T) {
		v, err := p.Parse(`  {
	"foo"   : "bar"  }	 `)
		if err != nil {
			t.Fatalf("cannot parse object: %s", err)
		}
		tp := v.Type()
		if tp != TypeObject {
			t.Fatalf("unexpected type obtained for object: %#v", v)
		}
		o, err := v.Object()
		if err != nil {
			t.Fatalf("cannot obtain object: %s", err)
		}
		vv := o.Get("foo")
		if vv.Type() != TypeString {
			t.Fatalf("unexpected type for foo item: got %d; want %d", vv.Type(), TypeString)
		}
		vv = o.Get("non-existing key")
		if vv != nil {
			t.Fatalf("unexpected value obtained for non-existing key: %#v", vv)
		}

		s := v.String()
		if s != `{"foo":"bar"}` {
			t.Fatalf("unexpected string representation for object; got %q; want %q", s, `{"foo":"bar"}`)
		}
	})

	t.Run("object-multi-elements", func(t *testing.T) {
		v, err := p.Parse(`{"foo": [1,2,3  ]  ,"bar":{},"baz":123.456}`)
		if err != nil {
			t.Fatalf("cannot parse object: %s", err)
		}
		tp := v.Type()
		if tp != TypeObject {
			t.Fatalf("unexpected type obtained for object: %#v", v)
		}
		o, err := v.Object()
		if err != nil {
			t.Fatalf("cannot obtain object: %s", err)
		}
		vv := o.Get("foo")
		if vv.Type() != TypeArray {
			t.Fatalf("unexpected type for foo item; got %d; want %d", vv.Type(), TypeArray)
		}
		vv = o.Get("bar")
		if vv.Type() != TypeObject {
			t.Fatalf("unexpected type for bar item; got %d; want %d", vv.Type(), TypeObject)
		}
		vv = o.Get("baz")
		if vv.Type() != TypeNumber {
			t.Fatalf("unexpected type for baz item; got %d; want %d", vv.Type(), TypeNumber)
		}
		vv = o.Get("non-existing-key")
		if vv != nil {
			t.Fatalf("unexpected value obtained for non-existing key: %#v", vv)
		}

		s := v.String()
		if s != `{"foo":[1,2,3],"bar":{},"baz":123.456}` {
			t.Fatalf("unexpected string representation for object; got %q; want %q", s, `{"foo":[1,2,3],"bar":{},"baz":123.456}`)
		}
	})

	t.Run("array-one-element", func(t *testing.T) {
		v, err := p.Parse(`   [{"bar":[  [],[[]]   ]} ]  `)
		if err != nil {
			t.Fatalf("cannot parse array: %s", err)
		}
		tp := v.Type()
		if tp != TypeArray {
			t.Fatalf("unexpected type obtained for array: %#v", v)
		}
		a, err := v.Array()
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if len(a) != 1 {
			t.Fatalf("unexpected array len; got %d; want %d", len(a), 1)
		}
		if a[0].Type() != TypeObject {
			t.Fatalf("unexpected type for a[0]; got %d; want %d", a[0].Type(), TypeObject)
		}

		s := v.String()
		if s != `[{"bar":[[],[[]]]}]` {
			t.Fatalf("unexpected string representation for array; got %q; want %q", s, `[{"bar":[[],[[]]]}]`)
		}
	})

	t.Run("array-multi-elements", func(t *testing.T) {
		v, err := p.Parse(`   [1,"foo",{"bar":[     ],"baz":""}    ,[  "x" ,	"y"   ]     ]   `)
		if err != nil {
			t.Fatalf("cannot parse array: %s", err)
		}
		tp := v.Type()
		if tp != TypeArray {
			t.Fatalf("unexpected type obtained for array: %#v", v)
		}
		a, err := v.Array()
		if err != nil {
			t.Fatalf("unexpected error: %s", err)
		}
		if len(a) != 4 {
			t.Fatalf("unexpected array len; got %d; want %d", len(a), 4)
		}
		if a[0].Type() != TypeNumber {
			t.Fatalf("unexpected type for a[0]; got %d; want %d", a[0].Type(), TypeNumber)
		}
		if a[1].Type() != TypeString {
			t.Fatalf("unexpected type for a[1]; got %d; want %d", a[1].Type(), TypeString)
		}
		if a[2].Type() != TypeObject {
			t.Fatalf("unexpected type for a[2]; got %d; want %d", a[2].Type(), TypeObject)
		}
		if a[3].Type() != TypeArray {
			t.Fatalf("unexpected type for a[3]; got %d; want %d", a[3].Type(), TypeArray)
		}

		s := v.String()
		if s != `[1,"foo",{"bar":[],"baz":""},["x","y"]]` {
			t.Fatalf("unexpected string representation for array; got %q; want %q", s, `[1,"foo",{"bar":[],"baz":""},["x","y"]]`)
		}
	})

	t.Run("complex-object", func(t *testing.T) {
		s := `{"foo":[-1.345678,[[[[[]]]],{}],"bar"],"baz":{"bbb":123}}`
		v, err := p.Parse(s)
		if err != nil {
			t.Fatalf("cannot parse complex object: %s", err)
		}
		if v.Type() != TypeObject {
			t.Fatalf("unexpected type obtained for object: %#v", v)
		}

		ss := v.String()
		if ss != s {
			t.Fatalf("unexpected string representation for object; got %q; want %q", ss, s)
		}

		s = strings.TrimSpace(largeFixture)
		v, err = p.Parse(s)
		if err != nil {
			t.Fatalf("cannot parse largeFixture: %s", err)
		}
		ss = v.String()
		if ss != s {
			t.Fatalf("unexpected string representation for object; got\n%q; want\n%q", ss, s)
		}
	})

	t.Run("complex-object-visit-all", func(t *testing.T) {
		n := 0
		var f func(k []byte, v *Value)
		f = func(k []byte, v *Value) {
			switch v.Type() {
			case TypeObject:
				o, err := v.Object()
				if err != nil {
					t.Fatalf("cannot obtain object: %s", err)
				}
				o.Visit(f)
			case TypeArray:
				a, err := v.Array()
				if err != nil {
					t.Fatalf("unexpected error: %s", err)
				}
				for _, vv := range a {
					f(nil, vv)
				}
			case TypeString:
				sb, err := v.StringBytes()
				if err != nil {
					t.Fatalf("cannot obtain string: %s", err)
				}
				n += len(sb)
			case TypeNumber:
				nn, err := v.Int()
				if err != nil {
					t.Fatalf("cannot obtain int: %s", err)
				}
				n += nn
			}
		}

		s := strings.TrimSpace(largeFixture)
		v, err := p.Parse(s)
		if err != nil {
			t.Fatalf("cannot parse largeFixture: %s", err)
		}
		o, err := v.Object()
		if err != nil {
			t.Fatalf("cannot obtain object: %s", err)
		}
		o.Visit(f)

		if n != 21473 {
			t.Fatalf("unexpected n; got %d; want %d", n, 21473)
		}

		// Make sure the json remains valid after visiting all the items.
		ss := v.String()
		if ss != s {
			t.Fatalf("unexpected string representation for object; got\n%q; want\n%q", ss, s)
		}

	})
}

func TestParseBigObject(t *testing.T) {
	const itemsCount = 10000

	// build big json object
	var ss []string
	for i := 0; i < itemsCount; i++ {
		s := fmt.Sprintf(`"key_%d": "value_%d"`, i, i)
		ss = append(ss, s)
	}
	s := "{" + strings.Join(ss, ",") + "}"

	// parse it
	var p Parser
	v, err := p.Parse(s)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// Look up object items
	for i := 0; i < itemsCount; i++ {
		k := fmt.Sprintf("key_%d", i)
		expectedV := fmt.Sprintf("value_%d", i)
		sb := v.GetStringBytes(k)
		if string(sb) != expectedV {
			t.Fatalf("unexpected value obtained; got %q; want %q", sb, expectedV)
		}
	}

	// verify non-existing key returns nil
	sb := v.GetStringBytes("non-existing-key")
	if sb != nil {
		t.Fatalf("unexpected non-nil value for non-existing-key: %q", sb)
	}
}

func TestParseGetConcurrent(t *testing.T) {
	concurrency := 10
	ch := make(chan error, concurrency)
	s := `{"foo": "bar", "empty_obj": {}}`
	for i := 0; i < concurrency; i++ {
		go func() {
			ch <- testParseGetSerial(s)
		}()
	}
	for i := 0; i < concurrency; i++ {
		select {
		case err := <-ch:
			if err != nil {
				t.Fatalf("unexpected error during concurrent test: %s", err)
			}
		case <-time.After(time.Second):
			t.Fatalf("timeout")
		}
	}
}

func testParseGetSerial(s string) error {
	var p Parser
	for i := 0; i < 100; i++ {
		v, err := p.Parse(s)
		if err != nil {
			return fmt.Errorf("cannot parse %q: %s", s, err)
		}
		sb := v.GetStringBytes("foo")
		if string(sb) != "bar" {
			return fmt.Errorf("unexpected value for key=%q; got %q; want %q", "foo", sb, "bar")
		}
		vv := v.Get("empty_obj", "non-existing-key")
		if vv != nil {
			return fmt.Errorf("unexpected non-nil value got: %s", vv)
		}
	}
	return nil
}
