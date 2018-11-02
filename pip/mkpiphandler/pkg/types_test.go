package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMakeGoTypeList(t *testing.T) {
	goTypes, err := makeGoTypeList([]string{
		"BoOlEaN",
		"InTeGeR",
		"NeTwOrK",
		"SeT Of dOmAiNs",
	})
	assert.NoError(t, err)
	assert.Equal(t, []string{
		goTypeBool,
		goTypeInt64,
		goTypeNetIPNet,
		goTypeDomainTree,
	}, goTypes)
}

func TestMakeGoTypeListWithUnknownType(t *testing.T) {
	goTypes, err := makeGoTypeList([]string{"unknown"})
	assert.EqualError(t, err, "argument 0: unknown type \"unknown\"", "go types: %#v", goTypes)
}

func TestGetGoType(t *testing.T) {
	goType, goZero, err := getGoType("BoOlEaN")
	assert.NoError(t, err)
	assert.Equal(t, goTypeBool, goType)
	assert.Equal(t, "false", goZero)
}

func TestGetGoTypeWithUnknownType(t *testing.T) {
	goType, goZero, err := getGoType("unknown")
	assert.EqualError(t, err, "result: unknown type \"unknown\"", "go type: %q, go zero: %q", goType, goZero)
}
