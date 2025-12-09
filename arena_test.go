package fastjson

import (
	"fmt"
	"testing"
	"time"
)

func TestArena(t *testing.T) {
	t.Run("serial", func(t *testing.T) {
		var a Arena
		for i := 0; i < 10; i++ {
			if err := testArena(&a); err != nil {
				t.Fatal(err)
			}
			a.Reset()
		}
	})
	t.Run("concurrent", func(t *testing.T) {
		var ap ArenaPool
		workers := 4
		ch := make(chan error, workers)
		for i := 0; i < workers; i++ {
			go func() {
				a := ap.Get()
				defer ap.Put(a)
				var err error
				for i := 0; i < 10; i++ {
					if err = testArena(a); err != nil {
						break
					}
				}
				ch <- err
			}()
		}
		for i := 0; i < workers; i++ {
			select {
			case err := <-ch:
				if err != nil {
					t.Fatal(err)
				}
			case <-time.After(time.Second):
				t.Fatalf("timeout")
			}
		}
	})
}

func testArena(a *Arena) error {
	o := a.NewObject()
	o.Set("nil1", a.NewNull())
	o.Set("nil2", nil)
	o.Set("false", a.NewFalse())
	o.Set("true", a.NewTrue())
	ni := a.NewNumberInt(123)
	o.Set("ni", ni)
	o.Set("nf", a.NewNumberFloat64(1.23))
	o.Set("ns", a.NewNumberString("34.43"))
	s := a.NewString("foo")
	o.Set("str1", s)
	o.Set("str2", a.NewStringBytes([]byte("xx")))

	aa := a.NewArray()
	aa.SetArrayItem(0, s)
	aa.Set("1", ni)
	aa.SetArrayItem(2, nil)
	aa.SetArrayItem(3, a.NewNull())
	o.Set("a", aa)
	obj := a.NewObject()
	obj.Set("s", s)
	o.Set("obj", obj)

	str := o.String()
	strExpected := `{"nil1":null,"nil2":null,"false":false,"true":true,"ni":123,"nf":1.23,"ns":34.43,"str1":"foo","str2":"xx","a":["foo",123,null,null],"obj":{"s":"foo"}}`
	if str != strExpected {
		return fmt.Errorf("unexpected json\ngot\n%s\nwant\n%s", str, strExpected)
	}
	return nil
}
