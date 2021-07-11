package parsecli

import (
	"errors"
	"fmt"
	"strings"
)

func Slice(input []string) (flags map[string]string, args []string, err error) {
	flags = make(map[string]string)

	var currentIndex int

	peek := func() *string {
		if currentIndex >= len(input) {
			return nil
		}
		y := input[currentIndex]
		return &y
	}

	next := func() *string {
		y := peek()
		currentIndex += 1
		return y
	}

	var parsingArguments bool
	for peek() != nil {
		itemPointer := next()
		if itemPointer == nil {
			break
		}
		item := *itemPointer

		hyphenPrefixLength := countPrefixLength(item, '-')

		// if we've stopped getting flags and we're moving on to arguments
		if hyphenPrefixLength < 1 && !parsingArguments {
			parsingArguments = true
		}

		if parsingArguments {

			var y string
			if isStringDelimiter(rune(item[0])) {
				x, err := untilEndOfString(item, next)
				if err != nil {
					return nil, nil, err
				}
				y = x
			} else {
				y = item
			}

			args = append(args, y)

			continue
		}

		switch hyphenPrefixLength {
		case 2:
			key, value, err := flag(strings.TrimPrefix(item, "--"), next)
			if err != nil {
				return nil, nil, err
			}
			flags[strings.ToLower(key)] = value
		case 1:
			key, value, err := flag(strings.TrimPrefix(item, "-"), next)
			if err != nil {
				return nil, nil, err
			}

			if len(key) > 1 {
				// more than one key - for example `-sm=hello`
				// this should be treated as `-s -m=hello`
				for i := 0; i < len(key) - 1; i += 1 {
					char := string(key[i])
					flags[strings.ToLower(char)] = "true"
				}

				flags[strings.ToLower(string(key[len(key) - 1]))] = value
			} else {
				// just the one key - for example `-s`
				flags[strings.ToLower(key)] = value
			}
		default:
			if hyphenPrefixLength > 2 {
				return nil, nil, errors.New("flags must have a maximum of two hyphens preceding the flag name")
			} else {
				return nil, nil, errors.New("flags must have at least one hyphen preceding the flag name")
			}
		}

	}

	return flags, args, nil
}

func flag(in string, next func() *string) (key, value string, err error) {

	split := strings.Split(in, "=")

	switch len(split) {
	case 1:
		// single flag, eg `--verbose` - means `--verbose=true`
		return split[0], "true", nil
	case 2:
		// `--verbose=hello`

		val := split[1]
		if isStringDelimiter(rune(val[0])) {
			x, err := untilEndOfString(val, next)
			if err != nil {
				return "", "", err
			}
			val = x
		}

		return split[0], val, nil
	default:
		// more than one equals sign
		return "", "", fmt.Errorf("invalid format flag %#v", in)
	}
}

func isStringDelimiter(r rune) bool {
	return r == '"' || r == '\''
}

func untilEndOfString(starting string, next func() *string) (string, error) {
	// this assumes that the strings from `peek`/`next` were split by ` ` characters
	if x := starting[0]; !isStringDelimiter(rune(x)) {
		panic("starting string in untilEndOfString must begin with a double or single quote")
	}

	startingCharacter := starting[0]

	buf := []string{starting[1:]}
loop:
	for {
		nxt := next()
		if nxt == nil {
			return "", errors.New("end of string literal reached without terminating character")
		}
		switch x := strings.Index(*nxt, string(startingCharacter)); x {
		case -1:
			buf = append(buf, *nxt)
		default:
			if x != len(*nxt) - 1 {
				return "", errors.New("end of string literal must be followed by whitespace")
			}
			buf = append(buf, (*nxt)[:len(*nxt) - 1])
			break loop
		}
	}
	return strings.Join(buf, " "), nil
}

func countPrefixLength(s string, prefix rune) int {
	var c int
	for _, char := range s {
		if char == prefix {
			c += 1
		} else {
			break
		}
	}
	return c
}