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
			case '(': matcher.StartCapturingGroup(); break;
			case ')': matcher.CloseCapturingGroup(); break;
			case '\\':
				switch {
					case parser.Matches('d'): matcher.Digit(); break;
					case parser.Matches('w'): matcher.Alpha(); break;
					case parser.Matches('\\'): matcher.Literal('\\'); break;
					case IsDigit(parser.Peek()):
						matcher.Backreference(int(parser.Peek() - '0'))
						parser.Advance()
						break
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

	return matcher.MatchLine(line), nil
}
