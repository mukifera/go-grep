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
	matcher := NewMatcher()

	matcher.Start()
	for ; !parser.AtEnd(); {
		current_rune := parser.Advance()
		switch current_rune {
			case '^': matcher.StartAnchor(); break;
			case '$': matcher.EndAnchor(); break;
			case '+': matcher.OneOrMore(); break;
			case '?': matcher.ZeroOrOne(); break;
			case '.': matcher.WildCard(); break;
			case '|': matcher.Alternate(); break;
			case '(': matcher.StartGroup(); break;
			case ')': matcher.CloseGroup(); break;
			case '\\':
				switch {
					case parser.Matches('d'): matcher.Digit(); break;
					case parser.Matches('w'): matcher.Alpha(); break;
					case parser.Matches('\\'): matcher.Literal('\\'); break;
					default:
						return false, unsupported_err
				}
				break
			case '[':
				err := matcher.CharacterGroup(&parser)
				if err != nil {
					return false, err
				}
				break
			default: matcher.Literal(current_rune); break;
		}
	}
	matcher.End()

	parser = NewParser(string(line))
	
	for i := 0; i <= len(parser.contents); i++ {
		matched := false
		var states []MatcherState
		states = append(states, NewMatcherState(i, matcher.list.head))
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
