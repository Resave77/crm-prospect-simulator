package service

import "crm-prospect-simulator/backend/internal/admin/model"

type permissionSeed struct {
	Key         string
	Name        string
	Description string
	GroupKey    string
	ParentKey   *string
	NodeType    model.PermissionNodeType
	RoutePath   *string
	SortOrder   int
}

type PermissionCatalogEntry struct {
	Key         string
	Name        string
	Description string
	GroupKey    string
	ParentKey   *string
	NodeType    model.PermissionNodeType
	RoutePath   *string
	SortOrder   int
}

func strPtr(v string) *string { return &v }

var permissionCatalog = []permissionSeed{
	{Key: "menu_admin_dashboard", Name: "Admin Dashboard", GroupKey: "dashboard", NodeType: model.PermissionNodeMenu, RoutePath: strPtr("/admin/dashboard"), SortOrder: 10},
	{Key: "view_admin_dashboard", Name: "View Admin Dashboard", GroupKey: "dashboard", ParentKey: strPtr("menu_admin_dashboard"), NodeType: model.PermissionNodeAction, RoutePath: strPtr("/admin/dashboard"), SortOrder: 11},
	{Key: "menu_sales_dashboard", Name: "Sales Dashboard", GroupKey: "dashboard", NodeType: model.PermissionNodeMenu, RoutePath: strPtr("/sales/dashboard"), SortOrder: 12},
	{Key: "view_sales_dashboard", Name: "View Sales Dashboard", GroupKey: "dashboard", ParentKey: strPtr("menu_sales_dashboard"), NodeType: model.PermissionNodeAction, RoutePath: strPtr("/sales/dashboard"), SortOrder: 13},
	{Key: "view_team_dashboard", Name: "View Team Dashboard", GroupKey: "dashboard", ParentKey: strPtr("menu_sales_dashboard"), NodeType: model.PermissionNodeAction, SortOrder: 14},
	{Key: "menu_accounts", Name: "Accounts", GroupKey: "accounts", NodeType: model.PermissionNodeMenu, RoutePath: strPtr("/admin/accounts"), SortOrder: 20},
	{Key: "view_accounts", Name: "View Accounts", GroupKey: "accounts", ParentKey: strPtr("menu_accounts"), NodeType: model.PermissionNodeAction, RoutePath: strPtr("/admin/accounts"), SortOrder: 21},
	{Key: "view_account_detail", Name: "View Account Detail", GroupKey: "accounts", ParentKey: strPtr("menu_accounts"), NodeType: model.PermissionNodeAction, SortOrder: 22},
	{Key: "create_account", Name: "Create Account", GroupKey: "accounts", ParentKey: strPtr("menu_accounts"), NodeType: model.PermissionNodeAction, SortOrder: 23},
	{Key: "update_account", Name: "Update Account", GroupKey: "accounts", ParentKey: strPtr("menu_accounts"), NodeType: model.PermissionNodeAction, SortOrder: 24},
	{Key: "update_account_status", Name: "Update Account Status", GroupKey: "accounts", ParentKey: strPtr("menu_accounts"), NodeType: model.PermissionNodeAction, SortOrder: 25},
	{Key: "reset_account_password", Name: "Reset Account Password", GroupKey: "accounts", ParentKey: strPtr("menu_accounts"), NodeType: model.PermissionNodeAction, SortOrder: 26},
	{Key: "menu_roles", Name: "Roles", GroupKey: "roles", NodeType: model.PermissionNodeMenu, RoutePath: strPtr("/admin/role-management"), SortOrder: 30},
	{Key: "view_roles", Name: "View Roles", GroupKey: "roles", ParentKey: strPtr("menu_roles"), NodeType: model.PermissionNodeAction, RoutePath: strPtr("/admin/role-management"), SortOrder: 31},
	{Key: "view_role_detail", Name: "View Role Detail", GroupKey: "roles", ParentKey: strPtr("menu_roles"), NodeType: model.PermissionNodeAction, SortOrder: 32},
	{Key: "create_role", Name: "Create Role", GroupKey: "roles", ParentKey: strPtr("menu_roles"), NodeType: model.PermissionNodeAction, SortOrder: 33},
	{Key: "update_role", Name: "Update Role", GroupKey: "roles", ParentKey: strPtr("menu_roles"), NodeType: model.PermissionNodeAction, SortOrder: 34},
	{Key: "update_role_status", Name: "Update Role Status", GroupKey: "roles", ParentKey: strPtr("menu_roles"), NodeType: model.PermissionNodeAction, SortOrder: 35},
	{Key: "delete_role", Name: "Delete Role", GroupKey: "roles", ParentKey: strPtr("menu_roles"), NodeType: model.PermissionNodeAction, SortOrder: 36},
	{Key: "manage_role_permissions", Name: "Manage Role Permissions", GroupKey: "roles", ParentKey: strPtr("menu_roles"), NodeType: model.PermissionNodeAction, SortOrder: 37},
	{Key: "menu_sales_structure", Name: "Sales Structure", GroupKey: "sales_structure", NodeType: model.PermissionNodeMenu, RoutePath: strPtr("/admin/sales-structure"), SortOrder: 40},
	{Key: "view_sales_structure", Name: "View Sales Structure", GroupKey: "sales_structure", ParentKey: strPtr("menu_sales_structure"), NodeType: model.PermissionNodeAction, RoutePath: strPtr("/admin/sales-structure"), SortOrder: 41},
	{Key: "create_sales_assignment", Name: "Create Assignment", GroupKey: "sales_structure", ParentKey: strPtr("menu_sales_structure"), NodeType: model.PermissionNodeAction, SortOrder: 42},
	{Key: "move_sales_assignment", Name: "Move Assignment", GroupKey: "sales_structure", ParentKey: strPtr("menu_sales_structure"), NodeType: model.PermissionNodeAction, SortOrder: 43},
	{Key: "view_sales_assignment_history", Name: "View Assignment History", GroupKey: "sales_structure", ParentKey: strPtr("menu_sales_structure"), NodeType: model.PermissionNodeAction, SortOrder: 44},
	{Key: "menu_prospect_finder", Name: "Prospect Finder", GroupKey: "prospects", NodeType: model.PermissionNodeMenu, RoutePath: strPtr("/admin/prospect-finder"), SortOrder: 50},
	{Key: "menu_prospect_list", Name: "Prospect List", GroupKey: "prospects", NodeType: model.PermissionNodeMenu, RoutePath: strPtr("/admin/prospects/list"), SortOrder: 51},
	{Key: "menu_prospect_pipeline", Name: "Prospect Pipeline", GroupKey: "prospects", NodeType: model.PermissionNodeMenu, RoutePath: strPtr("/admin/prospects/pipeline"), SortOrder: 52},
	{Key: "menu_my_prospects", Name: "My Prospects", GroupKey: "prospects", NodeType: model.PermissionNodeMenu, RoutePath: strPtr("/sales/my-prospects"), SortOrder: 53},
	{Key: "menu_sales_pipeline", Name: "Sales Pipeline", GroupKey: "prospects", NodeType: model.PermissionNodeMenu, RoutePath: strPtr("/sales/pipeline"), SortOrder: 54},
	{Key: "view_prospect_finder", Name: "View Prospect Finder", GroupKey: "prospects", ParentKey: strPtr("menu_prospect_finder"), NodeType: model.PermissionNodeAction, RoutePath: strPtr("/admin/prospect-finder"), SortOrder: 55},
	{Key: "view_prospect_list", Name: "View Prospect List", GroupKey: "prospects", ParentKey: strPtr("menu_prospect_list"), NodeType: model.PermissionNodeAction, RoutePath: strPtr("/admin/prospects/list"), SortOrder: 56},
	{Key: "view_prospect_pipeline", Name: "View Prospect Pipeline", GroupKey: "prospects", ParentKey: strPtr("menu_prospect_pipeline"), NodeType: model.PermissionNodeAction, RoutePath: strPtr("/admin/prospects/pipeline"), SortOrder: 57},
	{Key: "view_prospect_detail", Name: "View Prospect Detail", GroupKey: "prospects", ParentKey: strPtr("menu_prospect_list"), NodeType: model.PermissionNodeAction, SortOrder: 58},
	{Key: "view_my_prospects", Name: "View My Prospects", GroupKey: "prospects", ParentKey: strPtr("menu_my_prospects"), NodeType: model.PermissionNodeAction, RoutePath: strPtr("/sales/my-prospects"), SortOrder: 59},
	{Key: "view_my_prospect_detail", Name: "View My Prospect Detail", GroupKey: "prospects", ParentKey: strPtr("menu_my_prospects"), NodeType: model.PermissionNodeAction, SortOrder: 60},
	{Key: "create_prospect", Name: "Create Prospect", GroupKey: "prospects", ParentKey: strPtr("menu_prospect_finder"), NodeType: model.PermissionNodeAction, SortOrder: 61},
	{Key: "update_prospect_pipeline", Name: "Update Prospect Pipeline", GroupKey: "prospects", ParentKey: strPtr("menu_sales_pipeline"), NodeType: model.PermissionNodeAction, SortOrder: 62},
	{Key: "delete_prospect", Name: "Delete Prospect", GroupKey: "prospects", ParentKey: strPtr("menu_prospect_list"), NodeType: model.PermissionNodeAction, SortOrder: 63},
	{Key: "request_prospect_deletion", Name: "Request Prospect Deletion", GroupKey: "prospects", ParentKey: strPtr("menu_my_prospects"), NodeType: model.PermissionNodeAction, SortOrder: 64},
	{Key: "approve_prospect_deletion", Name: "Approve Prospect Deletion", GroupKey: "prospects", ParentKey: strPtr("menu_prospect_list"), NodeType: model.PermissionNodeAction, SortOrder: 65},
	{Key: "reject_prospect_deletion", Name: "Reject Prospect Deletion", GroupKey: "prospects", ParentKey: strPtr("menu_prospect_list"), NodeType: model.PermissionNodeAction, SortOrder: 66},
	{Key: "convert_prospect", Name: "Convert Prospect", GroupKey: "prospects", ParentKey: strPtr("menu_prospect_list"), NodeType: model.PermissionNodeAction, SortOrder: 67},
	{Key: "manage_prospect_comments", Name: "Comments", GroupKey: "prospects", ParentKey: strPtr("menu_prospect_list"), NodeType: model.PermissionNodeAction, SortOrder: 68},
	{Key: "view_ai_summary", Name: "View AI Summary", GroupKey: "prospects", NodeType: model.PermissionNodeAction, SortOrder: 69},
	{Key: "view_ai_menu_profiling", Name: "View AI Menu Profiling", GroupKey: "prospects", NodeType: model.PermissionNodeAction, SortOrder: 70},
	{Key: "use_prospect_ai_chat", Name: "Use Prospect AI Chat", GroupKey: "prospects", NodeType: model.PermissionNodeAction, SortOrder: 71},
	{Key: "menu_customers", Name: "Customers", GroupKey: "customers", NodeType: model.PermissionNodeMenu, RoutePath: strPtr("/admin/customers"), SortOrder: 70},
	{Key: "menu_my_customers", Name: "My Customers", GroupKey: "customers", NodeType: model.PermissionNodeMenu, RoutePath: strPtr("/sales/my-customers"), SortOrder: 71},
	{Key: "view_customers", Name: "View Customers", GroupKey: "customers", ParentKey: strPtr("menu_customers"), NodeType: model.PermissionNodeAction, RoutePath: strPtr("/admin/customers"), SortOrder: 72},
	{Key: "view_customer_detail", Name: "View Customer Detail", GroupKey: "customers", ParentKey: strPtr("menu_customers"), NodeType: model.PermissionNodeAction, SortOrder: 73},
	{Key: "view_my_customers", Name: "View My Customers", GroupKey: "customers", ParentKey: strPtr("menu_my_customers"), NodeType: model.PermissionNodeAction, RoutePath: strPtr("/sales/my-customers"), SortOrder: 74},
	{Key: "view_my_customer_detail", Name: "View My Customer Detail", GroupKey: "customers", ParentKey: strPtr("menu_my_customers"), NodeType: model.PermissionNodeAction, SortOrder: 75},
	{Key: "create_customer", Name: "Create Customer", GroupKey: "customers", ParentKey: strPtr("menu_customers"), NodeType: model.PermissionNodeAction, SortOrder: 76},
	{Key: "update_customer", Name: "Update Customer", GroupKey: "customers", ParentKey: strPtr("menu_customers"), NodeType: model.PermissionNodeAction, SortOrder: 77},
	{Key: "delete_customer", Name: "Delete Customer", GroupKey: "customers", ParentKey: strPtr("menu_customers"), NodeType: model.PermissionNodeAction, SortOrder: 78},
	{Key: "view_company_detail", Name: "View Company Detail", GroupKey: "customers", ParentKey: strPtr("menu_customers"), NodeType: model.PermissionNodeAction, SortOrder: 79},
	{Key: "update_company", Name: "Update Company", GroupKey: "customers", ParentKey: strPtr("menu_customers"), NodeType: model.PermissionNodeAction, SortOrder: 80},
	{Key: "menu_visit_monitoring", Name: "Visit Monitoring", GroupKey: "visits", NodeType: model.PermissionNodeMenu, RoutePath: strPtr("/admin/visit-monitoring"), SortOrder: 90},
	{Key: "view_visit_monitoring", Name: "View Visit Monitoring", GroupKey: "visits", ParentKey: strPtr("menu_visit_monitoring"), NodeType: model.PermissionNodeAction, RoutePath: strPtr("/admin/visit-monitoring"), SortOrder: 91},
	{Key: "view_own_visits", Name: "View Own Visits", GroupKey: "visits", NodeType: model.PermissionNodeAction, SortOrder: 92},
	{Key: "check_in_prospect", Name: "Check In Prospect", GroupKey: "visits", ParentKey: strPtr("menu_my_prospects"), NodeType: model.PermissionNodeAction, SortOrder: 93},
	{Key: "check_out_prospect", Name: "Check Out Prospect", GroupKey: "visits", ParentKey: strPtr("menu_my_prospects"), NodeType: model.PermissionNodeAction, SortOrder: 94},
	{Key: "check_in_customer", Name: "Check In Customer", GroupKey: "visits", ParentKey: strPtr("menu_my_customers"), NodeType: model.PermissionNodeAction, SortOrder: 95},
	{Key: "check_out_customer", Name: "Check Out Customer", GroupKey: "visits", ParentKey: strPtr("menu_my_customers"), NodeType: model.PermissionNodeAction, SortOrder: 96},
	{Key: "update_visit_result", Name: "Update Visit Result", GroupKey: "visits", NodeType: model.PermissionNodeAction, SortOrder: 97},
	{Key: "view_visit_evidence", Name: "View Visit Evidence", GroupKey: "visits", ParentKey: strPtr("menu_visit_monitoring"), NodeType: model.PermissionNodeAction, SortOrder: 98},
	{Key: "delete_visit", Name: "Delete Visit", GroupKey: "visits", NodeType: model.PermissionNodeAction, SortOrder: 99},
	{Key: "menu_reports", Name: "Reports", GroupKey: "reports", NodeType: model.PermissionNodeMenu, RoutePath: strPtr("/admin/reports"), SortOrder: 100},
	{Key: "view_reports", Name: "View Reports", GroupKey: "reports", ParentKey: strPtr("menu_reports"), NodeType: model.PermissionNodeAction, RoutePath: strPtr("/admin/reports"), SortOrder: 101},
	{Key: "menu_sales_history", Name: "Sales History", GroupKey: "profile", NodeType: model.PermissionNodeMenu, RoutePath: strPtr("/sales/history"), SortOrder: 110},
	{Key: "view_sales_history", Name: "View Sales History", GroupKey: "profile", ParentKey: strPtr("menu_sales_history"), NodeType: model.PermissionNodeAction, RoutePath: strPtr("/sales/history"), SortOrder: 111},
	{Key: "menu_profile", Name: "Profile", GroupKey: "profile", NodeType: model.PermissionNodeMenu, RoutePath: strPtr("/sales/profile"), SortOrder: 112},
	{Key: "view_own_profile", Name: "View Own Profile", GroupKey: "profile", ParentKey: strPtr("menu_profile"), NodeType: model.PermissionNodeAction, RoutePath: strPtr("/sales/profile"), SortOrder: 113},
	{Key: "change_own_password", Name: "Change Own Password", GroupKey: "profile", NodeType: model.PermissionNodeAction, SortOrder: 114},
}

