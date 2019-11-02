package fastjson

import (
	"fmt"
	"sync"
	"testing"
)

func TestParserPoolRecycled(t *testing.T) {
	var news int
	ppr := ParserPoolRecycled{
		sync.Pool{New: func() interface{} { news++; return new(ParserRecyclable) }},
		100,
	}
	for i := 333; i > 0; i-- {
		v := ppr.Get()
		ppr.Put(v)
	}
	if news > 4 {
		t.Fatalf("Expected exactly 4 calls to Put (not %d)", news)
	}
}

func TestScannerPoolRecycled(t *testing.T) {
	var news int
	spr := ScannerPoolRecycled{
		sync.Pool{New: func() interface{} { news++; return new(ScannerRecyclable) }},
		100,
	}
	for i := 333; i > 0; i-- {
		v := spr.Get()
		spr.Put(v)
	}
	if news > 4 {
		t.Fatalf("Expected exactly 4 calls to Put (not %d)", news)
	}
}

func BenchmarkParserPoolRecycled(b *testing.B) {
	for _, n := range []int{0, 100, 10000, 1000000} {
		b.Run(fmt.Sprintf("maxreuse_%d", n), func(b *testing.B) {
			benchmarkParserPoolRecycled(b, n)
		})
	}
}

func benchmarkParserPoolRecycled(b *testing.B, maxReuse int) {
	b.ReportAllocs()
	spr := NewParserPoolRecycled(maxReuse)
	var v interface{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			v = spr.Get()
			spr.Put(v.(*ParserRecyclable))
		}
	})
	_ = v
}

func BenchmarkScannerPoolRecycled(b *testing.B) {
	for _, n := range []int{0, 100, 10000, 1000000} {
		b.Run(fmt.Sprintf("maxreuse_%d", n), func(b *testing.B) {
			benchmarkScannerPoolRecycled(b, n)
		})
	}
}

func benchmarkScannerPoolRecycled(b *testing.B, maxReuse int) {
	b.ReportAllocs()
	spr := NewScannerPoolRecycled(maxReuse)
	var v interface{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			v = spr.Get()
			spr.Put(v.(*ScannerRecyclable))
		}
	})
	_ = v
}
