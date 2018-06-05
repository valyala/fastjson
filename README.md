[![Build Status](https://travis-ci.org/valyala/fastjson.svg)](https://travis-ci.org/valyala/fastjson)
[![GoDoc](https://godoc.org/github.com/valyala/fastjson?status.svg)](http://godoc.org/github.com/valyala/fastjson)
[![Go Report](https://goreportcard.com/badge/github.com/valyala/fastjson)](https://goreportcard.com/report/github.com/valyala/fastjson)
[![codecov](https://codecov.io/gh/valyala/fastjson/branch/master/graph/badge.svg)](https://codecov.io/gh/valyala/fastjson)

# fastjson - fast JSON parser for Go


## Features

  * Fast. As usual, up to 15x faster than the standard [encoding/json](https://golang.org/pkg/encoding/json/).
    See [benchmarks](#benchmarks).
  * Parses arbitrary JSON without schema, reflection, struct magic and code generation
    contrary to [easyjson](https://github.com/mailru/easyjson).
  * Provides simple [API](http://godoc.org/github.com/valyala/fastjson).
  * Outperforms [jsonparser](https://github.com/buger/jsonparser) and [gjson](https://github.com/tidwall/gjson)
    when accessing multiple unrelated fields, since `fastjson` parses the input JSON only once.
  * Validates the parsed JSON unlike [gjson](https://github.com/tidwall/gjson).
  * May parse array containing values with distinct types (aka non-homogenous types).
    For instance, `fastjson` easily parses the following JSON array `[123, "foo", [456], {"k": "v"}, null]`.


## Known limitations

  * Requies extra care to work with - references to certain objects recursively
    returned by [Parser](https://godoc.org/github.com/valyala/fastjson#Parser)
    must be released before the next call to [Parse](https://godoc.org/github.com/valyala/fastjson#Parser.Parse).
    Otherwise the program may work improperly and/or may crash.
    Adhere recommendations from [docs](https://godoc.org/github.com/valyala/fastjson).
  * Cannot parse JSON from `io.Reader`. There is [Scanner](https://godoc.org/github.com/valyala/fastjson#Scanner)
    for parsing stream of JSON values from a string.


## Usage

One-liner accessing a single field:
```go
	s := []byte(`{"foo": [123, "bar"]}`)
	fmt.Printf("foo.0=%s\n", fastjson.GetInt(s, "foo", "0"))

	// Output:
	// foo.1=123
```

Accessing multiple fields with error handling:
```go
        var p fastjson.Parser
        v, err := p.Parse(`{
                "str": "bar",
                "int": 123,
                "float": 1.23,
                "bool": true,
                "arr": [1, "foo", {}]
        }`)
        if err != nil {
                log.Fatal(err)
        }
        fmt.Printf("foo=%s\n", v.GetStringBytes("str"))
        fmt.Printf("int=%d\n", v.GetInt("int"))
        fmt.Printf("float=%f\n", v.GetFloat64("float"))
        fmt.Printf("bool=%v\n", v.GetBool("bool"))
        fmt.Printf("arr.1=%s\n", v.GetStringBytes("arr", "1"))

        // Output:
        // foo=bar
        // int=123
        // float=1.230000
        // bool=true
        // arr.1=foo
```

See also [examples](https://godoc.org/github.com/valyala/fastjson#pkg-examples).


## Security

  * `fastjson` shouldn't crash or panic when parsing input strings specially crafted
    by an attacker. It must return error on invalid input JSON.
  * `fastjson` requires up to `sizeof(Value) * len(inputJSON)` bytes of memory
    for parsing `inputJSON` string. Limit the maximum size of the `inputJSON`
    before parsing it in order to limit the maximum memory usage.


## Benchmarks

Legend:

  * `small` - small-length parsing. JSON length is 190 bytes.
  * `medium` - medium-length parsing. JSON length is 2.4KB.
  * `large` - large-length parsing. JSON length is 24KB.
  * `stdjson-map` - parsing into a `map[string]interface{}` using `encoding/json`.
  * `stdjson-struct` - parsing into a struct containing
    a subset of fields of the parsed JSON, using `encoding/json`.
  * `stdjson-empty-struct` - parsing into an empty struct using `encoding/json`.
    This is the fastests possible solution for `encoding/json`, may be used
    for json validation.
  * `fastjson` - parsing using `fastjson` without fields access.
  * `fastjson-get` - parsing using `fastjson` with fields access similar to `stdjson-struct`.

```
$ GOMAXPROCS=1 go test -bench=Parse$
goos: linux
goarch: amd64
pkg: github.com/valyala/fastjson
BenchmarkParse/small/stdjson-map         	  200000	      6631 ns/op	  28.50 MB/s	     960 B/op	      51 allocs/op
BenchmarkParse/small/stdjson-struct      	  500000	      3141 ns/op	  60.17 MB/s	     224 B/op	       4 allocs/op
BenchmarkParse/small/stdjson-empty-struct         	  500000	      2360 ns/op	  80.07 MB/s	     168 B/op	       2 allocs/op
BenchmarkParse/small/fastjson                     	 3000000	       457 ns/op	 413.04 MB/s	       0 B/op	       0 allocs/op
BenchmarkParse/small/fastjson-get                 	 2000000	       660 ns/op	 286.15 MB/s	       0 B/op	       0 allocs/op
BenchmarkParse/medium/stdjson-map                 	   50000	     38041 ns/op	  61.20 MB/s	   10195 B/op	     208 allocs/op
BenchmarkParse/medium/stdjson-struct              	   30000	     45253 ns/op	  51.44 MB/s	    9173 B/op	     258 allocs/op
BenchmarkParse/medium/stdjson-empty-struct        	  100000	     19950 ns/op	 116.69 MB/s	     280 B/op	       5 allocs/op
BenchmarkParse/medium/fastjson                    	  300000	      4063 ns/op	 572.97 MB/s	       0 B/op	       0 allocs/op
BenchmarkParse/medium/fastjson-get                	  300000	      4174 ns/op	 557.70 MB/s	       0 B/op	       0 allocs/op
BenchmarkParse/large/stdjson-map                  	    2000	    592670 ns/op	  47.45 MB/s	  210752 B/op	    2785 allocs/op
BenchmarkParse/large/stdjson-struct               	    5000	    272348 ns/op	 103.26 MB/s	   15616 B/op	     353 allocs/op
BenchmarkParse/large/stdjson-empty-struct         	    5000	    238023 ns/op	 118.15 MB/s	     280 B/op	       5 allocs/op
BenchmarkParse/large/fastjson                     	   30000	     52874 ns/op	 531.89 MB/s	       9 B/op	       0 allocs/op
BenchmarkParse/large/fastjson-get                 	   30000	     53173 ns/op	 528.89 MB/s	       9 B/op	       0 allocs/op
PASS
ok  	github.com/valyala/fastjson	25.181s
```

As you can see, `fastsjon` outperforms `encoding/json`:

  * by up to a factor of 15x for `small`-length parsing;
  * by up to a factor of 11x for `medium`-length and `large`-length parsing.


## FAQ

  * Q: _There are a ton of other high-perf packages for JSON parsing in Go. Why creating yet another package?_
    A: Because other packages require either rigid JSON schema via struct magic
       and code generation or perform poorly when multiple unrelated fields
       must be obtained from the parsed JSON.
       Additionally, `fastjson` provides nicer [API](http://godoc.org/github.com/valyala/fastjson).

  * Q: _What is the main purpose for `fastjson`?_
    A: High-perf JSON parsing for [RTB](https://www.iab.com/wp-content/uploads/2015/05/OpenRTB_API_Specification_Version_2_3_1.pdf)
       and other [JSON-RPC](https://en.wikipedia.org/wiki/JSON-RPC) services.
       Use [gjson](https://github.com/tidwall/gjson) if you need fetching only a few fields from the JSON.

  * Q: _Why fastjson doesn't provide fast marshaling (serialization)?_
    A: Because other solutions exist. I'd recommend [quicktemplate](https://github.com/valyala/quicktemplate#use-cases)
       for high-performance JSON marshaling :)

  * Q: _`fastjson` crashes my program!_
    A: There is high probability of improper use.
       * Make sure you don't hold references to objects recursively returned by `Parser` / `Scanner`
         beyond the next `Parser.Parse` / `Scanner.Next` call
         if such restriction is mentioned in [docs](https://github.com/valyala/fastjson/issues/new).
       * Make sure you don't access `fastjson` objects from concurrently running goroutines
         if such restriction is mentioned in [docs](https://github.com/valyala/fastjson/issues/new).
       * Build and run your program with [-race](https://golang.org/doc/articles/race_detector.html) flag.
         Make sure the race detector detects zero races.
       * If your program continue crashing after fixing issues mentioned above, [file a bug](https://github.com/valyala/fastjson/issues/new).
