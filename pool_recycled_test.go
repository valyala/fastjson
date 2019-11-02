// +build !race

// The behavior of sync.Pool is not deterministic under race mode

package fastjson

import (
	"fmt"
	"testing"
	"math/rand"
	"sync"
)

func TestParserPoolRecycled(t *testing.T) {
	var news int
	ppr := &ParserPoolRecycled{
		sync.Pool{New: func() interface{} { news++; return new(ParserRecyclable) }},
		100,
	}
	var v *Value
	for i := 333; i > 0; i-- {
		var json = []byte(fmt.Sprintf(`{"%d":"test"}`, i))
		pr := ppr.Get()
		v, _ = pr.ParseBytes(json)
		ppr.Put(pr)
	}
	_ = v
	if news != 4 {
		t.Fatalf("Expected exactly 4 calls to Put (not %d)", news)
	}
}

func TestScannerPoolRecycled(t *testing.T) {
	var news int
	spr := &ScannerPoolRecycled{
		sync.Pool{New: func() interface{} { news++; return new(ScannerRecyclable) }},
		100,
	}
	var v *Value
	for i := 333; i > 0; i-- {
		var json = []byte(fmt.Sprintf(`{"%d":"test"}`, i))
		sr := spr.Get()
		sr.InitBytes(json)
		v = sr.Value()
		spr.Put(sr)
	}
	_ = v
	if news != 4 {
		t.Fatalf("Expected exactly 4 calls to Put (not %d)", news)
	}
}

func BenchmarkParserPoolRecycled(b *testing.B) {
	for _, n := range []int{0, 10, 1000} {
		b.Run(fmt.Sprintf("maxreuse_%d", n), func(b *testing.B) {
			benchmarkParserPoolRecycled(b, n)
		})
	}
}

func benchmarkParserPoolRecycled(b *testing.B, maxReuse int) {
	b.ReportAllocs()
	ppr := NewParserPoolRecycled(maxReuse)
	var v *Value
	for i := b.N; i > 0; i-- {
		var json = []byte(fmt.Sprintf(`{"%d":"test"}`, i))
		pr := ppr.Get()
		v, _ = pr.ParseBytes(json)
		ppr.Put(pr)
	}
	_ = v
}

func BenchmarkScannerPoolRecycled(b *testing.B) {
	for _, n := range []int{0, 10, 1000} {
		b.Run(fmt.Sprintf("maxreuse_%d", n), func(b *testing.B) {
			benchmarkScannerPoolRecycled(b, n)
		})
	}
}

func benchmarkScannerPoolRecycled(b *testing.B, maxReuse int) {
	b.ReportAllocs()
	spr := NewScannerPoolRecycled(maxReuse)
	var v *Value
	for i := b.N; i > 0; i-- {
		var json = []byte(fmt.Sprintf(`{"%d":"test","foo":"bar}`, rand.Int()))
		sr := spr.Get()
		sr.InitBytes(json)
		v = sr.Value()
		spr.Put(sr)
	}
	_ = v
}
