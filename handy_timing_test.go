package fastjson

import (
	"encoding/json"
	"fmt"
	"testing"
)

func BenchmarkValidate(b *testing.B) {
	b.Run("small", func(b *testing.B) {
		benchmarkValidate(b, smallFixture)
	})
	b.Run("medium", func(b *testing.B) {
		benchmarkValidate(b, mediumFixture)
	})
	b.Run("large", func(b *testing.B) {
		benchmarkValidate(b, largeFixture)
	})
}

func benchmarkValidate(b *testing.B, s string) {
	b.Run("stdjson", func(b *testing.B) {
		benchmarkValidateStdJSON(b, s)
	})
	b.Run("fastjson", func(b *testing.B) {
		benchmarkValidateFastJSON(b, s)
	})
}

func benchmarkValidateStdJSON(b *testing.B, s string) {
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

func benchmarkValidateFastJSON(b *testing.B, s string) {
	b.ReportAllocs()
	b.SetBytes(int64(len(s)))
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if err := Validate(s); err != nil {
				panic(fmt.Errorf("unexpected error: %s", err))
			}
		}
	})
}
