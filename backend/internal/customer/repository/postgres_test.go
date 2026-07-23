package repository

import "testing"

func TestSimulationCodesAreUniqueAndStable(t *testing.T) {
	parentOne := simulationParentCode(1)
	parentTwo := simulationParentCode(2)
	if parentOne != "PC-000001" || parentTwo != "PC-000002" || parentOne == parentTwo {
		t.Fatalf("unexpected parent simulation codes: %q %q", parentOne, parentTwo)
	}
	customerOne := simulationCustomerCode(parentOne, 1)
	customerTwo := simulationCustomerCode(parentOne, 2)
	if customerOne != "PC-000001-S001" || customerTwo != "PC-000001-S002" || customerOne == customerTwo {
		t.Fatalf("unexpected customer simulation codes: %q %q", customerOne, customerTwo)
	}
}
