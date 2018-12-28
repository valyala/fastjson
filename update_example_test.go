package fastjson_test

import (
	"fmt"
	"log"

	"github.com/valyala/fastjson"
)

func ExampleObject_Del() {
	v := fastjson.MustParse(`{"foo": 123, "bar": [1,2], "baz": "xyz"}`)
	o, err := v.Object()
	if err != nil {
		log.Fatalf("cannot otain object: %s", err)
	}
	fmt.Printf("%s\n", o)

	o.Del("bar")
	fmt.Printf("%s\n", o)

	o.Del("foo")
	fmt.Printf("%s\n", o)

	o.Del("baz")
	fmt.Printf("%s\n", o)

	// Output:
	// {"foo":123,"bar":[1,2],"baz":"xyz"}
	// {"foo":123,"baz":"xyz"}
	// {"baz":"xyz"}
	// {}
}

func ExampleValue_Del() {
	v := fastjson.MustParse(`{"foo": 123, "bar": [1,2], "baz": "xyz"}`)
	fmt.Printf("%s\n", v)

	v.Del("foo")
	fmt.Printf("%s\n", v)

	v.Get("bar").Del("0")
	fmt.Printf("%s\n", v)

	// Output:
	// {"foo":123,"bar":[1,2],"baz":"xyz"}
	// {"bar":[1,2],"baz":"xyz"}
	// {"bar":[2],"baz":"xyz"}
}

func ExampleValue_Set() {
	v := fastjson.MustParse(`{"foo":1,"bar":[2,3]}`)

	// Replace `foo` value with "xyz"
	v.Set("foo", fastjson.MustParse(`"xyz"`))
	// Add "newv":123
	v.Set("newv", fastjson.MustParse(`123`))
	fmt.Printf("%s\n", v)

	// Replace `bar.1` with {"x":"y"}
	v.Get("bar").Set("1", fastjson.MustParse(`{"x":"y"}`))
	// Add `bar.3="qwe"
	v.Get("bar").Set("3", fastjson.MustParse(`"qwe"`))
	fmt.Printf("%s\n", v)

	// Output:
	// {"foo":"xyz","bar":[2,3],"newv":123}
	// {"foo":"xyz","bar":[2,{"x":"y"},null,"qwe"],"newv":123}
}
