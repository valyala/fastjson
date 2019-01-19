package fastjson

import (
	"strconv"
)

// Resets the parser and clears all the values, making them invalid.
// Call before calling parser.New* methods to avoid memory leak.
func (p *Parser) Reset() {
	p.c.reset()
}

// NewObject returns a new Value with the parameter as its initial content.
//
// The returned value is valid until the next call to Parse* or Reset.
func (p *Parser) NewObject(m map[string]*Value) *Value {
	o := p.c.getValue()
	o.t = TypeObject
	for k, v := range m {
		kv := o.o.getKV()
		kv.k = k
		kv.v = v
	}
	return o
}

// NewObject returns a new Value with the parameter as its initial content.
//
// This function is slower than the Parser.NewObject for re-used Parser.
func NewObject(m map[string]*Value) *Value {
	o := new(Value)
	o.t = TypeObject
	for k, v := range m {
		kv := o.o.getKV()
		kv.k = k
		kv.v = v
	}
	return o
}

// NewBool returns a new Value with the parameter as its initial content.
//
// The returned value is valid until the next call to Parse* or Reset.
func (p *Parser) NewBool(b bool) *Value {
	v := p.c.getValue()
	if b {
		v.t = TypeTrue
	} else {
		v.t = TypeFalse
	}
	return v
}

// NewBool returns a new Value with the parameter as its initial content.
//
// This function is slower than the Parser.NewBool for re-used Parser.
func NewBool(b bool) *Value {
	v := new(Value)
	if b {
		v.t = TypeTrue
	} else {
		v.t = TypeFalse
	}
	return v
}

// NewArray returns a new Value with the parameter as its initial content.
// The parameter is then owned by returned Value and must not be re-used.
//
// The returned value is valid until the next call to Parse* or Reset.
func (p *Parser) NewArray(a []*Value) *Value {
	o := p.c.getValue()
	o.t = TypeArray
	o.a = a
	return o
}

// NewArray returns a new Value with the parameter as its initial content.
// The parameter is then owned by returned Value and must not be re-used.
//
// This function is slower than the Parser.NewArray for re-used Parser.
func NewArray(a []*Value) *Value {
	o := new(Value)
	o.t = TypeArray
	o.a = a
	return o
}

// NewString returns a new Value with the parameter as its initial content.
//
// The returned value is valid until the next call to Parse* or Reset.
func (p *Parser) NewString(s string) *Value {
	o := p.c.getValue()
	o.t = TypeString
	o.s = s
	return o
}

// NewString returns a new Value with the parameter as its initial content.
//
// This function is slower than the Parser.NewString for re-used Parser.
func NewString(s string) *Value {
	o := new(Value)
	o.t = TypeString
	o.s = s
	return o
}

// NewInt returns a new Value with the parameter as its initial content.
func (p *Parser) NewInt(v int) *Value {
	o := p.c.getValue()
	o.t = TypeNumber
	o.s = strconv.FormatInt(int64(v), 10)
	return o
}

// NewInt returns a new Value with the parameter as its initial content.
//
// This function is slower than the Parser.NewInt64 for re-used Parser.
func NewInt(v int) *Value {
	o := new(Value)
	o.t = TypeNumber
	o.s = strconv.FormatInt(int64(v), 10)
	return o
}

// NewFloat64 returns a new Value with the parameter as its initial content.
//
// The returned value is valid until the next call to Parse* or Reset.
func (p *Parser) NewFloat64(v float64) *Value {
	o := p.c.getValue()
	o.t = TypeNumber
	o.s = strconv.FormatFloat(v, 'G', -1, 64)
	return o
}

// NewFloat64 returns a new Value with the parameter as its initial content.
//
// This function is slower than the Parser.NewFloat64 for re-used Parser.
func NewFloat64(v float64) *Value {
	o := new(Value)
	o.t = TypeNumber
	o.s = strconv.FormatFloat(v, 'G', -1, 64)
	return o
}

// NewNull returns a new Value with null.
//
// The returned value is valid until the next call to Parse* or Reset.
func (p *Parser) NewNull() *Value {
	o := p.c.getValue()
	o.t = TypeNull
	return o
}

// NewNull returns a new Value with null.
//
// This function is slower than the Parser.NewNull for re-used Parser.
func NewNull() *Value {
	o := new(Value)
	o.t = TypeNull
	return o
}
