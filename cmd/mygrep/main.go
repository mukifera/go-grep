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

type RuneMatcherFunc func (rune) bool

func matchLine(line []byte, pattern string) (bool, error) {

	unsupported_err := fmt.Errorf("unsupported pattern: %q", pattern)
	
	parser := NewParser(pattern)
	
	var funcs []RuneMatcherFunc
	for ; !parser.AtEnd(); {
		current_rune := parser.Advance()
		switch current_rune {
			case '\\':
				switch {
					case parser.Matches('d'): funcs = append(funcs, Matchers.Digit); break;
					case parser.Matches('w'): funcs = append(funcs, Matchers.Alpha); break;
					default:
						return false, unsupported_err
				}
				break
			case '[':
				matcher, err := Matchers.CharacterGroup(&parser)
				if err != nil {
					return false, err
				}
				funcs = append(funcs, matcher)
				break
			default:
				funcs = append(funcs, Matchers.Literal(current_rune))
				break
		}
	}

	runes := []rune(string(line))
	funcs_count := len(funcs)

	for i := 0; i + funcs_count <= len(runes); i++ {
		ok := true
		for j := 0; j < funcs_count; j++ {
			if runes[i+j] == '\n' || runes[i+j] == 0 {
				ok = false
				break
			}
			if !funcs[j](runes[i+j]) {
				ok = false
				break
			}
		}
		if ok {
			return true, nil
		}
	}

	return false, nil
}
