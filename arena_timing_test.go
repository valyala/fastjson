package fastjson

import (
	"sync/atomic"
	"testing"
)

func BenchmarkArenaTypicalUse(b *testing.B) {
	// Determine the length of created object
	var aa Arena
	obj := benchCreateArenaObject(&aa)
	objLen := len(obj.String())
	b.SetBytes(int64(objLen))
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		var buf []byte
		var a Arena
		var sink int
		for pb.Next() {
			obj := benchCreateArenaObject(&a)
			buf = obj.MarshalTo(buf[:0])
			a.Reset()
			sink += len(buf)
		}
		atomic.AddUint64(&Sink, uint64(sink))
	})
}

func benchCreateArenaObject(a *Arena) *Value {
	o := a.NewObject()
	o.Set("key1", a.NewNumberInt(123))
	o.Set("key2", a.NewNumberFloat64(-1.23))

	// Create a string only once and use multuple times as a performance optimization.
	s := a.NewString("foobar")
	aa := a.NewArray()
	for i := range 10 {
		aa.SetArrayItem(i, s)
	}
	o.Set("key3", aa)
	return o
}

var Sink uint64
