package pdp

import "testing"

func TestMapperPCAOrders(t *testing.T) {
	if totalMapperPCAOrders != len(MapperPCAOrderNames) {
		t.Errorf("Expected total number of order values to be equal to number of their names "+
			"but got totalMapperPCAOrders = %d and len(MapperPCAOrderNames) = %d",
			totalMapperPCAOrders, len(MapperPCAOrderNames))
	}
}
