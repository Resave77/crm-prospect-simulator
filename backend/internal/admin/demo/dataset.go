package demo

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	EmailDomain      = "demo.yummy.local"
	EmployeeIDPrefix = "DEMO-"
	DemoPassword     = "DemoYummy2026!"
)

var (
	JulyStart   = mustDate("2026-07-01")
	JulyEnd     = mustDate("2026-07-31")
	AugustStart = mustDate("2026-08-01")
)

type Role struct {
	Key         string
	ID          uuid.UUID
	Name        string
	Level       int
	Description string
}

type User struct {
	Key        string
	ID         uuid.UUID
	EmployeeID string
	FullName   string
	Email      string
	Phone      string
	SystemRole string
	ManagerKey string
}

type Assignment struct {
	Key       string
	ID        uuid.UUID
	UserKey   string
	RoleKey   string
	ParentKey string
	From      time.Time
	To        *time.Time
	Scenario  string
}

type Dataset struct {
	Roles       []Role
	Users       []User
	Assignments []Assignment
	Unassigned  []string
}

func BuildDataset() Dataset {
	roles := []Role{
		role("level-1", "Sales Level 1", 1, "National sales leadership responsible for strategic direction and all-region performance governance."),
		role("level-2-collector", "Sales Level 2 + Collector", 2, "Regional leadership with collector coordination for account settlement and route productivity."),
		role("level-3-billing", "Sales Level 3 + Billing", 3, "Supervisor layer coordinating billing follow-up, field coaching, and regional execution quality."),
		role("level-4", "Sales Level 4", 4, "Core field sales role for outlet coverage, prospect follow-up, and daily route execution."),
		role("level-4-collector", "Sales Level 4 + Collector", 4, "Field sales role with additional collection responsibility for assigned customer routes."),
		role("level-4-merchandising", "Sales Level 4 + Merchandising", 4, "Field execution role focused on outlet merchandising, planogram compliance, and product visibility."),
		role("level-4-collector-billing", "Sales Level 4 + Collector + Billing", 4, "Field role combining route execution, collection coordination, and billing follow-up."),
		role("admin-sales", "Admin Sales", 4, "Sales administration support for order coordination, documentation, and reporting assistance."),
	}

	users := []User{
		user("director", "DEMO-DIR-001", "Bima Hartanto", "bima.hartanto", "0812-7000-0001", "SALES_MANAGER", ""),
		user("rm-jakarta", "DEMO-RM-JKT-001", "Regional Manager Jakarta - Aditya Prakoso", "aditya.prakoso", "0812-7000-0101", "SALES_MANAGER", ""),
		user("rm-west", "DEMO-RM-BDG-001", "Regional Manager Bandung - Rina Kartikasari", "rina.kartikasari", "0812-7000-0102", "SALES_MANAGER", ""),
		user("rm-east", "DEMO-RM-SBY-001", "Regional Manager Surabaya - Wisnu Mahendra", "wisnu.mahendra", "0812-7000-0103", "SALES_MANAGER", ""),
		user("spv-jakbar", "DEMO-SPV-JKT-001", "Supervisor Jakarta Barat - Dimas Saputra", "dimas.saputra", "0812-7000-0201", "SALES_MANAGER", ""),
		user("spv-jaksel", "DEMO-SPV-JKT-002", "Supervisor Jakarta Selatan - Sari Wulandari", "sari.wulandari", "0812-7000-0202", "SALES_MANAGER", ""),
		user("spv-bandung", "DEMO-SPV-BDG-001", "Supervisor Bandung Raya - Fajar Nugroho", "fajar.nugroho", "0812-7000-0203", "SALES_MANAGER", ""),
		user("spv-semarang", "DEMO-SPV-SMG-001", "Supervisor Semarang - Lestari Putri", "lestari.putri", "0812-7000-0204", "SALES_MANAGER", ""),
		user("spv-surabaya", "DEMO-SPV-SBY-001", "Supervisor Surabaya - Hendra Wijaya", "hendra.wijaya", "0812-7000-0205", "SALES_MANAGER", ""),
		user("spv-malang", "DEMO-SPV-MLG-001", "Supervisor Malang - Kartika Dewi", "kartika.dewi", "0812-7000-0206", "SALES_MANAGER", ""),
		user("spv-medan", "DEMO-SPV-MDN-001", "Supervisor Medan - Arif Setiawan", "arif.setiawan", "0812-7000-0207", "SALES_MANAGER", ""),
		user("spv-makassar", "DEMO-SPV-MKS-001", "Supervisor Makassar - Nur Aisyah", "nur.aisyah", "0812-7000-0208", "SALES_MANAGER", ""),
	}

	level4 := []struct {
		key, employeeID, name, email, phone, manager string
	}{
		{"se-jkt-001", "DEMO-SE-JKT-001", "Sales Executive Jakarta Barat - Andi Firmansyah", "andi.firmansyah", "0812-7000-1001", "spv-jakbar"},
		{"se-jkt-002", "DEMO-SE-JKT-002", "Sales Representative Jakarta Barat - Maya Salsabila", "maya.salsabila", "0812-7000-1002", "spv-jakbar"},
		{"se-jkt-003", "DEMO-SE-JKT-003", "Merchandiser Jakarta Barat - Rafi Kurniawan", "rafi.kurniawan", "0812-7000-1003", "spv-jakbar"},
		{"se-jkt-004", "DEMO-SE-JKT-004", "Key Account Executive Jakarta Barat - Citra Permata", "citra.permata", "0812-7000-1004", "spv-jakbar"},
		{"se-jks-001", "DEMO-SE-JKS-001", "Sales Executive Jakarta Selatan - Bagus Prasetyo", "bagus.prasetyo", "0812-7000-1101", "spv-jaksel"},
		{"se-jks-002", "DEMO-SE-JKS-002", "Sales Representative Jakarta Selatan - Nabila Fitri", "nabila.fitri", "0812-7000-1102", "spv-jaksel"},
		{"se-jks-003", "DEMO-SE-JKS-003", "Merchandiser Jakarta Selatan - Yoga Ramadhan", "yoga.ramadhan", "0812-7000-1103", "spv-jaksel"},
		{"se-jks-004", "DEMO-SE-JKS-004", "Admin Sales Jakarta Selatan - Melati Anggraini", "melati.anggraini", "0812-7000-1104", "spv-jaksel"},
		{"se-bdg-001", "DEMO-SE-BDG-001", "Sales Executive Bandung - Galih Pamungkas", "galih.pamungkas", "0812-7000-1201", "spv-bandung"},
		{"se-bdg-002", "DEMO-SE-BDG-002", "Sales Representative Bandung - Intan Maharani", "intan.maharani", "0812-7000-1202", "spv-bandung"},
		{"se-bdg-003", "DEMO-SE-BDG-003", "Merchandiser Bandung - Reza Maulana", "reza.maulana", "0812-7000-1203", "spv-bandung"},
		{"se-bdg-004", "DEMO-SE-BDG-004", "Collector Bandung - Yuni Astuti", "yuni.astuti", "0812-7000-1204", "spv-bandung"},
		{"se-smg-001", "DEMO-SE-SMG-001", "Sales Executive Semarang - Rendra Wibowo", "rendra.wibowo", "0812-7000-1301", "spv-semarang"},
		{"se-smg-002", "DEMO-SE-SMG-002", "Sales Representative Semarang - Tika Marlina", "tika.marlina", "0812-7000-1302", "spv-semarang"},
		{"se-smg-003", "DEMO-SE-SMG-003", "Merchandiser Semarang - Ilham Fauzi", "ilham.fauzi", "0812-7000-1303", "spv-semarang"},
		{"se-smg-004", "DEMO-SE-SMG-004", "Billing Officer Semarang - Putri Larasati", "putri.larasati", "0812-7000-1304", "spv-semarang"},
		{"se-sby-001", "DEMO-SE-SBY-001", "Sales Executive Surabaya - Eko Santoso", "eko.santoso", "0812-7000-1401", "spv-surabaya"},
		{"se-sby-002", "DEMO-SE-SBY-002", "Sales Representative Surabaya - Dewi Lestari", "dewi.lestari", "0812-7000-1402", "spv-surabaya"},
		{"se-sby-003", "DEMO-SE-SBY-003", "Merchandiser Surabaya - Farhan Hidayat", "farhan.hidayat", "0812-7000-1403", "spv-surabaya"},
		{"se-sby-004", "DEMO-SE-SBY-004", "Collector Surabaya - Tiara Amalia", "tiara.amalia", "0812-7000-1404", "spv-surabaya"},
		{"se-mlg-001", "DEMO-SE-MLG-001", "Sales Executive Malang - Hendri Cahyono", "hendri.cahyono", "0812-7000-1501", "spv-malang"},
		{"se-mlg-002", "DEMO-SE-MLG-002", "Sales Representative Malang - Ayu Puspitasari", "ayu.puspitasari", "0812-7000-1502", "spv-malang"},
		{"se-mlg-003", "DEMO-SE-MLG-003", "Merchandiser Malang - Bayu Kencana", "bayu.kencana", "0812-7000-1503", "spv-malang"},
		{"se-mlg-004", "DEMO-SE-MLG-004", "Admin Sales Malang - Niken Prameswari", "niken.prameswari", "0812-7000-1504", "spv-malang"},
		{"se-mdn-001", "DEMO-SE-MDN-001", "Sales Executive Medan - Rizal Nasution", "rizal.nasution", "0812-7000-1601", "spv-medan"},
		{"se-mdn-002", "DEMO-SE-MDN-002", "Sales Representative Medan - Cut Rahmawati", "cut.rahmawati", "0812-7000-1602", "spv-medan"},
		{"se-mdn-003", "DEMO-SE-MDN-003", "Merchandiser Medan - Fikri Anwar", "fikri.anwar", "0812-7000-1603", "spv-medan"},
		{"se-mdn-004", "DEMO-SE-MDN-004", "Collector Medan - Sinta Dewayani", "sinta.dewayani", "0812-7000-1604", "spv-medan"},
		{"se-mks-001", "DEMO-SE-MKS-001", "Sales Executive Makassar - Yusuf Ardiansyah", "yusuf.ardiansyah", "0812-7000-1701", "spv-makassar"},
		{"se-mks-002", "DEMO-SE-MKS-002", "Sales Representative Makassar - Mega Febriani", "mega.febriani", "0812-7000-1702", "spv-makassar"},
		{"se-mks-003", "DEMO-SE-MKS-003", "Merchandiser Makassar - Akbar Ramli", "akbar.ramli", "0812-7000-1703", "spv-makassar"},
		{"se-mks-004", "DEMO-SE-MKS-004", "Billing Officer Makassar - Desi Anggraeni", "desi.anggraeni", "0812-7000-1704", "spv-makassar"},
		{"se-jkt-005", "DEMO-SE-JKT-005", "Sales Representative Jakarta Barat - Wawan Gunawan", "wawan.gunawan", "0812-7000-1005", "spv-jakbar"},
		{"se-bdg-005", "DEMO-SE-BDG-005", "Key Account Executive Bandung - Hana Oktaviani", "hana.oktaviani", "0812-7000-1205", "spv-bandung"},
		{"se-sby-005", "DEMO-SE-SBY-005", "Sales Representative Surabaya - Doni Kurnia", "doni.kurnia", "0812-7000-1405", "spv-surabaya"},
		{"se-mdn-005", "DEMO-SE-MDN-005", "Merchandiser Medan - Ratu Amelia", "ratu.amelia", "0812-7000-1605", "spv-medan"},
	}
	for _, item := range level4 {
		users = append(users, user(item.key, item.employeeID, item.name, item.email, item.phone, "SALES_EXECUTIVE", item.manager))
	}
	unassigned := []string{"ua-jkt-001", "ua-bdg-001", "ua-sby-001", "ua-mks-001"}
	users = append(users,
		user("ua-jkt-001", "DEMO-UA-JKT-001", "Sales Executive Jakarta Trainee - Pramana Aji", "pramana.aji", "0812-7000-9001", "SALES_EXECUTIVE", "spv-jakbar"),
		user("ua-bdg-001", "DEMO-UA-BDG-001", "Sales Representative Bandung Trainee - Nadya Kirana", "nadya.kirana", "0812-7000-9002", "SALES_EXECUTIVE", "spv-bandung"),
		user("ua-sby-001", "DEMO-UA-SBY-001", "Merchandiser Surabaya Trainee - Oscar Wicaksono", "oscar.wicaksono", "0812-7000-9003", "SALES_EXECUTIVE", "spv-surabaya"),
		user("ua-mks-001", "DEMO-UA-MKS-001", "Collector Makassar Trainee - Ratih Permadi", "ratih.permadi", "0812-7000-9004", "SALES_EXECUTIVE", "spv-makassar"),
	)

	assignments := buildAssignments(level4)
	return Dataset{Roles: roles, Users: users, Assignments: assignments, Unassigned: unassigned}
}

