package main

import (
	"fmt"
	"io"
	"os"
)

// Usage: echo <input_text> | your_program.sh -E <pattern>
func main() {
	if len(os.Args) < 3 || os.Args[1] != "-E" {
		fmt.Fprintf(os.Stderr, "usage: mygrep -E <pattern>\n")
		os.Exit(2) // 1 means no lines were selected, >1 means error
	}

	pattern := os.Args[2]

	line, err := io.ReadAll(os.Stdin) // assume we're only dealing with a single line
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: read input text: %v\n", err)
		os.Exit(2)
	}

	ok, err := matchLine(line, pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	if !ok {
		os.Exit(1)
	}

	// default exit code is 0 which means success
}

func matchLine(line []byte, pattern string) (bool, error) {

	unsupported_err := fmt.Errorf("unsupported pattern: %q", pattern)
	
	parser := NewParser(pattern)
	
	matcher_list := NewMatcherList()

	for ; !parser.AtEnd(); {
		current_rune := parser.Advance()
		switch current_rune {
			case '^': matcher_list.AddNode(Matchers.StartOfString); break;
			case '$': matcher_list.AddNode(Matchers.EndOfString); break;
			case '+': matcher_list.tail.next = append(matcher_list.tail.next, matcher_list.tail)
			case '\\':
				switch {
					case parser.Matches('d'): matcher_list.AddNode(Matchers.Digit); break;
					case parser.Matches('w'): matcher_list.AddNode(Matchers.Alpha); break;
					case parser.Matches('\\'): matcher_list.AddNode(Matchers.Literal('\\')); break;
					default:
						return false, unsupported_err
				}
				break
			case '[':
				matcher, err := Matchers.CharacterGroup(&parser)
				if err != nil {
					return false, err
				}
				matcher_list.AddNode(matcher)
				break
			default:
				matcher_list.AddNode(Matchers.Literal(current_rune))
				break
		}
	}

	parser = NewParser(string(line))
	
	for i := 0; i <= len(parser.contents); i++ {
		matched := false
		var states []MatcherState
		states = append(states, NewMatcherState(i, matcher_list.head))
		for ; len(states) != 0;{
			var new_states []MatcherState
			for _, state := range(states) {
				parser.current = state.rune_index
				ok, n := state.matcher_node.matcher_func(&parser)
				if ok {
					if len(state.matcher_node.next) == 0 {
						matched = true
						break
					}
					for _, next := range(state.matcher_node.next) {
						new_states = append(new_states, NewMatcherState(state.rune_index + n, next))
					}
				}
			}
			if matched {
				break
			}
			states = new_states
		}
		if matched {
			return true, nil
		}
	}

	return false, nil
}
