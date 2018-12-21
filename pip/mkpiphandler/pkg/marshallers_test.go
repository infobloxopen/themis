package pkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarshallerMap(t *testing.T) {
	for k := range typeMap {
		_, ok := marshallerMap[k]
		assert.True(t, ok, "missing marshaller for type %q", k)
	}

	for k := range marshallerMap {
		_, ok := typeMap[k]
		assert.True(t, ok, "extra marshaller for unknown type %q", k)
	}
}