func buildAssignments(level4 []struct{ key, employeeID, name, email, phone, manager string }) []Assignment {
	assignments := []Assignment{
		assignment("jul-director", "director", "level-1", "", JulyStart, nil, "same"),
		assignment("jul-rm-jakarta", "rm-jakarta", "level-2-collector", "director", JulyStart, nil, "same"),
		assignment("jul-rm-west", "rm-west", "level-2-collector", "director", JulyStart, nil, "same"),
		assignment("jul-rm-east", "rm-east", "level-2-collector", "director", JulyStart, nil, "same"),
	}
	supervisors := []struct{ key, parent string }{
		{"spv-jakbar", "rm-jakarta"}, {"spv-jaksel", "rm-jakarta"}, {"spv-bandung", "rm-west"}, {"spv-semarang", "rm-west"},
		{"spv-surabaya", "rm-east"}, {"spv-malang", "rm-east"}, {"spv-medan", "rm-west"}, {"spv-makassar", "rm-east"},
	}
	promotedUsers := map[string]bool{"spv-medan": true, "spv-makassar": true, "spv-malang": true}
	for _, s := range supervisors {
		if promotedUsers[s.key] {
			continue
		}
		assignments = append(assignments,
			assignment("jul-"+s.key, s.key, "level-3-billing", s.parent, JulyStart, nil, "same"),
		)
	}
	roleCycle := []string{"level-4", "level-4-collector", "level-4-merchandising", "level-4-collector-billing", "admin-sales"}
	promotionOldParents := map[string]string{
		"spv-malang":   "spv-surabaya",
		"spv-medan":    "spv-semarang",
		"spv-makassar": "spv-surabaya",
	}
	for i, item := range level4 {
		parent := item.manager
		julyParent := parent
		scenario := "same"
		roleKey := roleCycle[i%len(roleCycle)]
		julyRole := roleKey
		if oldParent, ok := promotionOldParents[parent]; ok {
			julyParent = oldParent
			scenario = "moved-supervisor"
		}
		if item.key == "se-jkt-003" || item.key == "se-bdg-002" || item.key == "se-sby-003" {
			julyRole = "level-4"
			scenario = "role-change"
		}
		if item.key == "se-jks-002" || item.key == "se-smg-002" || item.key == "se-mlg-002" {
			julyParent = "spv-jakbar"
			scenario = "moved-supervisor"
		}
		if item.key == "se-mdn-003" {
			julyParent = "spv-semarang"
			scenario = "moved-region"
		}
		if item.key == "se-mks-002" {
			julyParent = "spv-surabaya"
			scenario = "moved-region"
		}
		assignments = append(assignments,
			assignment("jul-"+item.key, item.key, julyRole, julyParent, JulyStart, &JulyEnd, scenario),
			assignment("aug-"+item.key, item.key, roleKey, parent, AugustStart, nil, scenario),
		)
	}
	promotions := []struct{ user, oldParent, newParent string }{
		{"spv-medan", "spv-semarang", "rm-west"},
		{"spv-makassar", "spv-surabaya", "rm-east"},
		{"spv-malang", "spv-surabaya", "rm-east"},
	}
	for _, p := range promotions {
		assignments = append(assignments,
			assignment("jul-promo-"+p.user, p.user, "level-4", p.oldParent, JulyStart, &JulyEnd, "promoted"),
			assignment("aug-promo-"+p.user, p.user, "level-3-billing", p.newParent, AugustStart, nil, "promoted"),
		)
	}
	return assignments
}

