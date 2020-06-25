package fastjson

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

func BenchmarkParseRawString(b *testing.B) {
	for _, s := range []string{`""`, `"a"`, `"abcd"`, `"abcdefghijk"`, `"qwertyuiopasdfghjklzxcvb"`} {
		b.Run(s, func(b *testing.B) {
			benchmarkParseRawString(b, s)
		})
	}
}

func benchmarkParseRawString(b *testing.B, s string) {
	b.ReportAllocs()
	b.SetBytes(int64(len(s)))
	s = s[1:] // skip the opening '"'
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rs, tail, err := parseRawString(s)
			if err != nil {
				panic(fmt.Errorf("cannot parse %q: %s", s, err))
			}
			if rs != s[:len(s)-1] {
				panic(fmt.Errorf("invalid string obtained; got %q; want %q", rs, s[:len(s)-1]))
			}
			if len(tail) > 0 {
				panic(fmt.Errorf("non-empty tail got: %q", tail))
			}
		}
	})
}

func BenchmarkParseRawNumber(b *testing.B) {
	for _, s := range []string{"1", "1234", "123456", "-1234", "1234567890.1234567", "-1.32434e+12"} {
		b.Run(s, func(b *testing.B) {
			benchmarkParseRawNumber(b, s)
		})
	}
}

func benchmarkParseRawNumber(b *testing.B, s string) {
	b.ReportAllocs()
	b.SetBytes(int64(len(s)))
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			rn, tail, err := parseRawNumber(s)
			if err != nil {
				panic(fmt.Errorf("cannot parse %q: %s", s, err))
			}
			if rn != s {
				panic(fmt.Errorf("invalid number obtained; got %q; want %q", rn, s))
			}
			if len(tail) > 0 {
				panic(fmt.Errorf("non-empty tail got: %q", tail))
			}
		}
	})
}

func BenchmarkObjectGet(b *testing.B) {
	for _, itemsCount := range []int{10, 100, 1000, 10000, 100000} {
		b.Run(fmt.Sprintf("items_%d", itemsCount), func(b *testing.B) {
			for _, lookupsCount := range []int{0, 1, 2, 4, 8, 16, 32, 64} {
				b.Run(fmt.Sprintf("lookups_%d", lookupsCount), func(b *testing.B) {
					benchmarkObjectGet(b, itemsCount, lookupsCount)
				})
			}
		})
	}
}

func benchmarkObjectGet(b *testing.B, itemsCount, lookupsCount int) {
	b.StopTimer()
	var ss []string
	for i := 0; i < itemsCount; i++ {
		s := fmt.Sprintf(`"key_%d": "value_%d"`, i, i)
		ss = append(ss, s)
	}
	s := "{" + strings.Join(ss, ",") + "}"
	key := fmt.Sprintf("key_%d", len(ss)/2)
	expectedValue := fmt.Sprintf("value_%d", len(ss)/2)
	b.StartTimer()
	b.ReportAllocs()
	b.SetBytes(int64(len(s)))

	b.RunParallel(func(pb *testing.PB) {
		p := benchPool.Get()
		for pb.Next() {
			v, err := p.Parse(s)
			if err != nil {
				panic(fmt.Errorf("unexpected error: %s", err))
			}
			o := v.GetObject()
			for i := 0; i < lookupsCount; i++ {
				sb := o.Get(key).GetStringBytes()
				if string(sb) != expectedValue {
					panic(fmt.Errorf("unexpected value; got %q; want %q", sb, expectedValue))
				}
			}
		}
		benchPool.Put(p)
	})
}

func BenchmarkMarshalTo(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		benchmarkMarshalTo(b, smallFixture)
	})
	b.Run("medium", func(b *testing.B) {
		benchmarkMarshalTo(b, mediumFixture)
	})
	b.Run("large", func(b *testing.B) {
		benchmarkMarshalTo(b, largeFixture)
	})
	b.Run("canada", func(b *testing.B) {
		benchmarkMarshalTo(b, canadaFixture)
	})
	b.Run("citm", func(b *testing.B) {
		benchmarkMarshalTo(b, citmFixture)
	})
	b.Run("twitter", func(b *testing.B) {
		benchmarkMarshalTo(b, twitterFixture)
	})
}

func benchmarkMarshalTo(b *testing.B, s string) {
	p := benchPool.Get()
	v, err := p.Parse(s)
	if err != nil {
		panic(fmt.Errorf("unexpected error: %s", err))
	}

	b.ReportAllocs()
	b.SetBytes(int64(len(s)))
	b.RunParallel(func(pb *testing.PB) {
		var b []byte
		for pb.Next() {
			// It is ok calling v.MarshalTo from concurrent
			// goroutines, since MarshalTo doesn't modify v.
			b = v.MarshalTo(b[:0])
		}
	})
	benchPool.Put(p)
}

