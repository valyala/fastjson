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
	var v2 *Value
	for i := 333; i > 0; i-- {
		var json = fmt.Sprintf(`{"%d":"test"}`, i)
		pr := ppr.Get()
		v, _ = pr.Parse(json)
		v2, _ = pr.ParseBytes([]byte(json))
		ppr.Put(pr)
	}
	if news != 7 {
		t.Fatalf("Expected exactly 7 calls to Put (not %d)", news)
	}
	ppr = NewParserPoolRecycled(10)
	if ppr.maxReuse != 10 {
		t.Fatalf("Expected maxReuse to be 10 (not %d)", ppr.maxReuse)
	}
	pr := ppr.Get()
	_ = pr
	_ = v
	_ = v2
}

func TestScannerPoolRecycled(t *testing.T) {
	var news int
	spr := &ScannerPoolRecycled{
		sync.Pool{New: func() interface{} { news++; return new(ScannerRecyclable) }},
		100,
	}
	var v *Value
	var v2 *Value
	for i := 333; i > 0; i-- {
		var json = fmt.Sprintf(`{"%d":"test"}`, i)
		sr := spr.Get()
		sr.Init(json)
		v = sr.Value()
		sr.InitBytes([]byte(json))
		v2 = sr.Value()
		spr.Put(sr)
	}
	if news != 7 {
		t.Fatalf("Expected exactly 7 calls to Put (not %d)", news)
	}
	spr = NewScannerPoolRecycled(10)
	if spr.maxReuse != 10 {
		t.Fatalf("Expected maxReuse to be 10 (not %d)", spr.maxReuse)
	}
	sr := spr.Get()
	_ = sr
	_ = v
	_ = v2
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
		var json = fmt.Sprintf(`{"%d":"test"}`, i)
		pr := ppr.Get()
		v, _ = pr.ParseBytes([]byte(json))
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
		var json = fmt.Sprintf(`{"%d":"test","foo":"bar}`, rand.Int())
		sr := spr.Get()
		sr.InitBytes([]byte(json))
		v = sr.Value()
		spr.Put(sr)
	}
	_ = v
}
