package main

type Parser struct {
	contents []rune
	current int
}

func NewParser(contents string) Parser {
	var parser Parser
	parser.contents = []rune(contents)
	parser.current = 0
	return parser
}

func (p *Parser) Advance() rune {
	if p.AtEnd() { return 0; }
	r := p.contents[p.current]
	p.current += 1
	return r
}

func (p *Parser) Peek() rune {
	return p.contents[p.current]
}

func (p *Parser) Matches(r rune) bool {
	if p.Peek() == r {
		p.Advance()
		return true
	}
	return false
}

func (p *Parser) AtEnd() bool {
	return p.current >= len(p.contents)
}