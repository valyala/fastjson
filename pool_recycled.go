package fastjson

import (
	"sync"
)

// ParserPoolRecycled enables JSON Parser pooling for semi-structured JSON
//
// MaxReuse can be set to prevent a parser from being returned to the pool after a certain number of uses
//
// Without this protection, reusing a parser for unstructured JSON indefinitely may cause significant memory growth
type ParserPoolRecycled struct {
	sync.Pool
	maxReuse int
}

func NewParserPoolRecycled(maxReuse int) *ParserPoolRecycled {
	return &ParserPoolRecycled{
		sync.Pool{New: func() interface{} { return new(ParserRecyclable) }},
		maxReuse,
	}
}

func (p *ParserPoolRecycled) Get() *ParserRecyclable {
	return p.Pool.Get().(*ParserRecyclable)
}

func (p *ParserPoolRecycled) Put(cs *ParserRecyclable) {
	if cs.n > p.maxReuse {
		return
	}
	p.Pool.Put(cs)
}

// ParserRecyclable adds a counter to a Parser for use with ParserPoolRecycled
type ParserRecyclable struct {
	Parser
	n int
}

func (p *ParserRecyclable) Parse(s string) (*Value, error) {
	p.n++
	return p.Parser.Parse(s)
}

func (p *ParserRecyclable) ParseBytes(b []byte) (*Value, error) {
	p.n++
	return p.Parser.ParseBytes(b)
}

// ScannerPoolRecycled enables JSON Scanner pooling for semi-structured JSON
//
// MaxReuse can be set to prevent a scanner from being returned to the pool after a certain number of uses
//
// Without this protection, reusing a scanner for unstructured JSON indefinitely may cause significant memory growth
type ScannerPoolRecycled struct {
	sync.Pool
	maxReuse int
}

func NewScannerPoolRecycled(maxReuse int) *ScannerPoolRecycled {
	return &ScannerPoolRecycled{
		sync.Pool{New: func() interface{} { return new(ScannerRecyclable) }},
		maxReuse,
	}
}

func (p *ScannerPoolRecycled) Get() *ScannerRecyclable {
	return p.Pool.Get().(*ScannerRecyclable)
}

func (p *ScannerPoolRecycled) Put(cs *ScannerRecyclable) {
	if cs.n > p.maxReuse {
		return
	}
	p.Pool.Put(cs)
}

// ScannerRecyclable adds a counter to a Scanner for use with ScannerPoolRecycled
type ScannerRecyclable struct {
	Scanner
	n int
}

func (p *ScannerRecyclable) Init(s string) {
	p.n++
	p.Scanner.Init(s)
}

func (p *ScannerRecyclable) InitBytes(b []byte) {
	p.n++
	p.Scanner.InitBytes(b)
}
