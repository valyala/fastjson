[![Build Status](https://travis-ci.org/valyala/fastjson.svg)](https://travis-ci.org/valyala/fastjson)
[![GoDoc](https://godoc.org/github.com/valyala/fastjson?status.svg)](http://godoc.org/github.com/valyala/fastjson)
[![Go Report](https://goreportcard.com/badge/github.com/valyala/fastjson)](https://goreportcard.com/report/github.com/valyala/fastjson)
[![codecov](https://codecov.io/gh/valyala/fastjson/branch/master/graph/badge.svg)](https://codecov.io/gh/valyala/fastjson)

# fastjson - fast JSON parser and validator for Go


## Features

  * Fast. As usual, up to 15x faster than the standard [encoding/json](https://golang.org/pkg/encoding/json/).
    See [benchmarks](#benchmarks).
  * Parses arbitrary JSON without schema, reflection, struct magic and code generation
    contrary to [easyjson](https://github.com/mailru/easyjson).
  * Provides simple [API](http://godoc.org/github.com/valyala/fastjson).
  * Outperforms [jsonparser](https://github.com/buger/jsonparser) and [gjson](https://github.com/tidwall/gjson)
    when accessing multiple unrelated fields, since `fastjson` parses the input JSON only once.
  * Validates the parsed JSON unlike [jsonparser](https://github.com/buger/jsonparser)
    and [gjson](https://github.com/tidwall/gjson).
  * May quickly extract a part of the original JSON with `Value.Get(...).MarshalTo` and modify it
    with [Del](https://godoc.org/github.com/valyala/fastjson#Value.Del)
    and [Set](https://godoc.org/github.com/valyala/fastjson#Value.Set) functions.
  * May parse array containing values with distinct types (aka non-homogenous types).
    For instance, `fastjson` easily parses the following JSON array `[123, "foo", [456], {"k": "v"}, null]`.
  * `fastjson` preserves the original order of object items when calling
    [Object.Visit](https://godoc.org/github.com/valyala/fastjson#Object.Visit).


## Known limitations

  * Requies extra care to work with - references to certain objects recursively
    returned by [Parser](https://godoc.org/github.com/valyala/fastjson#Parser)
    must be released before the next call to [Parse](https://godoc.org/github.com/valyala/fastjson#Parser.Parse).
    Otherwise the program may work improperly.
    Adhere recommendations from [docs](https://godoc.org/github.com/valyala/fastjson).
  * Cannot parse JSON from `io.Reader`. There is [Scanner](https://godoc.org/github.com/valyala/fastjson#Scanner)
    for parsing stream of JSON values from a string.


## Usage

One-liner accessing a single field:
```go
	s := []byte(`{"foo": [123, "bar"]}`)
	fmt.Printf("foo.0=%d\n", fastjson.GetInt(s, "foo", "0"))

	// Output:
	// foo.0=123
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


## Performance optimization tips

  * Re-use [Parser](https://godoc.org/github.com/valyala/fastjson#Parser) and [Scanner](https://godoc.org/github.com/valyala/fastjson#Scanner)
    for parsing many JSONs. This reduces memory allocations overhead.
    [ParserPool](https://godoc.org/github.com/valyala/fastjson#ParserPool) may be useful in this case.
  * Prefer calling `Value.Get*` on the value returned from [Parser](https://godoc.org/github.com/valyala/fastjson#Parser)
    instead of calling `Get*` one-liners when multiple fields
    must be obtained from JSON, since each `Get*` one-liner re-parses
    the input JSON again.
  * Prefer calling once [Value.Get](https://godoc.org/github.com/valyala/fastjson#Value.Get)
    for common prefix paths and then calling `Value.Get*` on the returned value
    for distinct suffix paths.
  * Prefer iterating over array returned from [Value.GetArray](https://godoc.org/github.com/valyala/fastjson#Object.Visit)
    with a range loop instead of calling `Value.Get*` for each array item.


## Benchmarks

Go 1.11 has been used for benchmarking.

Legend:

  * `small` - parse [small.json](testdata/small.json) (190 bytes).
  * `medium` - parse [medium.json](testdata/medium.json) (2.3KB).
  * `large` - parse [large.json](testdata/large.json) (28KB).
  * `canada` - parse [canada.json](testdata/canada.json) (2.2MB).
  * `citm` - parse [citm_catalog.json](testdata/citm_catalog.json) (1.7MB).
  * `twitter` - parse [twitter.json](testdata/twitter.json) (617KB).

  * `stdjson-map` - parse into a `map[string]interface{}` using `encoding/json`.
  * `stdjson-struct` - parse into a struct containing
    a subset of fields of the parsed JSON, using `encoding/json`.
  * `stdjson-empty-struct` - parse into an empty struct using `encoding/json`.
    This is the fastest possible solution for `encoding/json`, may be used
    for json validation. See also benchmark results for json validation.
  * `fastjson` - parse using `fastjson` without fields access.
  * `fastjson-get` - parse using `fastjson` with fields access similar to `stdjson-struct`.

```
$ GOMAXPROCS=1 go test github.com/valyala/fastjson -bench='Parse$'
goos: linux
goarch: amd64
pkg: github.com/valyala/fastjson
BenchmarkParse/small/stdjson-map         	  200000	      7192 ns/op	  26.41 MB/s	     960 B/op	      51 allocs/op
BenchmarkParse/small/stdjson-struct      	  300000	      3737 ns/op	  50.83 MB/s	     224 B/op	       4 allocs/op
BenchmarkParse/small/stdjson-empty-struct         	  500000	      2600 ns/op	  73.06 MB/s	     168 B/op	       2 allocs/op
BenchmarkParse/small/fastjson                     	 5000000	       388 ns/op	 489.13 MB/s	       0 B/op	       0 allocs/op
BenchmarkParse/small/fastjson-get                 	 2000000	       646 ns/op	 294.06 MB/s	       0 B/op	       0 allocs/op
BenchmarkParse/medium/stdjson-map                 	   30000	     41063 ns/op	  56.72 MB/s	   10197 B/op	     208 allocs/op
BenchmarkParse/medium/stdjson-struct              	   30000	     48311 ns/op	  48.21 MB/s	    9174 B/op	     258 allocs/op
BenchmarkParse/medium/stdjson-empty-struct        	  100000	     21862 ns/op	 106.53 MB/s	     280 B/op	       5 allocs/op
BenchmarkParse/medium/fastjson                    	  500000	      3345 ns/op	 696.14 MB/s	       0 B/op	       0 allocs/op
BenchmarkParse/medium/fastjson-get                	  500000	      3466 ns/op	 671.90 MB/s	       0 B/op	       0 allocs/op
BenchmarkParse/large/stdjson-map                  	    2000	    612924 ns/op	  45.88 MB/s	  210708 B/op	    2785 allocs/op
BenchmarkParse/large/stdjson-struct               	    5000	    286674 ns/op	  98.08 MB/s	   15616 B/op	     353 allocs/op
BenchmarkParse/large/stdjson-empty-struct         	    5000	    256255 ns/op	 109.73 MB/s	     280 B/op	       5 allocs/op
BenchmarkParse/large/fastjson                     	   30000	     45108 ns/op	 623.34 MB/s	       9 B/op	       0 allocs/op
BenchmarkParse/large/fastjson-get                 	   30000	     45246 ns/op	 621.45 MB/s	       9 B/op	       0 allocs/op
BenchmarkParse/canada/stdjson-map                 	      20	  68012405 ns/op	  33.10 MB/s	12260502 B/op	  392539 allocs/op
BenchmarkParse/canada/stdjson-struct              	      20	  68169831 ns/op	  33.02 MB/s	12260122 B/op	  392534 allocs/op
BenchmarkParse/canada/stdjson-empty-struct        	     100	  17780456 ns/op	 126.60 MB/s	     280 B/op	       5 allocs/op
BenchmarkParse/canada/fastjson                    	     200	   5205172 ns/op	 432.47 MB/s	  382354 B/op	     571 allocs/op
BenchmarkParse/canada/fastjson-get                	     200	   5234266 ns/op	 430.06 MB/s	  382354 B/op	     571 allocs/op
BenchmarkParse/citm/stdjson-map                   	      50	  28011849 ns/op	  61.66 MB/s	 5214001 B/op	   95402 allocs/op
BenchmarkParse/citm/stdjson-struct                	     100	  14615515 ns/op	 118.18 MB/s	    1989 B/op	      75 allocs/op
BenchmarkParse/citm/stdjson-empty-struct          	     100	  14602927 ns/op	 118.28 MB/s	     280 B/op	       5 allocs/op
BenchmarkParse/citm/fastjson                      	    1000	   1972784 ns/op	 875.52 MB/s	   14652 B/op	      30 allocs/op
BenchmarkParse/citm/fastjson-get                  	    1000	   1971890 ns/op	 875.91 MB/s	   14652 B/op	      30 allocs/op
BenchmarkParse/twitter/stdjson-map                	     100	  11367144 ns/op	  55.56 MB/s	 2188391 B/op	   31267 allocs/op
BenchmarkParse/twitter/stdjson-struct             	     300	   5556788 ns/op	 113.65 MB/s	     408 B/op	       6 allocs/op
BenchmarkParse/twitter/stdjson-empty-struct       	     300	   5526507 ns/op	 114.27 MB/s	     408 B/op	       6 allocs/op
BenchmarkParse/twitter/fastjson                   	    2000	    782514 ns/op	 807.03 MB/s	    2540 B/op	       2 allocs/op
BenchmarkParse/twitter/fastjson-get               	    2000	    786814 ns/op	 802.62 MB/s	    2540 B/op	       2 allocs/op
```

Benchmark results for json validation:

```
$ GOMAXPROCS=1 go test github.com/valyala/fastjson -bench='Validate$'
goos: linux
goarch: amd64
pkg: github.com/valyala/fastjson
BenchmarkValidate/small/stdjson 	 2000000	       930 ns/op	 204.24 MB/s	      72 B/op	       2 allocs/op
BenchmarkValidate/small/fastjson         	 3000000	       402 ns/op	 472.43 MB/s	       0 B/op	       0 allocs/op
BenchmarkValidate/medium/stdjson         	  200000	     10356 ns/op	 224.88 MB/s	     184 B/op	       5 allocs/op
BenchmarkValidate/medium/fastjson        	  300000	      4005 ns/op	 581.40 MB/s	       0 B/op	       0 allocs/op
BenchmarkValidate/large/stdjson          	   10000	    125185 ns/op	 224.61 MB/s	     184 B/op	       5 allocs/op
BenchmarkValidate/large/fastjson         	   30000	     54981 ns/op	 511.41 MB/s	       0 B/op	       0 allocs/op
BenchmarkValidate/canada/stdjson         	     200	   8494224 ns/op	 265.01 MB/s	     184 B/op	       5 allocs/op
BenchmarkValidate/canada/fastjson        	     500	   3883159 ns/op	 579.70 MB/s	       0 B/op	       0 allocs/op
BenchmarkValidate/citm/stdjson           	     200	   7086020 ns/op	 243.75 MB/s	     184 B/op	       5 allocs/op
BenchmarkValidate/citm/fastjson          	    1000	   1828194 ns/op	 944.76 MB/s	       0 B/op	       0 allocs/op
BenchmarkValidate/twitter/stdjson        	     500	   2718753 ns/op	 232.28 MB/s	     312 B/op	       6 allocs/op
BenchmarkValidate/twitter/fastjson       	    2000	   1081693 ns/op	 583.82 MB/s	       0 B/op	       0 allocs/op
```

## FAQ

  * Q: _There are a ton of other high-perf packages for JSON parsing in Go. Why creating yet another package?_
    A: Because other packages require either rigid JSON schema via struct magic
       and code generation or perform poorly when multiple unrelated fields
       must be obtained from the parsed JSON.
       Additionally, `fastjson` provides nicer [API](http://godoc.org/github.com/valyala/fastjson).

  * Q: _What is the main purpose for `fastjson`?_
    A: High-perf JSON parsing for [RTB](https://www.iab.com/wp-content/uploads/2015/05/OpenRTB_API_Specification_Version_2_3_1.pdf)
       and other [JSON-RPC](https://en.wikipedia.org/wiki/JSON-RPC) services.

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
