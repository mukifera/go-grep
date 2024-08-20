package main

import "errors"

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

func literalMatcher(r rune) RuneMatcherFunc {
	return func (p *Parser, _ *MatcherState) (bool, int) {
		if p.AtEnd() { return false, 1 }
		return r == p.Peek(), 1
	}
}