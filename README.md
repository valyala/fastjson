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
    must be released before then next call to [Parse](https://godoc.org/github.com/valyala/fastjson#Parser.Parse).
    Otherwise the program may work improperly and/or may crash.
    Adhere recommendations from [docs](https://godoc.org/github.com/valyala/fastjson).
  * Cannot parse JSON from `io.Reader`. There is [Scanner](https://godoc.org/github.com/valyala/fastjson#Scanner)
    for parsing stream of JSON values from a string.


## Benchmarks

Legend:

  * `small` - small-length parsing. JSON length is 190 bytes.
  * `medium` - medium-length parsing. JSON length is 2.4KB.
  * `large` - large-length parsing. JSON length is 24KB.
  * `stdjson-map` - `encoding/json`, parsing into a `map[string]interface{}`
  * `stdjson-struct` - `encoding/json`, parsing into a struct containing
    a subset of fields of the parsed JSON.
  * `stdjson-empty-struct` - `encoding/json`, parsing into an empty struct.
    This is the fastests possible solution for `encoding/json`, may be used
    for json validation.
  * `fastjson` - standard fastjson parsing.

```
$ GOMAXPROCS=1 go test github.com/valyala/fastjson -bench=.
goos: linux
goarch: amd64
pkg: github.com/valyala/fastjson
BenchmarkParse/small/stdjson-map         	  200000	      7079 ns/op	  26.70 MB/s	     960 B/op	      51 allocs/op
BenchmarkParse/small/stdjson-struct      	  500000	      3256 ns/op	  58.04 MB/s	     224 B/op	       4 allocs/op
BenchmarkParse/small/stdjson-empty-struct         	  500000	      2490 ns/op	  75.89 MB/s	     168 B/op	       2 allocs/op
BenchmarkParse/small/fastjson                     	 3000000	       482 ns/op	 391.81 MB/s	       0 B/op	       0 allocs/op
BenchmarkParse/medium/stdjson-map                 	   30000	     40381 ns/op	  57.65 MB/s	   10196 B/op	     208 allocs/op
BenchmarkParse/medium/stdjson-struct              	   30000	     47130 ns/op	  49.39 MB/s	    9174 B/op	     258 allocs/op
BenchmarkParse/medium/stdjson-empty-struct        	  100000	     20798 ns/op	 111.93 MB/s	     280 B/op	       5 allocs/op
BenchmarkParse/medium/fastjson                    	  300000	      4218 ns/op	 551.84 MB/s	       0 B/op	       0 allocs/op
BenchmarkParse/large/stdjson-map                  	    2000	    618105 ns/op	  45.50 MB/s	  210758 B/op	    2785 allocs/op
BenchmarkParse/large/stdjson-struct               	    5000	    283911 ns/op	  99.06 MB/s	   15616 B/op	     353 allocs/op
BenchmarkParse/large/stdjson-empty-struct         	    5000	    249950 ns/op	 112.51 MB/s	     280 B/op	       5 allocs/op
BenchmarkParse/large/fastjson                     	   30000	     56235 ns/op	 500.09 MB/s	       9 B/op	       0 allocs/op
PASS
ok  	github.com/valyala/fastjson	19.762s
```

As you can see, `fastsjon` outperforms `encoding/json`:

  * by a factor of 15x for `small`-length parsing;
  * by a factor of 11x for `medium`-length and `large`-length parsing.


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

  * Q: _`fastjson` crashes my program!_ A: There is high probability of improper use.
       * Make sure you don't hold references to objects recursively returned by `Parser` / `Scanner`
         beyond the next `Parser.Parse` / `Scanner.Next` call
         if such restriction is mentioned in [docs](https://github.com/valyala/fastjson/issues/new).
       * Make sure you don't access `fastjson` objects from concurrently running goroutines
         if such restriction is mentioned in [docs](https://github.com/valyala/fastjson/issues/new).
       * If you program continue crashing after fixing issues mentioned above, [file a bug](https://github.com/valyala/fastjson/issues/new).
