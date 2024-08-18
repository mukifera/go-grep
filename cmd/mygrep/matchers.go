package main

import "errors"

type RuneMatcherFunc func (*Parser) (bool, int)
type matchersT struct{}

type MatcherState struct {
	rune_index int
	matcher_node *MatcherNode
}

func NewMatcherState(rune_index int, matcher_node *MatcherNode) MatcherState {
	return MatcherState{
		rune_index: rune_index,
		matcher_node: matcher_node,
	}
}

type MatcherNode struct {
	matcher_func RuneMatcherFunc
	next []*MatcherNode
}

type MatcherList struct {
	head *MatcherNode
	tail *MatcherNode
}

func NewMatcherList() MatcherList {
	var list MatcherList
	list.head = nil
	list.tail = nil
	return list
}

func (list *MatcherList) AddNode(f RuneMatcherFunc) {
	var node MatcherNode
	node.matcher_func = f
	if list.tail != nil {
		list.tail.next = append(list.tail.next, &node)
	}
	list.tail = &node
	if list.head == nil {
		list.head = list.tail
	}
}

var Matchers matchersT

func (matchersT) Letter(p *Parser) (bool, int) {
	if p.AtEnd() { return false, 1 }
	r := p.Peek()
	if 'a' <= r && r <= 'z' { return true, 1; }
	if 'A' <= r && r <= 'Z' { return true, 1; }
	return false, 1
}

func (matchersT) Digit(p *Parser) (bool, int) {
	if p.AtEnd() { return false, 1 }
	r := p.Peek()
	return '0' <= r && r <= '9', 1
}

func (matchersT) Alpha(p *Parser) (bool, int) {
	if ok, n := Matchers.Digit(p); ok { return ok, n; }
	if ok, n := Matchers.Letter(p); ok { return ok, n; }
	if p.AtEnd() { return false, 1 }
	return p.Peek() == '_', 1
}

func (matchersT) Literal(r rune) RuneMatcherFunc {
	return func (p *Parser) (bool, int) {
		if p.AtEnd() { return false, 1 }
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
		if p.AtEnd() { return false, 1; }
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

func (matchersT) EndOfString(p *Parser) (bool, int) {
	return p.AtEnd() || p.Peek() == '\n', 0
}