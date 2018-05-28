package fastjson

import (
	"sync"
)

// ParserPool may be used for pooling parsers for similarly typed JSONs.
type ParserPool struct {
	pool sync.Pool
}

// Get returns a parser from pp.
//
// The parser must be Put to pp after use.
func (pp *ParserPool) Get() *Parser {
	v := pp.pool.Get()
	if v == nil {
		return &Parser{}
	}
	return v.(*Parser)
}

// Put returns p to pp.
func (pp *ParserPool) Put(p *Parser) {
	pp.pool.Put(p)
}
