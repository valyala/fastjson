package fastjson

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

func BenchmarkCompact(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		benchmarkCompact(b, smallFixture)
	})
	b.Run("medium", func(b *testing.B) {
		benchmarkCompact(b, mediumFixture)
	})
	b.Run("large", func(b *testing.B) {
		benchmarkCompact(b, largeFixture)
	})
	b.Run("canada", func(b *testing.B) {
		benchmarkCompact(b, canadaFixture)
	})
	b.Run("citm", func(b *testing.B) {
		benchmarkCompact(b, citmFixture)
	})
	b.Run("twitter", func(b *testing.B) {
		benchmarkCompact(b, twitterFixture)
	})
}

func benchmarkCompact(b *testing.B, s string) {
	b.Run("stdjson", func(b *testing.B) {
		benchmarkCompactStdJSON(b, s)
	})
	b.Run("fastjson", func(b *testing.B) {
		benchmarkCompactFastJSON(b, s)
	})
}

func benchmarkCompactStdJSON(b *testing.B, s string) {
	b.ReportAllocs()
	b.SetBytes(int64(len(s)))
	bb := s2b(s)
	b.RunParallel(func(pb *testing.PB) {
		var buf bytes.Buffer
		for pb.Next() {
			buf.Reset()
			if err := json.Compact(&buf, bb); err != nil {
				panic(fmt.Errorf("unexpected error: %s", err))
			}
		}
	})
}

func benchmarkCompactFastJSON(b *testing.B, s string) {
	b.ReportAllocs()
	b.SetBytes(int64(len(s)))
	bb := s2b(s)
	b.RunParallel(func(pb *testing.PB) {
		var dst []byte
		for pb.Next() {
			dst = dst[:0]
			var err error
			dst, err = Compact(dst, bb)
			if err != nil {
				panic(fmt.Errorf("unexpected error: %s", err))
			}
		}
	})
}
