A regex matcher written in Go. Functionality is a subset of [`grep`](https://en.wikipedia.org/wiki/Grep)

## How to run

Use the `grep-go.sh` script to build and run the matcher. The command will exit with code 0 on a successful match, and exit code 1 otherwise.

### Usage
```sh
echo -n [text] | ./grep-go.sh -E [regex]
```

### Example
```sh
echo -n "abcd" | ./grep-go.sh -E "ab" # exits with code 0
echo -n "abcd" | ./grep-go.sh -E "z" # exits with code 1
```

## Supported syntax
- Character literals
- Digit class `\d` and the alphanumeric class `\w`
- Positive character groups `[abc]`
- Negative character groups `[^abc]`
- Start of string anchor `^`, and end of string anchor `$`
- One or more quantifier `+`
- Zero or one quantifier `?`
- Wild card `.`
- Alternation to combine multiple patterns `(abc|def)`
- Backreferences up to 9 groups, with nesting allowed `(\w+) and \1` 