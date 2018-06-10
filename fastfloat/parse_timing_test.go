package fastfloat

import (
	"fmt"
	"strconv"
	"sync/atomic"
	"testing"
)

func BenchmarkParseBestEffort(b *testing.B) {
	for _, s := range []string{"0", "12", "12345", "1234567890", "1234.45678", "1234e45", "12.34e-34"} {
		b.Run(s, func(b *testing.B) {
			benchmarkParseBestEffort(b, s)
		})
	}
}

func benchmarkParseBestEffort(b *testing.B, s string) {
	b.Run("std", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(s)))
		b.RunParallel(func(pb *testing.PB) {
			var f float64
			for pb.Next() {
				ff, err := strconv.ParseFloat(s, 64)
				if err != nil {
					panic(fmt.Errorf("unexpected error: %s", err))
				}
				f += ff
			}
			atomic.AddUint64(&Sink, uint64(f))
		})
	})
	b.Run("custom", func(b *testing.B) {
		b.ReportAllocs()
		b.SetBytes(int64(len(s)))
		b.RunParallel(func(pb *testing.PB) {
			var f float64
			for pb.Next() {
				f += ParseBestEffort(s)
			}
			atomic.AddUint64(&Sink, uint64(f))
		})
	})
}

var Sink uint64
