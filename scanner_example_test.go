package fastjson_test

import (
	"fmt"
	"github.com/valyala/fastjson"
	"log"
)

func ExampleScanner() {
	var sc fastjson.Scanner

	sc.Init(`   {"foo":  "bar"  }[  ]
		12345"xyz" true false null    `)

	for sc.Next() {
		fmt.Printf("%s\n", sc.Value())
	}
	if err := sc.Error(); err != nil {
		log.Fatalf("unexpected error: %s", err)
	}

	// Output:
	// {"foo":"bar"}
	// []
	// 12345
	// "xyz"
	// true
	// false
	// null
}

func ExampleScanner_reuse() {
	var sc fastjson.Scanner

	// The sc may be re-used in order to reduce the number
	// of memory allocations.
	for i := 0; i < 3; i++ {
		s := fmt.Sprintf(`[%d] "%d"`, i, i)
		sc.Init(s)
		for sc.Next() {
			fmt.Printf("%s,", sc.Value())
		}
		if err := sc.Error(); err != nil {
			log.Fatalf("unexpected error: %s", err)
		}
		fmt.Printf("\n")
	}

	// Output:
	// [0],"0",
	// [1],"1",
	// [2],"2",
}