func role(key, name string, level int, description string) Role {
	return Role{Key: key, ID: deterministicID("role", key), Name: name, Level: level, Description: description}
}

func user(key, employeeID, fullName, emailLocal, phone, systemRole, managerKey string) User {
	return User{Key: key, ID: deterministicID("user", key), EmployeeID: employeeID, FullName: fullName, Email: emailLocal + "@" + EmailDomain, Phone: phone, SystemRole: systemRole, ManagerKey: managerKey}
}

func assignment(key, userKey, roleKey, parentKey string, from time.Time, to *time.Time, scenario string) Assignment {
	return Assignment{Key: key, ID: deterministicID("assignment", key), UserKey: userKey, RoleKey: roleKey, ParentKey: parentKey, From: from, To: to, Scenario: scenario}
}

func deterministicID(kind, key string) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte("crm-prospect-simulator/demo/"+kind+"/"+key))
}

func mustDate(value string) time.Time {
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		panic(err)
	}
	return parsed
}

func NormalizeRoleName(name string) string {
	return strings.ToLower(strings.Join(strings.Fields(strings.TrimSpace(name)), " "))
}

func Counts(ds Dataset) map[int]int {
	roles := map[string]int{}
	for _, role := range ds.Roles {
		roles[role.Key] = role.Level
	}
	counts := map[int]int{}
	for _, assignment := range ds.Assignments {
		if assignment.From.After(AugustStart) || (assignment.To != nil && assignment.To.Before(AugustStart)) {
			continue
		}
		counts[roles[assignment.RoleKey]]++
	}
	return counts
}

