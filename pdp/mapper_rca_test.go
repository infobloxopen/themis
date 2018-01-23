package pdp

import "testing"

func TestMapperRCAOrders(t *testing.T) {
	if totalMapperRCAOrders != len(MapperRCAOrderNames) {
		t.Errorf("Expected total number of order values to be equal to number of their names "+
			"but got totalMapperRCAOrders = %d and len(MapperRCAOrderNames) = %d",
			totalMapperRCAOrders, len(MapperRCAOrderNames))
	}
}
