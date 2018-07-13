package fastjson_test

import (
	"fmt"
	"github.com/valyala/fastjson"
	"log"
	"strconv"
)

func ExampleParser_Parse() {
	var p fastjson.Parser
	v, err := p.Parse(`{"foo":"bar", "baz": 123}`)
	if err != nil {
		log.Fatalf("cannot parse json: %s", err)
	}

	fmt.Printf("foo=%s, baz=%d", v.GetStringBytes("foo"), v.GetInt("baz"))

	// Output:
	// foo=bar, baz=123
}

func ExampleParser_Parse_reuse() {
	var p fastjson.Parser

	// p may be re-used for parsing multiple json strings.
	// This improves parsing speed by reducing the number
	// of memory allocations.
	//
	// Parse call invalidates all the objects previously obtained from p,
	// so don't hold these objects after parsing the next json.

	for i := 0; i < 3; i++ {
		s := fmt.Sprintf(`["foo_%d","bar_%d","%d"]`, i, i, i)
		v, err := p.Parse(s)
		if err != nil {
			log.Fatalf("cannot parse json: %s", err)
		}
		key := strconv.Itoa(i)
		fmt.Printf("a[%d]=%s\n", i, v.GetStringBytes(key))
	}

	// Output:
	// a[0]=foo_0
	// a[1]=bar_1
	// a[2]=2
}

func ExampleValue_MarshalTo() {
	s := `{
		"name": "John",
		"items": [
			{
				"key": "foo",
				"value": 123.456,
				"arr": [1, "foo"]
			},
			{
				"key": "bar",
				"field": [3, 4, 5]
			}
		]
	}`
	var p fastjson.Parser
	v, err := p.Parse(s)
	if err != nil {
		log.Fatalf("cannot parse json: %s", err)
	}

	// Marshal items.0 into newly allocated buffer.
	buf := v.Get("items", "0").MarshalTo(nil)
	fmt.Printf("items.0 = %s\n", buf)

	// Re-use buf for marshaling items.1.
	buf = v.Get("items", "1").MarshalTo(buf[:0])
	fmt.Printf("items.1 = %s\n", buf)

	// Output:
	// items.0 = {"key":"foo","value":123.456,"arr":[1,"foo"]}
	// items.1 = {"key":"bar","field":[3,4,5]}
}

func ExampleValue_Get() {
	s := `{"foo":[{"bar":{"baz":123,"x":"434"},"y":[]},[null, false]],"qwe":true}`
	var p fastjson.Parser
	v, err := p.Parse(s)
	if err != nil {
		log.Fatalf("cannot parse json: %s", err)
	}

	vv := v.Get("foo", "0", "bar", "x")
	fmt.Printf("foo[0].bar.x=%s\n", vv.GetStringBytes())

	vv = v.Get("qwe")
	fmt.Printf("qwe=%v\n", vv.GetBool())

	vv = v.Get("foo", "1")
	fmt.Printf("foo[1]=%s\n", vv)

	vv = v.Get("foo").Get("1").Get("1")
	fmt.Printf("foo[1][1]=%s\n", vv)

	// non-existing key
	vv = v.Get("foo").Get("bar").Get("baz", "1234")
	fmt.Printf("foo.bar.baz[1234]=%v\n", vv)

	// Output:
	// foo[0].bar.x=434
	// qwe=true
	// foo[1]=[null,false]
	// foo[1][1]=false
	// foo.bar.baz[1234]=<nil>
}

func ExampleValue_Type() {
	s := `{
		"object": {},
		"array": [],
		"string": "foobar",
		"number": 123.456,
		"true": true,
		"false": false,
		"null": null
	}`

	var p fastjson.Parser
	v, err := p.Parse(s)
	if err != nil {
		log.Fatalf("cannot parse json: %s", err)
	}

	fmt.Printf("%s\n", v.Get("object").Type())
	fmt.Printf("%s\n", v.Get("array").Type())
	fmt.Printf("%s\n", v.Get("string").Type())
	fmt.Printf("%s\n", v.Get("number").Type())
	fmt.Printf("%s\n", v.Get("true").Type())
	fmt.Printf("%s\n", v.Get("false").Type())
	fmt.Printf("%s\n", v.Get("null").Type())

	// Output:
	// object
	// array
	// string
	// number
	// true
	// false
	// null
}

func ExampleObject_Visit() {
	s := `{
		"obj": { "foo": 1234 },
		"arr": [ 23,4, "bar" ],
		"str": "foobar"
	}`

	var p fastjson.Parser
	v, err := p.Parse(s)
	if err != nil {
		log.Fatalf("cannot parse json: %s", err)
	}
	o, err := v.Object()
	if err != nil {
		log.Fatalf("cannot obtain object from json value: %s", err)
	}

	o.Visit(func(k []byte, v *fastjson.Value) {
		switch string(k) {
		case "obj":
			fmt.Printf("object %s\n", v)
		case "arr":
			fmt.Printf("array %s\n", v)
		case "str":
			fmt.Printf("string %s\n", v)
		}
	})

	// Output:
	// object {"foo":1234}
	// array [23,4,"bar"]
	// string "foobar"
}

func ExampleValue_GetStringBytes() {
	s := `[
		{"foo": "bar"},
		[123, "baz"]
	]`

	var p fastjson.Parser
	v, err := p.Parse(s)
	if err != nil {
		log.Fatalf("cannot parse json: %s", err)
	}
	fmt.Printf("v[0].foo = %q\n", v.GetStringBytes("0", "foo"))
	fmt.Printf("v[1][1] = %q\n", v.GetStringBytes("1", "1"))
	fmt.Printf("v[1][0] = %q\n", v.GetStringBytes("1", "0"))
	fmt.Printf("v.foo.bar.baz = %q\n", v.GetStringBytes("foo", "bar", "baz"))

	// Output:
	// v[0].foo = "bar"
	// v[1][1] = "baz"
	// v[1][0] = ""
	// v.foo.bar.baz = ""
}