func DefaultPermissionCatalog() []PermissionCatalogEntry {
	items := make([]PermissionCatalogEntry, 0, len(permissionCatalog))
	for _, item := range permissionCatalog {
		items = append(items, PermissionCatalogEntry{
			Key:         item.Key,
			Name:        item.Name,
			Description: item.Description,
			GroupKey:    item.GroupKey,
			ParentKey:   item.ParentKey,
			NodeType:    item.NodeType,
			RoutePath:   item.RoutePath,
			SortOrder:   item.SortOrder,
		})
	}
	return items
}

var landingPagePermissions = map[string]string{
	"/admin/dashboard":          "view_admin_dashboard",
	"/admin/accounts":           "view_accounts",
	"/admin/role-management":    "view_roles",
	"/admin/sales-structure":    "view_sales_structure",
	"/admin/prospect-finder":    "view_prospect_finder",
	"/admin/prospects/list":     "view_prospect_list",
	"/admin/prospects/pipeline": "view_prospect_pipeline",
	"/admin/customers":          "view_customers",
	"/admin/visit-monitoring":   "view_visit_monitoring",
	"/admin/reports":            "view_reports",
	"/sales/dashboard":          "view_sales_dashboard",
	"/sales/my-prospects":       "view_my_prospects",
	"/sales/pipeline":           "menu_sales_pipeline",
	"/sales/my-customers":       "view_my_customers",
	"/sales/history":            "view_sales_history",
	"/sales/profile":            "view_own_profile",
}
