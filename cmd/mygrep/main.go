package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"unicode/utf8"
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

func isLetter(r rune) bool {
	if 'a' <= r && r <= 'z' { return true; }
	if 'A' <= r && r <= 'Z' { return true; }
	return false
}

func isDigit(r rune) bool {
	return '0' <= r && r <= '9'
}

func isAlpha(r rune) bool {
	return isDigit(r) || isLetter(r) || r == '_'
}

func matchLine(line []byte, pattern string) (bool, error) {
	if utf8.RuneCountInString(pattern) > 2 {
		return false, fmt.Errorf("unsupported pattern: %q", pattern)
	}

	var ok bool
	var fun func (rune) bool

	switch {
	case pattern == "\\d": fun = isDigit; break;
	case pattern == "\\w": fun = isAlpha; break;
	case isLetter(rune(pattern[0])):
		fun = func (r rune) bool {
			return r == rune(pattern[0])
		}
		break
	default: fun = nil; break;
	}

	ok = bytes.ContainsFunc(line, fun)

	return ok, nil
}
