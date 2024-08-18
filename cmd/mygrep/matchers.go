package main

import "errors"

type RuneMatcherFunc func (*Parser) (bool, int)
type matchersT struct{}

var Matchers matchersT

func (matchersT) Letter(p *Parser) (bool, int) {
	r := p.Peek()
	if 'a' <= r && r <= 'z' { return true, 1; }
	if 'A' <= r && r <= 'Z' { return true, 1; }
	return false, 1
}

func (matchersT) Digit(p *Parser) (bool, int) {
	r := p.Peek()
	return '0' <= r && r <= '9', 1
}

func (matchersT) Alpha(p *Parser) (bool, int) {
	if ok, n := Matchers.Digit(p); ok { return ok, n; }
	if ok, n := Matchers.Letter(p); ok { return ok, n; }
	return p.Peek() == '_', 1
}

func (matchersT) Literal(r rune) RuneMatcherFunc {
	return func (p *Parser) (bool, int) {
		return r == p.Peek(), 1
	}
}

func (matchersT) CharacterGroup(parser *Parser) (RuneMatcherFunc, error) {
	var class_funcs []RuneMatcherFunc
	positive := true
	if parser.Matches('^') {
		positive = false
	}
	for {
		if parser.Matches(']') {
			break
		}
		if parser.AtEnd() {
			return nil, errors.New("error parsing character class")
		}
		class_funcs = append(class_funcs, Matchers.Literal(parser.Advance()))
	}

	matcher := func (p *Parser) (bool, int) {
		for i := 0; i < len(class_funcs); i++ {
			if ok, n := class_funcs[i](p); ok { return positive, n; }
		}
		return !positive, 1
	}

	return matcher, nil
}

func (matchersT) StartOfString(p *Parser) (bool, int) {
	return p.current == 0, 0
}