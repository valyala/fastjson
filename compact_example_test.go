package fastjson_test

import (
	"fmt"
	"log"

	"github.com/valyala/fastjson"
)

func ExampleCompact() {
	src := []byte(`{
		"foo": [
			123,
			"bar"
		]
	}`)

	compacted, err := fastjson.Compact(nil, src)
	if err != nil {
		log.Fatalf("cannot compact JSON: %s", err)
	}
	fmt.Printf("%s", compacted)

	// Output:
	// {"foo":[123,"bar"]}
}
