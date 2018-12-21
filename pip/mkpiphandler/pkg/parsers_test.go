package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParserNameMap(t *testing.T) {
	for k := range typeMap {
		_, ok := parserNameMap[k]
		assert.True(t, ok, "missing marshaller for type %q", k)
	}

	for k := range parserNameMap {
		_, ok := typeMap[k]
		assert.True(t, ok, "extra marshaller for unknown type %q", k)
	}
}

func TestMakeArgParsers(t *testing.T) {
	p := makeArgParsers([]string{
		"BoOlEaN",
		"InTeGeR",
		"NeTwOrK",
		"SeT Of dOmAiNs",
	}, "false")
	assert.NotEmpty(t, p)

	p = makeArgParsers(nil, "nil")
	assert.Empty(t, p)
}

func TestMakeArgParser(t *testing.T) {
	p := makeArgParser(0, "BoOlEaN", "false", false)
	assert.NotZero(t, p)
	assert.Contains(t, p, "in,")

	p = makeArgParser(0, "BoOlEaN", "false", true)
	assert.NotZero(t, p)
	assert.Contains(t, p, "_,")
}
