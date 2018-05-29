package fastjson

import (
	"strings"
	"testing"
)

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

func TestValueGetInt(t *testing.T) {
	var p Parser

	v, err := p.Parse(`{"foo": 123, "bar": "433", "baz": true}`)
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	n := v.GetInt("foo")
	if n != 123 {
		t.Fatalf("unexpected value; got %d; want %d", n, 123)
	}
	n = v.GetInt("bar")
	if n != 0 {
		t.Fatalf("unexpected non-zero value; got %d", n)
	}
	f := v.GetFloat64("foo")
	if f != 123.0 {
		t.Fatalf("unexpected value; got %f; want %f", f, 123.0)
	}
	sb := v.GetStringBytes("bar")
	if string(sb) != "433" {
		t.Fatalf("unexpected value; got %q; want %q", sb, "443")
	}
	bv := v.GetBool("baz")
	if !bv {
		t.Fatalf("unexpected value; got %v; want %v", bv, true)
	}
	bv = v.GetBool("bar")
	if bv {
		t.Fatalf("unexpected value; got %v; want %v", bv, false)
	}
}

func TestValueGet(t *testing.T) {
	var pp ParserPool

	p := pp.Get()
	v, err := p.ParseBytes([]byte(`{"xx":33.33,"foo":[123,{"bar":["baz"],"x":"y"}]}`))
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	t.Run("positive", func(t *testing.T) {
		vv := v.Get("foo", "1")
		if vv == nil {
			t.Fatalf("cannot find the required value")
		}
		o := vv.Object()

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
				if v.Type() != TypeString {
					t.Fatalf("unexpected value type; got %d; want %d", v.Type(), TypeString)
				}
				sb := v.StringBytes()
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
		_, err := p.Parse("foobar")
		if err == nil {
			t.Fatalf("expecting non-nil error when parsing invalid json")
		}
		_, err = p.Parse("tree")
		if err == nil {
			t.Fatalf("expecting non-nil error when parsing invalid json")
		}
		_, err = p.Parse("nil")
		if err == nil {
			t.Fatalf("expecting non-nil error when parsing invalid json")
		}
		_, err = p.Parse("[foo]")
		if err == nil {
			t.Fatalf("expecting non-nil error when parsing invalid json")
		}
		_, err = p.Parse("{foo}")
		if err == nil {
			t.Fatalf("expecting non-nil error when parsing invalid json")
		}
		_, err = p.Parse("[123 34]")
		if err == nil {
			t.Fatalf("expecting non-nil error when parsing invalid json")
		}
		_, err = p.Parse(`{"foo" "bar"}`)
		if err == nil {
			t.Fatalf("expecting non-nil error when parsing invalid json")
		}
		_, err = p.Parse(`{"foo":123 "bar":"baz"}`)
		if err == nil {
			t.Fatalf("expecting non-nil error when parsing invalid json")
		}
		_, err = p.Parse("-2134.453eec+43")
		if err == nil {
			t.Fatalf("expecting non-nil error when parsing invalid json")
		}
		_, err = p.Parse("-2134.453E+43")
		if err != nil {
			t.Fatalf("unexpected error when paring number: %s", err)
		}
	})

	t.Run("incomplete-object", func(t *testing.T) {
		_, err := p.Parse(" {  ")
		if err == nil {
			t.Fatalf("expecting non-nil error when parsing incomplete object")
		}
		_, err = p.Parse(`{"foo"`)
		if err == nil {
			t.Fatalf("expecting non-nil error when parsing incomplete object")
		}
		_, err = p.Parse(`{"foo":`)
		if err == nil {
			t.Fatalf("expecting non-nil error when parsing incomplete object")
		}
		_, err = p.Parse(`{"foo":null`)
		if err == nil {
			t.Fatalf("expecting non-nil error when parsing incomplete object")
		}
		_, err = p.Parse(`{"foo":null,`)
		if err == nil {
			t.Fatalf("expecting non-nil error when parsing incomplete object")
		}
		_, err = p.Parse(`{"foo":null,}`)
		if err == nil {
			t.Fatalf("expecting non-nil error when parsing incomplete object")
		}
		_, err = p.Parse(`{"foo":null,"bar"}`)
		if err == nil {
			t.Fatalf("expecting non-nil error when parsing incomplete object")
		}
		_, err = p.Parse(`{"foo":null,"bar":"baz"}`)
		if err != nil {
			t.Fatalf("unexpected error when parsing object: %s", err)
		}
	})

	t.Run("incomplete-array", func(t *testing.T) {
		_, err := p.Parse("  [ ")
		if err == nil {
			t.Fatalf("expecting non-nil error when parsing incomplete array")
		}
		_, err = p.Parse("[123")
		if err == nil {
			t.Fatalf("expecting non-nil error when parsing incomplete array")
		}
		_, err = p.Parse("[123,")
		if err == nil {
			t.Fatalf("expecting non-nil error when parsing incomplete array")
		}
		_, err = p.Parse("[123,]")
		if err == nil {
			t.Fatalf("expecting non-nil error when parsing incomplete array")
		}
		_, err = p.Parse("[123,{}")
		if err == nil {
			t.Fatalf("expecting non-nil error when parsing incomplete array")
		}
		_, err = p.Parse("[123,{},]")
		if err == nil {
			t.Fatalf("expecting non-nil error when parsing incomplete array")
		}
		_, err = p.Parse("[123,{},[]]")
		if err != nil {
			t.Fatalf("unexpected error when parsing array: %s", err)
		}
	})

	t.Run("incomplete-string", func(t *testing.T) {
		_, err := p.Parse(`  "foo`)
		if err == nil {
			t.Fatalf("expecting non-nil error when parsing incomplete string")
		}
		_, err = p.Parse(`"foo\`)
		if err == nil {
			t.Fatalf("expecting non-nil error when parsing incomplete string")
		}
		_, err = p.Parse(`"foo\"`)
		if err == nil {
			t.Fatalf("expecting non-nil error when parsing incomplete string")
		}
		_, err = p.Parse(`"foo\\\""`)
		if err != nil {
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
		n := v.Object().Len()
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
		n := len(v.Array())
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
		n := v.Int()
		if n != 12345 {
			t.Fatalf("unexpected value obtained for integer; got %d; want %d", n, 12345)
		}
		s := v.String()
		if s != "12345" {
			t.Fatalf("unexpected string representation of integer; got %q; want %q", s, "12345")
		}
	})

	t.Run("float", func(t *testing.T) {
		v, err := p.Parse("-12.345")
		if err != nil {
			t.Fatalf("cannot parse integer: %s", err)
		}
		tp := v.Type()
		if tp != TypeNumber || tp.String() != "number" {
			t.Fatalf("unexpected type obtained for integer: %#v", v)
		}
		n := v.Float64()
		if n != -12.345 {
			t.Fatalf("unexpected value obtained for integer; got %f; want %f", n, -12.345)
		}
		s := v.String()
		if s != "-12.345000" {
			t.Fatalf("unexpected string representation of integer; got %q; want %q", s, "-12.345000")
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
		sb := v.StringBytes()
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
		sb := v.StringBytes()
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
		o := v.Object()
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
		o := v.Object()
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
		if s != `{"foo":[1,2,3],"bar":{},"baz":123.456000}` {
			t.Fatalf("unexpected string representation for object; got %q; want %q", s, `{"foo":[1,2,3],"bar":{},"baz":123.456000}`)
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
		a := v.Array()
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
		a := v.Array()
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
				v.Object().Visit(f)
			case TypeArray:
				for _, vv := range v.Array() {
					f(nil, vv)
				}
			case TypeString:
				n += len(v.StringBytes())
			case TypeNumber:
				n += v.Int()
			}
		}

		s := strings.TrimSpace(largeFixture)
		v, err := p.Parse(s)
		if err != nil {
			t.Fatalf("cannot parse largeFixture: %s", err)
		}
		v.Object().Visit(f)

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
