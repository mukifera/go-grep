package main

import "errors"

type matchersT struct{}

var Matchers matchersT

func (matchersT) Letter(r rune) bool {
	if 'a' <= r && r <= 'z' { return true; }
	if 'A' <= r && r <= 'Z' { return true; }
	return false
}

func (matchersT) Digit(r rune) bool {
	return '0' <= r && r <= '9'
}

func (matchersT) Alpha(r rune) bool {
	return Matchers.Digit(r) || Matchers.Letter(r) || r == '_'
}

func (matchersT) Literal(r rune) func(rune)bool {
	return func (rr rune) bool {
		return r == rr
	}
}

func (matchersT) CharacterGroup(parser *Parser) (RuneMatcherFunc, error) {
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
			return nil, errors.New("error parsing character class")
		}
		class_funcs = append(class_funcs, Matchers.Literal(parser.Advance()))
	}

	matcher := func (r rune) bool {
		for i := 0; i < len(class_funcs); i++ {
			if class_funcs[i](r) { return positive; }
		}
		return !positive
	}

	return matcher, nil
}