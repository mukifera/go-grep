package main

type RuneMatcherFunc func (*Parser, *MatcherState) (bool, int)

type MatcherState struct {
	rune_current int
	match_groups int
	matched_starts []int
	matched_ends []int
	unmatched_starts []int
	matcher_node *MatcherNode
}

func newMatcherState(rune_current int, matcher_node *MatcherNode) MatcherState {
	return MatcherState{
		rune_current: rune_current,
		match_groups: 0,
		matcher_node: matcher_node,
	}
}

func copyMatcherState(state *MatcherState) MatcherState {
	copied_state := newMatcherState(state.rune_current, state.matcher_node)
	copied_state.match_groups = state.match_groups
	copied_state.matched_starts = append(copied_state.matched_starts, state.matched_starts...)
	copied_state.matched_ends = append(copied_state.matched_ends, state.matched_ends...)
	copied_state.unmatched_starts = append(copied_state.unmatched_starts, state.unmatched_starts...)
	return copied_state
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

type matcherList struct {
	head *MatcherNode
	tail *MatcherNode
}

func newMatcherList() matcherList {
	var list matcherList
	list.head = nil
	list.tail = nil
	return list
}

func (list *matcherList) AppendNode(node *MatcherNode) {
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
	list *matcherList
	group_heads []*MatcherNode
	group_sinks []*MatcherNode
}

func NewMatcher() Matcher {
	var matcher Matcher
	list := newMatcherList()
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

func (matcher *Matcher) MatchLine(line []byte) bool {

	parser := NewParser(string(line))
	
	for i := 0; i <= len(parser.contents); i++ {
		matched := false
		states := []MatcherState{ newMatcherState(i, matcher.list.head) } 
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
					new_state := copyMatcherState(&state)
					new_state.rune_current += n
					if state.matcher_node.is_capturing {
						if state.matcher_node.is_sink {
							length := len(state.unmatched_starts)
							group_i := state.unmatched_starts[length-1]
							new_state.unmatched_starts = state.unmatched_starts[:length-1]
							new_state.matched_ends[group_i] = state.rune_current
						} else {
							new_state.matched_starts = append(new_state.matched_starts, state.rune_current)
							new_state.matched_ends = append(new_state.matched_ends, state.rune_current)
							new_state.unmatched_starts = append(new_state.unmatched_starts, new_state.match_groups)
							new_state.match_groups = state.match_groups + 1
						}
					}
					for _, next := range(state.matcher_node.next) {
						copy := copyMatcherState(&new_state)
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