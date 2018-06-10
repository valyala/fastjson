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
  * Validates the parsed JSON unlike [gjson](https://github.com/tidwall/gjson).
  * May parse array containing values with distinct types (aka non-homogenous types).
    For instance, `fastjson` easily parses the following JSON array `[123, "foo", [456], {"k": "v"}, null]`.


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
    This is the fastests possible solution for `encoding/json`, may be used
    for json validation. See also benchmark results for json validation.
  * `fastjson` - parse using `fastjson` without fields access.
  * `fastjson-get` - parse using `fastjson` with fields access similar to `stdjson-struct`.

```
$ GOMAXPROCS=1 go test github.com/valyala/fastjson -bench='Parse$'
goos: linux
goarch: amd64
pkg: github.com/valyala/fastjson
BenchmarkParse/small/stdjson-map         	  200000	      6782 ns/op	  28.01 MB/s	     960 B/op	      51 allocs/op
BenchmarkParse/small/stdjson-struct      	  500000	      3542 ns/op	  53.64 MB/s	     224 B/op	       4 allocs/op
BenchmarkParse/small/stdjson-empty-struct         	  500000	      2464 ns/op	  77.08 MB/s	     168 B/op	       2 allocs/op
BenchmarkParse/small/fastjson                     	 5000000	       388 ns/op	 489.36 MB/s	       0 B/op	       0 allocs/op
BenchmarkParse/small/fastjson-get                 	 2000000	       659 ns/op	 287.96 MB/s	       0 B/op	       0 allocs/op
BenchmarkParse/medium/stdjson-map                 	   50000	     38146 ns/op	  61.05 MB/s	   10194 B/op	     208 allocs/op
BenchmarkParse/medium/stdjson-struct              	   30000	     44775 ns/op	  52.01 MB/s	    9174 B/op	     258 allocs/op
BenchmarkParse/medium/stdjson-empty-struct        	  100000	     19530 ns/op	 119.25 MB/s	     280 B/op	       5 allocs/op
BenchmarkParse/medium/fastjson                    	  500000	      3319 ns/op	 701.58 MB/s	       0 B/op	       0 allocs/op
BenchmarkParse/medium/fastjson-get                	  500000	      3484 ns/op	 668.47 MB/s	       0 B/op	       0 allocs/op
BenchmarkParse/large/stdjson-map                  	    3000	    574665 ns/op	  48.93 MB/s	  210769 B/op	    2785 allocs/op
BenchmarkParse/large/stdjson-struct               	    5000	    269339 ns/op	 104.40 MB/s	   15616 B/op	     353 allocs/op
BenchmarkParse/large/stdjson-empty-struct         	   10000	    234472 ns/op	 119.92 MB/s	     280 B/op	       5 allocs/op
BenchmarkParse/large/fastjson                     	   30000	     46275 ns/op	 607.62 MB/s	       9 B/op	       0 allocs/op
BenchmarkParse/large/fastjson-get                 	   30000	     46155 ns/op	 609.20 MB/s	       9 B/op	       0 allocs/op
BenchmarkParse/canada/stdjson-map                 	      20	  64134136 ns/op	  35.10 MB/s	12260502 B/op	  392539 allocs/op
BenchmarkParse/canada/stdjson-struct              	      20	  62739259 ns/op	  35.88 MB/s	12260124 B/op	  392534 allocs/op
BenchmarkParse/canada/stdjson-empty-struct        	     100	  16367967 ns/op	 137.53 MB/s	     280 B/op	       5 allocs/op
BenchmarkParse/canada/fastjson                    	     200	   5458851 ns/op	 412.37 MB/s	  472007 B/op	     571 allocs/op
BenchmarkParse/canada/fastjson-get                	     200	   5868027 ns/op	 383.61 MB/s	  472007 B/op	     571 allocs/op
BenchmarkParse/citm/stdjson-map                   	      50	  26308047 ns/op	  65.65 MB/s	 5214088 B/op	   95402 allocs/op
BenchmarkParse/citm/stdjson-struct                	     100	  13500244 ns/op	 127.94 MB/s	    1989 B/op	      75 allocs/op
BenchmarkParse/citm/stdjson-empty-struct          	     100	  13552959 ns/op	 127.44 MB/s	     280 B/op	       5 allocs/op
BenchmarkParse/citm/fastjson                      	    1000	   1936514 ns/op	 891.91 MB/s	   17633 B/op	      30 allocs/op
BenchmarkParse/citm/fastjson-get                  	    1000	   1932256 ns/op	 893.88 MB/s	   17633 B/op	      30 allocs/op
BenchmarkParse/twitter/stdjson-map                	     100	  10643848 ns/op	  59.33 MB/s	 2186786 B/op	   31262 allocs/op
BenchmarkParse/twitter/stdjson-struct             	     300	   5113576 ns/op	 123.50 MB/s	     408 B/op	       6 allocs/op
BenchmarkParse/twitter/stdjson-empty-struct       	     300	   5115728 ns/op	 123.45 MB/s	     408 B/op	       6 allocs/op
BenchmarkParse/twitter/fastjson                   	    2000	    796466 ns/op	 792.89 MB/s	    2536 B/op	       2 allocs/op
BenchmarkParse/twitter/fastjson-get               	    2000	    797961 ns/op	 791.41 MB/s	    2536 B/op	       2 allocs/op
```

Benchmark results for json validation:

```
$ GOMAXPROCS=1 go test github.com/valyala/fastjson -bench='Validate$'
goos: linux
goarch: amd64
pkg: github.com/valyala/fastjson
BenchmarkValidate/small/stdjson 	 2000000	       869 ns/op	 218.55 MB/s	      72 B/op	       2 allocs/op
BenchmarkValidate/small/fastjson         	 5000000	       372 ns/op	 510.41 MB/s	       0 B/op	       0 allocs/op
BenchmarkValidate/medium/stdjson         	  200000	      9975 ns/op	 233.47 MB/s	     184 B/op	       5 allocs/op
BenchmarkValidate/medium/fastjson        	  500000	      3591 ns/op	 648.44 MB/s	       0 B/op	       0 allocs/op
BenchmarkValidate/large/stdjson          	   10000	    121676 ns/op	 231.09 MB/s	     184 B/op	       5 allocs/op
BenchmarkValidate/large/fastjson         	   30000	     51442 ns/op	 546.60 MB/s	       0 B/op	       0 allocs/op
BenchmarkValidate/canada/stdjson         	     200	   8385287 ns/op	 268.45 MB/s	     184 B/op	       5 allocs/op
BenchmarkValidate/canada/fastjson        	     500	   3759142 ns/op	 598.82 MB/s	       0 B/op	       0 allocs/op
BenchmarkValidate/citm/stdjson           	     200	   6892753 ns/op	 250.58 MB/s	     184 B/op	       5 allocs/op
BenchmarkValidate/citm/fastjson          	    1000	   1656027 ns/op	1042.98 MB/s	       0 B/op	       0 allocs/op
BenchmarkValidate/twitter/stdjson        	     500	   2622168 ns/op	 240.84 MB/s	     312 B/op	       6 allocs/op
BenchmarkValidate/twitter/fastjson       	    2000	   1006021 ns/op	 627.73 MB/s	       0 B/op	       0 allocs/op
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
