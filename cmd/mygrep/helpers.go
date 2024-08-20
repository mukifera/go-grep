package main

func IsLetter(r rune) bool {
	if 'a' <= r && r <= 'z' { return true; }
	if 'A' <= r && r <= 'Z' { return true; }
	return false
}

func IsDigit(r rune) bool {
	return '0' <= r && r <= '9'
}