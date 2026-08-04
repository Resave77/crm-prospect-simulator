package main

import "testing"

func TestOfficialSalesRolesAreProtected(t *testing.T) {
	official := []string{
		"Super Admin",
		"Sales Level 2 + Collector",
		"Sales Level 3 + Collector",
		"Billing",
		"Sales Level 4",
		"Sales Level 4 + Collector",
		"Sales Level 4 + Merchandising",
		"Sales Level 4 + Collector + Billing",
	}
	for _, name := range official {
		if !isOfficialSalesRole(name) {
			t.Fatalf("%q is not protected", name)
		}
	}
}

func TestObsoleteSalesRolesAreNotProtected(t *testing.T) {
	obsolete := []string{
		"Sales Level 1",
		"Sales Level 2",
		"Sales Level 3",
		"Sales Level 3 + Billing",
	}
	for _, name := range obsolete {
		if isOfficialSalesRole(name) {
			t.Fatalf("%q should not be protected", name)
		}
	}
}
