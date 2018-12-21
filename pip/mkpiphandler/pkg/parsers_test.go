package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeArgParsers(t *testing.T) {
	p, err := makeArgParsers([]string{
		"BoOlEaN",
		"InTeGeR",
		"NeTwOrK",
		"SeT Of dOmAiNs",
	}, "false")
	assert.NoError(t, err)
	assert.NotEmpty(t, p)

	p, err = makeArgParsers([]string{"unknown"}, "false")
	assert.EqualError(t, err, "argument 0: unknown type \"unknown\"", "parser: %#v", p)
}

func TestMakeArgParser(t *testing.T) {
	p, err := makeArgParser(0, "BoOlEaN", "false")
	assert.NoError(t, err)
	assert.NotZero(t, p)

	p, err = makeArgParser(0, "unknown", "")
	assert.EqualError(t, err, "argument 0: unknown type \"unknown\"", "parser: %q", p)
}
