package main

import "errors"

type RuneMatcherFunc func (*Parser, *MatcherState) (bool, int)

type MatcherState struct {
	rune_current int
	matched_starts []int
	matched_ends []int
	matcher_node *MatcherNode
}

func NewMatcherState(rune_current int, matcher_node *MatcherNode) MatcherState {
	return MatcherState{
		rune_current: rune_current,
		matcher_node: matcher_node,
	}
}

func CopyMatcherState(state *MatcherState) MatcherState {
	copiedState := NewMatcherState(state.rune_current, state.matcher_node)
	copiedState.matched_starts = append(copiedState.matched_starts, state.matched_starts...)
	copiedState.matched_ends = append(copiedState.matched_ends, state.matched_ends...)
	return copiedState
}

type MatcherNode struct {
	matcher_func RuneMatcherFunc
	next []*MatcherNode
	prev *MatcherNode
	is_sink bool
	is_capturing bool
}

func NewMatcherNode(f RuneMatcherFunc) MatcherNode {
	return MatcherNode{
		matcher_func: f,
		is_sink: false,
		is_capturing: false,
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
	f := func (*Parser, *MatcherState) (bool, int) {
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
	f := func (p *Parser, _ *MatcherState) (bool, int) {
		if p.AtEnd() { return false, 1 }
		return IsLetter(p.Peek()), 1
	}
	matcher.AppendMatcher(f)
}

func (matcher *Matcher) Digit() {
	f := func (p *Parser, _ *MatcherState) (bool, int) {
		if p.AtEnd() { return false, 1 }
		return IsDigit(p.Peek()), 1
	}
	matcher.AppendMatcher(f)
}

func (matcher *Matcher) Alpha() {
	f := func (p *Parser, _ *MatcherState) (bool, int) {
		if p.AtEnd() { return false, 1 }
		r := p.Peek()
		return IsDigit(r) || IsLetter(r) || r == '_', 1
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

	f := func (p *Parser, s *MatcherState) (bool, int) {
		if p.AtEnd() { return false, 1; }
		for i := 0; i < len(class_funcs); i++ {
			if ok, n := class_funcs[i](p, s); ok { return positive, n; }
		}
		return !positive, 1
	}
	matcher.AppendMatcher(f)
	return nil
}

func (matcher *Matcher) StartAnchor() {
	f := func (p *Parser, _ *MatcherState) (bool, int) {
		return p.current == 0, 0
	}
	matcher.AppendMatcher(f)
}

func (matcher *Matcher) EndAnchor() {
	f := func (p *Parser, _ *MatcherState) (bool, int) {
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
	node := NewMatcherNode(func (*Parser, *MatcherState) (bool, int) { return true, 0; })
	node.prev = matcher.list.tail.prev
	matcher.list.tail.prev.next = append(matcher.list.tail.prev.next, &node)
	matcher.list.AppendNode(&node)
}

func (matcher *Matcher) WildCard() {
	f := func (p *Parser, _ *MatcherState) (bool, int) {
		return !p.AtEnd() && p.Peek() != '\n', 1
	}
	matcher.AppendMatcher(f)
}

func (matcher *Matcher) Alternate() {
	group_i := len(matcher.group_heads) - 1
	matcher.list.tail.next = append(matcher.list.tail.next, matcher.group_sinks[group_i])
	matcher.list.tail = matcher.group_heads[group_i]
}

func (matcher *Matcher) Backreference(group_i int) {
	group_i -= 1
	f := func (p *Parser, s *MatcherState) (bool, int) {
		if group_i >= len(s.matched_starts) || group_i >= len(s.matched_ends) {
			return false, 0
		}
		start := s.matched_starts[group_i]
		end := s.matched_ends[group_i]
		length := end - start
		for i := 0; i < length; i++ {
			if p.Seek(p.current + i) != p.Seek(start + i) {
				return false, 0
			}
		}
		return true, length
	}
	matcher.AppendMatcher(f)
}

func (matcher *Matcher) StartCapturingGroup() {
	matcher.StartGroup()
	group_i := len(matcher.group_sinks) - 1
	matcher.group_heads[group_i].is_capturing = true
	matcher.group_sinks[group_i].is_capturing = true
}

func (matcher *Matcher) CloseCapturingGroup() {
	matcher.CloseGroup()
}

func (matcher *Matcher) MatchLine(line []byte) bool {

	parser := NewParser(string(line))
	
	for i := 0; i <= len(parser.contents); i++ {
		matched := false
		states := []MatcherState{ NewMatcherState(i, matcher.list.head) } 
		for ; len(states) != 0;{
			var new_states []MatcherState
			for _, state := range(states) {
				parser.current = state.rune_current
				ok, n := state.matcher_node.matcher_func(&parser, &state)
				if ok {
					if len(state.matcher_node.next) == 0 {
						matched = true
						break
					}
					new_state := CopyMatcherState(&state)
					new_state.rune_current += n
					if state.matcher_node.is_capturing {
						if state.matcher_node.is_sink {
							new_state.matched_ends = append(new_state.matched_ends, state.rune_current)
						} else {
							new_state.matched_starts = append(new_state.matched_starts, state.rune_current)
						}
					}
					for _, next := range(state.matcher_node.next) {
						copy := CopyMatcherState(&new_state)
						copy.matcher_node = next
						new_states = append(new_states, copy)
					}
					parser.current += n
				}
			}
			if matched {
				break
			}
			states = new_states
		}
		if matched {
			return true
		}
	}
	return false
}

func IsLetter(r rune) bool {
	if 'a' <= r && r <= 'z' { return true; }
	if 'A' <= r && r <= 'Z' { return true; }
	return false
}

func IsDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

func literalMatcher(r rune) RuneMatcherFunc {
	return func (p *Parser, _ *MatcherState) (bool, int) {
		if p.AtEnd() { return false, 1 }
		return r == p.Peek(), 1
	}
}