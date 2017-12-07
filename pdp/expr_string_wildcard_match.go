package pdp

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

type functionStringWildcardMatch struct {
	pattern Expression
	str     Expression
}

func makeFunctionStringWildcardMatch(pattern, str Expression) Expression {
	return functionStringWildcardMatch{
		pattern: pattern,
		str:     str}
}

func makeFunctionStringWildcardMatchAlt(args []Expression) Expression {
	if len(args) != 2 {
		panic(fmt.Errorf("function \"wildcard-match\" for String needs exactly two arguments but got %d", len(args)))
	}

	return makeFunctionStringWildcardMatch(args[0], args[1])
}

func (f functionStringWildcardMatch) GetResultType() int {
	return TypeBoolean
}

func (f functionStringWildcardMatch) describe() string {
	return "wildcard-match"
}

func (f functionStringWildcardMatch) calculate(ctx *Context) (AttributeValue, error) {
	pattern, err := ctx.calculateStringExpression(f.pattern)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "pattern"), f.describe())
	}

	str, err := ctx.calculateStringExpression(f.str)
	if err != nil {
		return undefinedValue, bindError(bindError(err, "string"), f.describe())
	}

	ok, err := wildcardMatch(pattern, str)
	if err != nil {
		return undefinedValue, bindError(err, f.describe())
	}
	return MakeBooleanValue(ok), nil
}

func functionStringWildcardMatchValidator(args []Expression) functionMaker {
	if len(args) != 2 || args[0].GetResultType() != TypeString || args[1].GetResultType() != TypeString {
		return nil
	}

	return makeFunctionStringWildcardMatchAlt
}

const (
	wcmNoTerm = iota
	wcmAnyTerm
	wcmOneTerm
)

var (
	wcmStarLength         = utf8.RuneLen('*')
	wcmQuestionMarkLength = utf8.RuneLen('?')
	wcmSlashLength        = utf8.RuneLen('\\')
)

type patternChunk struct {
	text string
	next int
	term int
}

var errInvalidWildcardPattern = errors.New("invalid wildcards match pattern")

func wcmGetNextChunk(pattern string) (patternChunk, error) {
	chunk := ""
	escaped := false
	start := 0
	for i, c := range pattern {
		switch c {
		default:
			escaped = false

		case '*':
			if !escaped {
				if start > 0 {
					return patternChunk{
						text: chunk + pattern[start:i],
						next: i + wcmStarLength,
						term: wcmAnyTerm,
					}, nil
				}

				return patternChunk{
					text: pattern[:i],
					next: i + wcmStarLength,
					term: wcmAnyTerm,
				}, nil
			}

			escaped = false

		case '?':
			if !escaped {
				if start > 0 {
					return patternChunk{
						text: chunk + pattern[start:i],
						next: i + wcmQuestionMarkLength,
						term: wcmOneTerm,
					}, nil
				}

				return patternChunk{
					text: pattern[:i],
					next: i + wcmQuestionMarkLength,
					term: wcmOneTerm,
				}, nil
			}

			escaped = false

		case '\\':
			if escaped {
				escaped = false
			} else {
				end := i - wcmSlashLength + 1
				if end > start {
					chunk += pattern[start:end]
				}
				start = end + wcmSlashLength
				escaped = true
			}
		}
	}

	if escaped {
		return patternChunk{}, errInvalidWildcardPattern
	}

	if start > 0 {
		return patternChunk{
			text: chunk + pattern[start:],
			next: -1,
			term: wcmNoTerm,
		}, nil
	}

	return patternChunk{
		text: pattern,
		next: -1,
		term: wcmNoTerm,
	}, nil
}

func wildcardMatch(pattern, str string) (bool, error) {
	p := pattern
	term := wcmNoTerm
	for len(p) > 0 {
		if term == wcmOneTerm {
			_, size := utf8.DecodeRuneInString(str)
			if size <= 0 {
				return false, nil
			}

			str = str[size:]
			term = wcmNoTerm
		}

		chunk, err := wcmGetNextChunk(p)
		if err != nil {
			return false, newInvalidWildcardPattern(pattern)
		}

		if len(chunk.text) > 0 {
			i := 0
			if term == wcmAnyTerm {
				i = strings.Index(str, chunk.text)
				if i < 0 {
					return false, nil
				}
			} else if !strings.HasPrefix(str, chunk.text) {
				return false, nil
			}

			i += len(chunk.text)
			str = str[i:]
		}

		if chunk.next >= 0 && chunk.next < len(p) {
			p = p[chunk.next:]
		} else {
			p = ""
		}
		term = chunk.term
	}

	switch term {
	case wcmOneTerm:
		_, size := utf8.DecodeRuneInString(str)
		if size > 0 {
			str = str[size:]
			_, size = utf8.DecodeRuneInString(str)
			return size <= 0, nil
		}

		return false, nil

	case wcmAnyTerm:
		return true, nil
	}

	return len(str) <= 0, nil
}
