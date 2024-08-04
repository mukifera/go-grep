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

type RuneMatcherFunc func (rune) bool

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

func buildPositiveGroup(str string) (RuneMatcherFunc, error) {
	err := fmt.Errorf("unsupported pattern: %q", str)
	if len(str) < 2 { return nil, err; }
	if str[0] != '[' || str[len(str)-1] != ']' { return nil, err; }
	chars := map[rune]bool{}
	for _, c := range(str[1:len(str)-1]) {
		r := rune(c)
		chars[r] = true
	}
	fun := func (r rune) bool {
		_, ok := chars[r]
		return ok
	}

	return fun, nil
}

func matchLine(line []byte, pattern string) (bool, error) {
	if utf8.RuneCountInString(pattern) == 0 {
		return false, fmt.Errorf("unsupported pattern: %q", pattern)
	}

	var ok bool
	var fun RuneMatcherFunc

	switch {
	case pattern == "\\d": fun = isDigit; break;
	case pattern == "\\w": fun = isAlpha; break;
	case pattern[0] == '[':
		f, err := buildPositiveGroup(pattern)
		if err != nil {
			return false, err
		}
		fun = f
		break
	case isLetter(rune(pattern[0])):
		fun = func (r rune) bool {
			return r == rune(pattern[0])
		}
		break
	default:
		return false, fmt.Errorf("unsupported pattern: %q", pattern)
	}

	ok = bytes.ContainsFunc(line, fun)

	return ok, nil
}
