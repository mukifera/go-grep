package main

import "errors"

type RuneMatcherFunc func (*Parser) (bool, int)

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
	prev *MatcherNode
	is_sink bool
}

func NewMatcherNode(f RuneMatcherFunc) MatcherNode {
	return MatcherNode{
		matcher_func: f,
		is_sink: false,
	}
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

func (list *MatcherList) AppendNode(node *MatcherNode) {
	if list.tail != nil {
		list.tail.next = append(list.tail.next, node)
		node.prev = list.tail
	}
	list.tail = node
	if list.head == nil {
		list.head = list.tail
	}
}

type Matcher struct {
	list *MatcherList
	group_heads []*MatcherNode
	group_sinks []*MatcherNode
}

func NewMatcher() Matcher {
	var matcher Matcher
	list := NewMatcherList()
	matcher.list = &list
	return matcher
}

func (matcher *Matcher) Start() {
	matcher.StartGroup()
}

func (matcher *Matcher) End() {
	matcher.CloseGroup()
}

func (matcher *Matcher) StartGroup() {
	f := func (*Parser) (bool, int) {
		return true, 0
	}
	head := NewMatcherNode(f)
	sink := NewMatcherNode(f)
	sink.is_sink = true
	sink.prev = &head
	matcher.group_heads = append(matcher.group_heads, &head)
	matcher.group_sinks = append(matcher.group_sinks, &sink)
	matcher.list.AppendNode(&head)
}

func (matcher *Matcher) CloseGroup() {
	group_i := len(matcher.group_heads) - 1
	matcher.list.AppendNode(matcher.group_sinks[group_i])
	matcher.group_heads = matcher.group_heads[:group_i]
	matcher.group_sinks = matcher.group_sinks[:group_i]
}

func (matcher *Matcher) AppendMatcher(f RuneMatcherFunc) {
	node := NewMatcherNode(f)
	matcher.list.AppendNode(&node)
}

func (matcher *Matcher) Letter() {
	f := func (p *Parser) (bool, int) {
		if p.AtEnd() { return false, 1 }
		return isLetter(p.Peek()), 1
	}
	matcher.AppendMatcher(f)
}

func (matcher *Matcher) Digit() {
	f := func (p *Parser) (bool, int) {
		if p.AtEnd() { return false, 1 }
		return isDigit(p.Peek()), 1
	}
	matcher.AppendMatcher(f)
}

func (matcher *Matcher) Alpha() {
	f := func (p *Parser) (bool, int) {
		if p.AtEnd() { return false, 1 }
		r := p.Peek()
		return isDigit(r) || isLetter(r) || r == '_', 1
	}
	matcher.AppendMatcher(f)
}

func (matcher *Matcher) Literal(r rune) {
	f := literalMatcher(r)
	matcher.AppendMatcher(f)
}

func (matcher *Matcher) CharacterGroup(parser *Parser) error {
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
			return errors.New("error parsing character class")
		}
		class_funcs = append(class_funcs, literalMatcher(parser.Advance()))
	}

	f := func (p *Parser) (bool, int) {
		if p.AtEnd() { return false, 1; }
		for i := 0; i < len(class_funcs); i++ {
			if ok, n := class_funcs[i](p); ok { return positive, n; }
		}
		return !positive, 1
	}
	matcher.AppendMatcher(f)
	return nil
}

func (matcher *Matcher) StartAnchor() {
	f := func (p *Parser) (bool, int) {
		return p.current == 0, 0
	}
	matcher.AppendMatcher(f)
}

func (matcher *Matcher) EndAnchor() {
	f := func (p *Parser) (bool, int) {
		return p.AtEnd() || p.Peek() == '\n', 0
	}
	matcher.AppendMatcher(f)
}

func (matcher *Matcher) OneOrMore() {
	tail := matcher.list.tail
	if tail.is_sink {
		tail.next = append(tail.next, tail.prev)
	} else {
		tail.next = append(tail.next, tail)
	}
}

func (matcher *Matcher) ZeroOrOne() {
	node := NewMatcherNode(func (*Parser) (bool, int) { return true, 0; })
	node.prev = matcher.list.tail.prev
	matcher.list.tail.prev.next = append(matcher.list.tail.prev.next, &node)
	matcher.list.AppendNode(&node)
}

func (matcher *Matcher) WildCard() {
	f := func (p *Parser) (bool, int) {
		return !p.AtEnd() && p.Peek() != '\n', 1
	}
	matcher.AppendMatcher(f)
}

func (matcher *Matcher) Alternate() {
	group_i := len(matcher.group_heads) - 1
	matcher.list.tail.next = append(matcher.list.tail.next, matcher.group_sinks[group_i])
	matcher.list.tail = matcher.group_heads[group_i]
}

func isLetter(r rune) bool {
	if 'a' <= r && r <= 'z' { return true; }
	if 'A' <= r && r <= 'Z' { return true; }
	return false
}

func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

func literalMatcher(r rune) RuneMatcherFunc {
	return func (p *Parser) (bool, int) {
		if p.AtEnd() { return false, 1 }
		return r == p.Peek(), 1
	}
}