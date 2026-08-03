package demo

import "testing"

func TestDatasetShapeAndHierarchy(t *testing.T) {
	ds := BuildDataset()
	if err := Validate(ds); err != nil {
		t.Fatal(err)
	}
	counts := Counts(ds)
	if counts[1] != 1 {
		t.Fatalf("level 1 current assignments=%d, want 1", counts[1])
	}
	if counts[2] != 3 {
		t.Fatalf("level 2 current assignments=%d, want 3", counts[2])
	}
	if counts[3] != 8 {
		t.Fatalf("level 3 current assignments=%d, want 8", counts[3])
	}
	if counts[4] < 35 || counts[4] > 40 {
		t.Fatalf("level 4 current assignments=%d, want 35-40", counts[4])
	}
	if len(ds.Users) < 45 || len(ds.Users) > 55 {
		t.Fatalf("users=%d, want 45-55", len(ds.Users))
	}
	if len(ds.Unassigned) < 3 || len(ds.Unassigned) > 6 {
		t.Fatalf("unassigned=%d, want 3-6", len(ds.Unassigned))
	}
}

func TestDatasetHistoryAndMoves(t *testing.T) {
	ds := BuildDataset()
	var july, august, promoted, movedSupervisor, movedRegion int
	for _, assignment := range ds.Assignments {
		if assignment.From.Equal(JulyStart) {
			july++
		}
		if assignment.From.Equal(AugustStart) {
			august++
		}
		switch assignment.Scenario {
		case "promoted":
			promoted++
		case "moved-supervisor":
			movedSupervisor++
		case "moved-region":
			movedRegion++
		}
	}
	if july == 0 || august == 0 {
		t.Fatalf("expected July and August history, got july=%d august=%d", july, august)
	}
	if promoted < 6 {
		t.Fatalf("promoted assignment rows=%d, want at least 6", promoted)
	}
	if movedSupervisor < 6 {
		t.Fatalf("moved-supervisor assignment rows=%d, want at least 6", movedSupervisor)
	}
	if movedRegion < 4 {
		t.Fatalf("moved-region assignment rows=%d, want at least 4", movedRegion)
	}
}

func TestDatasetDeterministicAndDemoMarked(t *testing.T) {
	first := BuildDataset()
	second := BuildDataset()
	if len(first.Users) != len(second.Users) || len(first.Roles) != len(second.Roles) || len(first.Assignments) != len(second.Assignments) {
		t.Fatal("dataset size changed between builds")
	}
	for i := range first.Users {
		if first.Users[i].ID != second.Users[i].ID {
			t.Fatalf("user %s id is not deterministic", first.Users[i].Key)
		}
		if first.Users[i].Email == "" || first.Users[i].Email[len(first.Users[i].Email)-len(EmailDomain):] != EmailDomain {
			t.Fatalf("user %s email is not demo-marked: %s", first.Users[i].Key, first.Users[i].Email)
		}
		if len(first.Users[i].EmployeeID) < len(EmployeeIDPrefix) || first.Users[i].EmployeeID[:len(EmployeeIDPrefix)] != EmployeeIDPrefix {
			t.Fatalf("user %s employee ID is not demo-marked: %s", first.Users[i].Key, first.Users[i].EmployeeID)
		}
	}
	for i := range first.Roles {
		if first.Roles[i].ID != second.Roles[i].ID {
			t.Fatalf("role %s id is not deterministic", first.Roles[i].Key)
		}
	}
	for i := range first.Assignments {
		if first.Assignments[i].ID != second.Assignments[i].ID {
			t.Fatalf("assignment %s id is not deterministic", first.Assignments[i].Key)
		}
	}
}