func BenchmarkParse(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		benchmarkParse(b, smallFixture)
	})
	b.Run("medium", func(b *testing.B) {
		benchmarkParse(b, mediumFixture)
	})
	b.Run("large", func(b *testing.B) {
		benchmarkParse(b, largeFixture)
	})
	b.Run("canada", func(b *testing.B) {
		benchmarkParse(b, canadaFixture)
	})
	b.Run("citm", func(b *testing.B) {
		benchmarkParse(b, citmFixture)
	})
	b.Run("twitter", func(b *testing.B) {
		benchmarkParse(b, twitterFixture)
	})
}

var (
	// small, medium and large fixtures are from https://github.com/buger/jsonparser/blob/f04e003e4115787c6272636780bc206e5ffad6c4/benchmark/benchmark.go
	smallFixture  = getFromFile("testdata/small.json")
	mediumFixture = getFromFile("testdata/medium.json")
	largeFixture  = getFromFile("testdata/large.json")

	// canada, citm and twitter fixtures are from https://github.com/serde-rs/json-benchmark/tree/0db02e043b3ae87dc5065e7acb8654c1f7670c43/data
	canadaFixture  = getFromFile("testdata/canada.json")
	citmFixture    = getFromFile("testdata/citm_catalog.json")
	twitterFixture = getFromFile("testdata/twitter.json")
)

func getFromFile(filename string) string {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(fmt.Errorf("cannot read %s: %s", filename, err))
	}
	return string(data)
}

func benchmarkParse(b *testing.B, s string) {
	b.Run("stdjson-map", func(b *testing.B) {
		benchmarkStdJSONParseMap(b, s)
	})
	b.Run("stdjson-struct", func(b *testing.B) {
		benchmarkStdJSONParseStruct(b, s)
	})
	b.Run("stdjson-empty-struct", func(b *testing.B) {
		benchmarkStdJSONParseEmptyStruct(b, s)
	})
	b.Run("fastjson", func(b *testing.B) {
		benchmarkFastJSONParse(b, s)
	})
	b.Run("fastjson-get", func(b *testing.B) {
		benchmarkFastJSONParseGet(b, s)
	})
}

func benchmarkFastJSONParse(b *testing.B, s string) {
	b.ReportAllocs()
	b.SetBytes(int64(len(s)))
	b.RunParallel(func(pb *testing.PB) {
		p := benchPool.Get()
		for pb.Next() {
			v, err := p.Parse(s)
			if err != nil {
				panic(fmt.Errorf("unexpected error: %s", err))
			}
			if v.Type() != TypeObject {
				panic(fmt.Errorf("unexpected value type; got %s; want %s", v.Type(), TypeObject))
			}
		}
		benchPool.Put(p)
	})
}

func benchmarkFastJSONParseGet(b *testing.B, s string) {
	b.ReportAllocs()
	b.SetBytes(int64(len(s)))
	b.RunParallel(func(pb *testing.PB) {
		p := benchPool.Get()
		var n int
		for pb.Next() {
			v, err := p.Parse(s)
			if err != nil {
				panic(fmt.Errorf("unexpected error: %s", err))
			}
			n += v.GetInt("sid")
			n += len(v.GetStringBytes("uuid"))
			p := v.Get("person")
			if p != nil {
				n++
			}
			c := v.Get("company")
			if c != nil {
				n++
			}
			u := v.Get("users")
			if u != nil {
				n++
			}
			a := v.GetArray("features")
			n += len(a)
			a = v.GetArray("topicSubTopics")
			n += len(a)
			o := v.Get("search_metadata")
			if o != nil {
				n++
			}
		}
		benchPool.Put(p)
	})
}

var benchPool ParserPool

func benchmarkStdJSONParseMap(b *testing.B, s string) {
	b.ReportAllocs()
	b.SetBytes(int64(len(s)))
	bb := s2b(s)
	b.RunParallel(func(pb *testing.PB) {
		var m map[string]interface{}
		for pb.Next() {
			if err := json.Unmarshal(bb, &m); err != nil {
				panic(fmt.Errorf("unexpected error: %s", err))
			}
		}
	})
}

func benchmarkStdJSONParseStruct(b *testing.B, s string) {
	b.ReportAllocs()
	b.SetBytes(int64(len(s)))
	bb := s2b(s)
	b.RunParallel(func(pb *testing.PB) {
		var m struct {
			Sid            int
			UUID           string
			Person         map[string]interface{}
			Company        map[string]interface{}
			Users          []interface{}
			Features       []map[string]interface{}
			TopicSubTopics map[string]interface{}
			SearchMetadata map[string]interface{}
		}
		for pb.Next() {
			if err := json.Unmarshal(bb, &m); err != nil {
				panic(fmt.Errorf("unexpected error: %s", err))
			}
		}
	})
}

func benchmarkStdJSONParseEmptyStruct(b *testing.B, s string) {
	b.ReportAllocs()
	b.SetBytes(int64(len(s)))
	bb := s2b(s)
	b.RunParallel(func(pb *testing.PB) {
		var m struct{}
		for pb.Next() {
			if err := json.Unmarshal(bb, &m); err != nil {
				panic(fmt.Errorf("unexpected error: %s", err))
			}
		}
	})
}
