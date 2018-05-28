package fastjson_test

import (
	"fmt"
	"github.com/valyala/fastjson"
	"log"
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