func Validate(ds Dataset) error {
	if len(ds.Users) < 45 || len(ds.Users) > 55 {
		return fmt.Errorf("demo users=%d, want 45-55", len(ds.Users))
	}
	counts := Counts(ds)
	if counts[1] != 1 || counts[2] != 3 || counts[3] != 8 || counts[4] < 35 || counts[4] > 40 {
		return fmt.Errorf("invalid current level counts: %+v", counts)
	}
	if len(ds.Unassigned) < 3 || len(ds.Unassigned) > 6 {
		return fmt.Errorf("unassigned=%d, want 3-6", len(ds.Unassigned))
	}
	userKeys := map[string]bool{}
	roleLevels := map[string]int{}
	for _, user := range ds.Users {
		userKeys[user.Key] = true
	}
	for _, role := range ds.Roles {
		roleLevels[role.Key] = role.Level
	}
	byUser := map[string][]Assignment{}
	for _, assignment := range ds.Assignments {
		if !userKeys[assignment.UserKey] {
			return fmt.Errorf("assignment %s references unknown user %s", assignment.Key, assignment.UserKey)
		}
		level := roleLevels[assignment.RoleKey]
		if level == 0 {
			return fmt.Errorf("assignment %s references unknown role %s", assignment.Key, assignment.RoleKey)
		}
		if level == 1 && assignment.ParentKey != "" {
			return fmt.Errorf("level 1 assignment %s has parent", assignment.Key)
		}
		if level > 1 {
			parentLevel := currentParentLevel(ds, assignment)
			if parentLevel != level-1 {
				return fmt.Errorf("assignment %s level %d has parent level %d", assignment.Key, level, parentLevel)
			}
		}
		byUser[assignment.UserKey] = append(byUser[assignment.UserKey], assignment)
	}
	for userKey, assignments := range byUser {
		for i := range assignments {
			for j := i + 1; j < len(assignments); j++ {
				if overlaps(assignments[i], assignments[j]) {
					return fmt.Errorf("overlap for %s between %s and %s", userKey, assignments[i].Key, assignments[j].Key)
				}
			}
		}
	}
	return nil
}

func currentParentLevel(ds Dataset, assignment Assignment) int {
	if assignment.ParentKey == "" {
		return 0
	}
	roleLevels := map[string]int{}
	for _, role := range ds.Roles {
		roleLevels[role.Key] = role.Level
	}
	for _, parentAssignment := range ds.Assignments {
		if parentAssignment.UserKey != assignment.ParentKey {
			continue
		}
		if parentAssignment.From.After(assignment.From) {
			continue
		}
		if parentAssignment.To != nil && parentAssignment.To.Before(assignment.From) {
			continue
		}
		return roleLevels[parentAssignment.RoleKey]
	}
	return 0
}

func overlaps(a, b Assignment) bool {
	aEnd := farFuture()
	bEnd := farFuture()
	if a.To != nil {
		aEnd = *a.To
	}
	if b.To != nil {
		bEnd = *b.To
	}
	return !a.From.After(bEnd) && !b.From.After(aEnd)
}

func farFuture() time.Time {
	return time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC)
}
