# Permission Inventory Audit

Phase 8D-0 audit only. This document inventories current routes, menus, API actions, and role-page gaps for a future Role Permission Explorer.

## 1. Current Route Inventory

Current frontend routing is system-role based. `ADMINISTRATOR` route meta also permits `SUPER_ADMIN` through `roleAllowed()`.

| Module | Display name | Route name | Path | View file | Layout | Current roles | Sidebar | Menu key | View key | Landing |
|---|---|---|---|---|---|---|---|---|---|---|
| Auth | Login | Login | `/login` | `views/Login/LoginView.vue` | None | Public | No | - | - | No |
| Auth | Change Password | ChangePassword | `/change-password` | `views/Auth/ChangePasswordView.vue` | None | Authenticated | No | - | `change_own_password` | No |
| System | Forbidden | Forbidden | `/forbidden` | `views/NotFoundView.vue` | None | Public | No | - | - | No |
| Admin | Dashboard | AdminDashboard | `/admin/dashboard` | `views/Admin/Dashboard/AdminDashboardView.vue` | AdminLayout | SUPER_ADMIN, ADMINISTRATOR | Yes | `menu_admin_dashboard` | `view_admin_dashboard` | Yes |
| Accounts | Account List | AdminAccounts | `/admin/accounts` | `views/Admin/Accounts/AdminAccountsView.vue` | AdminLayout | SUPER_ADMIN, ADMINISTRATOR | Yes | `menu_accounts` | `view_accounts` | Yes |
| Accounts | Create Account | AdminAccountCreate | `/admin/accounts/create` | `views/Admin/Accounts/AdminAccountCreateView.vue` | AdminLayout | SUPER_ADMIN, ADMINISTRATOR | No | - | `create_account` | No |
| Accounts | Account Detail | AdminAccountDetail | `/admin/accounts/:id` | `views/Admin/Accounts/AdminAccountDetailView.vue` | AdminLayout | SUPER_ADMIN, ADMINISTRATOR | No | - | `view_account_detail` | No |
| Accounts | Edit Account | AdminAccountEdit | `/admin/accounts/:id/edit` | `views/Admin/Accounts/AdminAccountEditView.vue` | AdminLayout | SUPER_ADMIN, ADMINISTRATOR | No | - | `update_account` | No |
| Roles | Role Management | AdminRoleManagement | `/admin/role-management` | `views/Admin/Roles/AdminRoleManagementView.vue` | AdminLayout | SUPER_ADMIN, ADMINISTRATOR | Yes | `menu_roles` | `view_roles` | Yes |
| Sales Organization | Sales Structure | AdminSalesStructure | `/admin/sales-structure` | `views/Admin/Roles/AdminSalesStructureView.vue` | AdminLayout | SUPER_ADMIN, ADMINISTRATOR | Yes | `menu_sales_structure` | `view_sales_structure` | Yes |
| Prospect | Prospect Finder | AdminProspectFinder | `/admin/prospect-finder` | `views/Admin/Prospect/ProspectFinderView.vue` | AdminLayout | SUPER_ADMIN, ADMINISTRATOR | Yes | `menu_prospect_finder` | `view_prospect_finder` | Yes |
| Prospect | Prospect Pipeline | AdminProspectPipeline | `/admin/prospects/pipeline` | `views/Admin/Prospect/ProspectPipelineView.vue` | AdminLayout | SUPER_ADMIN, ADMINISTRATOR | Yes | `menu_prospect_pipeline` | `view_prospect_pipeline` | Yes |
| Prospect | Prospect List | AdminProspectList | `/admin/prospects/list` | `views/Admin/Prospect/ProspectListView.vue` | AdminLayout | SUPER_ADMIN, ADMINISTRATOR | Yes | `menu_prospect_list` | `view_prospect_list` | Yes |
| Prospect | Prospect Review | AdminProspectReview | `/admin/prospects/:id/review` | `views/Admin/Prospect/ProspectReviewView.vue` | AdminLayout | SUPER_ADMIN, ADMINISTRATOR | No | - | `view_prospect_detail` | No |
| Prospect | Prospect Conversion | AdminProspectConvert | `/admin/prospects/:id/convert` | `views/Admin/Prospect/ProspectConversionView.vue` | AdminLayout | SUPER_ADMIN, ADMINISTRATOR | No | - | `convert_prospect` | No |
| Customer | Customer List | AdminCustomers | `/admin/customers` | `views/Admin/Customer/CustomerListView.vue` | AdminLayout | SUPER_ADMIN, ADMINISTRATOR | Yes | `menu_customers` | `view_customers` | Yes |
| Customer | Add Customer | AdminCustomerAdd | `/admin/customers/add` | `views/Admin/Customer/CustomerAddView.vue` | AdminLayout | SUPER_ADMIN, ADMINISTRATOR | No | - | `create_customer` | No |
| Customer | Customer Detail | AdminCustomerDetail | `/admin/customers/:id` | `views/Admin/Customer/CustomerDetailView.vue` | AdminLayout | SUPER_ADMIN, ADMINISTRATOR | No | - | `view_customer_detail` | No |
| Customer | Edit Customer | AdminCustomerEdit | `/admin/customers/:id/edit` | `views/Admin/Customer/CustomerEditView.vue` | AdminLayout | SUPER_ADMIN, ADMINISTRATOR | No | - | `update_customer` | No |
| Company | Add Company | AdminCompanyAdd | `/admin/companies/add` | `views/Admin/Company/CompanyAddView.vue` | AdminLayout | SUPER_ADMIN, ADMINISTRATOR | No | - | `create_company` | No |
| Company | Company Detail | AdminCompanyDetail | `/admin/companies/:id` | `views/Admin/Company/CompanyDetailView.vue` | AdminLayout | SUPER_ADMIN, ADMINISTRATOR | No | - | `view_company_detail` | No |
| Company | Edit Company | AdminCompanyEdit | `/admin/companies/:id/edit` | `views/Admin/Company/CompanyEditView.vue` | AdminLayout | SUPER_ADMIN, ADMINISTRATOR | No | - | `update_company` | No |
| Visit | Visit Monitoring | AdminVisitMonitoring | `/admin/visit-monitoring` | `views/Admin/Visit/VisitMonitoringView.vue` | AdminLayout | SUPER_ADMIN, ADMINISTRATOR | Yes | `menu_visit_monitoring` | `view_visit_monitoring` | Yes |
| Reports | Reports | AdminReports | `/admin/reports` | `views/Admin/SimulationPlaceholderView.vue` | AdminLayout | SUPER_ADMIN, ADMINISTRATOR | Yes | `menu_reports` | `view_reports` | Yes, but placeholder |
| Sales | Sales Dashboard | SalesDashboard | `/sales/dashboard` | `views/Sales/Dashboard/SalesDashboardView.vue` | SalesLayout | SALES_EXECUTIVE | Yes | `menu_sales_dashboard` | `view_sales_dashboard` | Yes |
| Sales Prospect | My Prospects | SalesMyProspects | `/sales/my-prospects` | `views/Sales/Prospect/MyProspectsView.vue` | SalesLayout | SALES_EXECUTIVE | Yes | `menu_my_prospects` | `view_my_prospects` | Yes |
| Sales Prospect | Prospect Detail | SalesProspectDetail | `/sales/my-prospects/:id` | `views/Sales/Prospect/ProspectDetailView.vue` | SalesLayout | SALES_EXECUTIVE | No | - | `view_my_prospect_detail` | No |
| Sales Visit | Prospect Check In | SalesProspectCheckIn | `/sales/my-prospects/:id/check-in` | `views/Sales/Visit/CheckInView.vue` | SalesLayout | SALES_EXECUTIVE | No | - | `check_in_prospect` | No |
| Sales Visit | Prospect Check In Success | SalesProspectCheckInSuccess | `/sales/my-prospects/:id/check-in/success` | `views/Sales/Visit/CheckInSuccessView.vue` | SalesLayout | SALES_EXECUTIVE | No | - | `view_visit_result` | No |
| Sales Visit | Prospect Visit Result | SalesProspectVisitResult | `/sales/my-prospects/:id/visit-result` | `views/Sales/Visit/VisitResultView.vue` | SalesLayout | SALES_EXECUTIVE | No | - | `update_visit_result` | No |
| Sales Visit | Prospect Check Out | SalesProspectCheckOut | `/sales/my-prospects/:id/check-out` | `views/Sales/Visit/CheckOutView.vue` | SalesLayout | SALES_EXECUTIVE | No | - | `check_out_prospect` | No |
| Sales Visit | Prospect Check Out Success | SalesProspectCheckOutSuccess | `/sales/my-prospects/:id/check-out/success` | `views/Sales/Visit/CheckOutSuccessView.vue` | SalesLayout | SALES_EXECUTIVE | No | - | `view_visit_result` | No |
| Sales Prospect | Sales Pipeline | SalesPipeline | `/sales/pipeline` | `views/Sales/Prospect/SalesPipelineView.vue` | SalesLayout | SALES_EXECUTIVE | No | `menu_sales_pipeline` | `view_sales_pipeline` | Yes |
| Sales Customer | My Customers | SalesMyCustomers | `/sales/my-customers` | `views/Sales/Customer/MyCustomersView.vue` | SalesLayout | SALES_EXECUTIVE | Yes | `menu_my_customers` | `view_my_customers` | Yes |
| Sales Customer | Customer Detail | SalesCustomerDetail | `/sales/my-customers/:id` | `views/Sales/Customer/CustomerDetailView.vue` | SalesLayout | SALES_EXECUTIVE | No | - | `view_my_customer_detail` | No |
| Sales Visit | Customer Check In | SalesCustomerCheckIn | `/sales/my-customers/:id/check-in` | `views/Sales/Visit/CheckInView.vue` | SalesLayout | SALES_EXECUTIVE | No | - | `check_in_customer` | No |
| Sales Visit | Customer Check In Success | SalesCustomerCheckInSuccess | `/sales/my-customers/:id/check-in/success` | `views/Sales/Visit/CheckInSuccessView.vue` | SalesLayout | SALES_EXECUTIVE | No | - | `view_visit_result` | No |
| Sales Visit | Customer Visit Result | SalesCustomerVisitResult | `/sales/my-customers/:id/visit-result` | `views/Sales/Visit/VisitResultView.vue` | SalesLayout | SALES_EXECUTIVE | No | - | `update_visit_result` | No |
| Sales Visit | Customer Check Out | SalesCustomerCheckOut | `/sales/my-customers/:id/check-out` | `views/Sales/Visit/CheckOutView.vue` | SalesLayout | SALES_EXECUTIVE | No | - | `check_out_customer` | No |
| Sales Visit | Customer Check Out Success | SalesCustomerCheckOutSuccess | `/sales/my-customers/:id/check-out/success` | `views/Sales/Visit/CheckOutSuccessView.vue` | SalesLayout | SALES_EXECUTIVE | No | - | `view_visit_result` | No |
| Sales | History | SalesHistory | `/sales/history` | `views/Sales/HistoryView.vue` | SalesLayout | SALES_EXECUTIVE | Yes | `menu_sales_history` | `view_sales_history` | Yes |
| Profile | Sales Profile | SalesProfile | `/sales/profile` | `views/Sales/ProfileView.vue` | SalesLayout | SALES_EXECUTIVE | Yes | `menu_profile` | `view_own_profile` | Yes |
| System | Not Found | NotFound | `/:pathMatch(.*)*` | `views/NotFoundView.vue` | None | Public | No | - | - | No |

## 2. Current Sidebar Inventory

### Admin / Super Admin

All admin sidebar entries are always rendered inside `AdminLayout.vue`; visibility comes from the `/admin` route guard, not per-item conditions.

| Label | Route | Icon | Current visibility | Proposed key |
|---|---|---|---|---|
| Dashboard | `/admin/dashboard` | `pi pi-home` | AdminLayout visible to SUPER_ADMIN/ADMINISTRATOR | `menu_admin_dashboard` |
| Customer | `/admin/customers` | `pi pi-users` | Same | `menu_customers` |
| Prospect Finder | `/admin/prospect-finder` | `pi pi-compass` | Same | `menu_prospect_finder` |
| Prospect List | `/admin/prospects/list` | `pi pi-list` | Same | `menu_prospect_list` |
| Prospect Pipeline | `/admin/prospects/pipeline` | `pi pi-th-large` | Same | `menu_prospect_pipeline` |
| Visit Monitoring | `/admin/visit-monitoring` | `pi pi-map-marker` | Same | `menu_visit_monitoring` |
| Accounts | `/admin/accounts` | `pi pi-user-edit` | Same | `menu_accounts` |
| Role Management | `/admin/role-management` | `pi pi-id-card` | Same | `menu_roles` |
| Sales Structure | `/admin/sales-structure` | `pi pi-sitemap` | Same | `menu_sales_structure` |
| Reports | `/admin/reports` | `pi pi-chart-bar` | Same | `menu_reports` |

Rendered names requested but not present as separate menu items: Customer Assignment, Prospect Assignment.

### Sales

Sales sidebar and mobile bottom nav render the same five entries. `SalesPipeline` route exists but is not currently in the Sales sidebar.

| Label | Route | Icon | Current visibility | Proposed key |
|---|---|---|---|---|
| Home | `/sales/dashboard` | `pi pi-home` | SalesLayout visible to SALES_EXECUTIVE | `menu_sales_dashboard` |
| Customer | `/sales/my-customers` | `pi pi-users` | Same | `menu_my_customers` |
| Prospect | `/sales/my-prospects` | `pi pi-briefcase` | Same | `menu_my_prospects` |
| History | `/sales/history` | `pi pi-history` | Same | `menu_sales_history` |
| Profile | `/sales/profile` | `pi pi-user` | Same | `menu_profile` |

### Sales Manager

No dedicated Sales Manager layout/menu exists. `homeFor('SALES_MANAGER')` returns `/forbidden`.

## 3. Current API / Action Inventory

### Auth

- `POST /auth/login`: login.
- `POST /auth/refresh`: refresh session.
- `POST /auth/logout`: logout.
- `GET /auth/me`: view own session.
- `POST /auth/logout-all`: revoke own sessions.
- `POST /auth/change-password`: change own password.

### Dashboards

- `GET /dashboard/admin`: view admin dashboard surface.
- `GET /dashboard/sales`: view sales dashboard surface.

### Accounts

- `GET /admin/users`: view account list.
- `GET /admin/users/:id`: view account detail.
- `POST /admin/users`: create account.
- `PATCH /admin/users/:id`: update account.
- `PATCH /admin/users/:id/status`: activate/deactivate account.
- `POST /admin/users/:id/reset-password`: reset account password.
- `GET /admin/users/options/managers`: view manager options.
- Delete account is not implemented.

### Role Management

- `GET /admin/sales-roles`: view roles.
- `GET /admin/sales-roles/:id`: view role detail API exists, but frontend uses a list/dialog rather than detail route.
- `POST /admin/sales-roles`: create role.
- `PATCH /admin/sales-roles/:id`: update role.
- `PATCH /admin/sales-roles/:id/status`: activate/deactivate role.
- `DELETE /admin/sales-roles/:id`: delete role.
- Permission mapping is not implemented.

### Sales Structure

- `GET /admin/sales-structure`: view structure.
- `POST /admin/sales-structure/assignments`: create assignment.
- `POST /admin/sales-structure/assignments/:id/move`: move assignment API exists.
- `GET /admin/sales-structure/users/:userId/history`: view assignment history API exists.
- Frontend currently exposes create assignment; move and history buttons are disabled and API wrappers are absent.

### Prospect Management

- Admin: pipeline, won queue, sales executive options, finder search/detail, save prospect, delete prospect, review/detail, comments, attachments, mention users, place details, visit list, prospect visits, approve/reject deletion, conversion form, convert.
- Sales: my prospects, prospect detail, transition/decision, check-in, check-out, comments, attachments, mention users, place details, request deletion.
- Implemented action candidates: view prospect list/detail/pipeline/finder, create prospect from finder, assign prospect on save, update pipeline/status, check in/out, comment, attach/download comment files, delete prospect, request/approve/reject deletion, convert prospect.

### Customer / Company Management

- Admin: customer list, customer paginated list, filter options, customer detail, customer place details, delete customer, parent-company search/detail/update, conversion from prospect.
- Sales: my customers, customer detail, customer place details.
- Implemented action candidates: view customer list/detail, convert prospect to customer, update parent company, delete customer, view place details.
- Admin customer add/edit views exist, but backend create/update customer-site endpoints are not registered apart from conversion and parent-company update; mark direct create/update customer as partial.

### Visit Management

- Admin: `GET /admin/visits`, `GET /admin/prospects/:prospectId/visits`, `POST /admin/visits/:visitId/delete`.
- Sales: `GET /sales/visits`, check-in/check-out through prospect/customer-backed prospect endpoints, `POST /sales/visits/:visitId/delete`.
- View evidence is implemented through selfie/static upload references and frontend visit modals/components.
- Approve outside-radius is not implemented.

### Reports

- Admin Reports route exists as a placeholder only. No report API endpoints are registered.

## 4. Proposed Permission Tree

Use only current routes/actions. Recommended default roles are initial suggestions, not implementation rules.

| Label | Key | Type | Related route/API | Status | Defaults |
|---|---|---|---|---|---|
| Admin Dashboard | `menu_admin_dashboard` | MENU | `/admin/dashboard` | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR |
| View Admin Dashboard | `view_admin_dashboard` | ACTION | `GET /dashboard/admin` | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR |
| Sales Dashboard | `menu_sales_dashboard` | MENU | `/sales/dashboard` | IMPLEMENTED | SALES_EXECUTIVE |
| View Sales Dashboard | `view_sales_dashboard` | ACTION | `GET /dashboard/sales` | IMPLEMENTED | SALES_EXECUTIVE |
| Accounts | `menu_accounts` | MENU | `/admin/accounts` | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR |
| View Accounts | `view_accounts` | ACTION | `GET /admin/users` | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR |
| View Account Detail | `view_account_detail` | ACTION | `GET /admin/users/:id` | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR |
| Create Account | `create_account` | ACTION | `POST /admin/users` | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR |
| Update Account | `update_account` | ACTION | `PATCH /admin/users/:id` | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR |
| Update Account Status | `update_account_status` | ACTION | `PATCH /admin/users/:id/status` | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR |
| Reset Account Password | `reset_account_password` | ACTION | `POST /admin/users/:id/reset-password` | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR |
| Roles | `menu_roles` | MENU | `/admin/role-management` | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR |
| View Roles | `view_roles` | ACTION | `GET /admin/sales-roles` | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR |
| View Role Detail | `view_role_detail` | ACTION | `GET /admin/sales-roles/:id` | PARTIAL | SUPER_ADMIN, ADMINISTRATOR |
| Create Role | `create_role` | ACTION | `POST /admin/sales-roles` | IMPLEMENTED | SUPER_ADMIN |
| Update Role | `update_role` | ACTION | `PATCH /admin/sales-roles/:id` | IMPLEMENTED | SUPER_ADMIN |
| Update Role Status | `update_role_status` | ACTION | `PATCH /admin/sales-roles/:id/status` | IMPLEMENTED | SUPER_ADMIN |
| Delete Role | `delete_role` | ACTION | `DELETE /admin/sales-roles/:id` | IMPLEMENTED | SUPER_ADMIN |
| Manage Role Permissions | `manage_role_permissions` | ACTION | Not present | PLANNED | SUPER_ADMIN |
| Sales Structure | `menu_sales_structure` | MENU | `/admin/sales-structure` | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR |
| View Sales Structure | `view_sales_structure` | ACTION | `GET /admin/sales-structure` | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR |
| Create Assignment | `create_sales_assignment` | ACTION | `POST /admin/sales-structure/assignments` | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR |
| Move Assignment | `move_sales_assignment` | ACTION | `POST /admin/sales-structure/assignments/:id/move` | PARTIAL | SUPER_ADMIN, ADMINISTRATOR |
| View Assignment History | `view_sales_assignment_history` | ACTION | `GET /admin/sales-structure/users/:userId/history` | PARTIAL | SUPER_ADMIN, ADMINISTRATOR |
| Prospect Finder | `menu_prospect_finder` | MENU | `/admin/prospect-finder` | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR |
| Prospect List | `menu_prospect_list` | MENU | `/admin/prospects/list` | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR |
| Prospect Pipeline | `menu_prospect_pipeline` | MENU | `/admin/prospects/pipeline` | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR |
| My Prospects | `menu_my_prospects` | MENU | `/sales/my-prospects` | IMPLEMENTED | SALES_EXECUTIVE |
| Sales Pipeline | `menu_sales_pipeline` | MENU | `/sales/pipeline` | PARTIAL | SALES_EXECUTIVE |
| View Prospect Finder | `view_prospect_finder` | ACTION | `GET /admin/prospect-finder/search` | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR |
| View Prospect List | `view_prospect_list` | ACTION | `/admin/prospects/list`, pipeline APIs | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR |
| View Prospect Detail | `view_prospect_detail` | ACTION | `GET /admin/prospects/:id` | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR |
| View My Prospects | `view_my_prospects` | ACTION | `GET /sales/prospects` | IMPLEMENTED | SALES_EXECUTIVE |
| View My Prospect Detail | `view_my_prospect_detail` | ACTION | `GET /sales/prospects/:id` | IMPLEMENTED | SALES_EXECUTIVE |
| Create Prospect | `create_prospect` | ACTION | `POST /admin/prospects` | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR |
| Update Prospect Pipeline | `update_prospect_pipeline` | ACTION | `PATCH /sales/prospects/:id/transition` | IMPLEMENTED | SALES_EXECUTIVE |
| Delete Prospect | `delete_prospect` | ACTION | `DELETE /admin/prospects/:id` | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR |
| Request Prospect Deletion | `request_prospect_deletion` | ACTION | `POST /sales/prospects/:id/request-deletion` | IMPLEMENTED | SALES_EXECUTIVE |
| Approve Prospect Deletion | `approve_prospect_deletion` | ACTION | `POST /admin/prospects/:id/approve-deletion` | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR |
| Reject Prospect Deletion | `reject_prospect_deletion` | ACTION | `POST /admin/prospects/:id/reject-deletion` | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR |
| Convert Prospect | `convert_prospect` | ACTION | `POST /admin/prospects/:id/convert` | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR |
| Comments | `manage_prospect_comments` | ACTION | comments/attachments APIs | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR, SALES_EXECUTIVE |
| Customers | `menu_customers` | MENU | `/admin/customers` | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR |
| My Customers | `menu_my_customers` | MENU | `/sales/my-customers` | IMPLEMENTED | SALES_EXECUTIVE |
| View Customers | `view_customers` | ACTION | `GET /admin/customers/list` | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR |
| View Customer Detail | `view_customer_detail` | ACTION | `GET /admin/customers/:id` | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR |
| View My Customers | `view_my_customers` | ACTION | `GET /sales/customers` | IMPLEMENTED | SALES_EXECUTIVE |
| View My Customer Detail | `view_my_customer_detail` | ACTION | `GET /sales/customers/:id` | IMPLEMENTED | SALES_EXECUTIVE |
| Create Customer | `create_customer` | ACTION | Direct admin customer route only | PARTIAL | SUPER_ADMIN, ADMINISTRATOR |
| Update Customer | `update_customer` | ACTION | Direct admin customer route only | PARTIAL | SUPER_ADMIN, ADMINISTRATOR |
| Delete Customer | `delete_customer` | ACTION | `DELETE /admin/customers/:id` | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR |
| View Company Detail | `view_company_detail` | ACTION | `GET /admin/companies/:id` | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR |
| Update Company | `update_company` | ACTION | `PATCH /admin/companies/:id` | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR |
| Visit Monitoring | `menu_visit_monitoring` | MENU | `/admin/visit-monitoring` | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR |
| View Visit Monitoring | `view_visit_monitoring` | ACTION | `GET /admin/visits` | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR |
| View Own Visits | `view_own_visits` | ACTION | `GET /sales/visits` | IMPLEMENTED | SALES_EXECUTIVE |
| Check In Prospect | `check_in_prospect` | ACTION | `POST /sales/prospects/:id/visits/check-in` | IMPLEMENTED | SALES_EXECUTIVE |
| Check Out Prospect | `check_out_prospect` | ACTION | `PATCH /sales/prospects/:id/visits/:visitId/check-out` | IMPLEMENTED | SALES_EXECUTIVE |
| Check In Customer | `check_in_customer` | ACTION | Same handler via customer route | IMPLEMENTED | SALES_EXECUTIVE |
| Check Out Customer | `check_out_customer` | ACTION | Same handler via customer route | IMPLEMENTED | SALES_EXECUTIVE |
| Update Visit Result | `update_visit_result` | ACTION | Visit result/check-out flow | IMPLEMENTED | SALES_EXECUTIVE |
| View Visit Evidence | `view_visit_evidence` | ACTION | uploads/selfie references and visit modals | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR |
| Delete Visit | `delete_visit` | ACTION | `POST /admin/visits/:visitId/delete`, `POST /sales/visits/:visitId/delete` | IMPLEMENTED | SUPER_ADMIN, ADMINISTRATOR, SALES_EXECUTIVE |
| Reports | `menu_reports` | MENU | `/admin/reports` | PARTIAL | SUPER_ADMIN, ADMINISTRATOR |
| View Reports | `view_reports` | ACTION | No API; placeholder route | PARTIAL | SUPER_ADMIN, ADMINISTRATOR |
| Sales History | `menu_sales_history` | MENU | `/sales/history` | IMPLEMENTED | SALES_EXECUTIVE |
| View Sales History | `view_sales_history` | ACTION | `GET /sales/visits` and view route | IMPLEMENTED | SALES_EXECUTIVE |
| Profile | `menu_profile` | MENU | `/sales/profile` | IMPLEMENTED | SALES_EXECUTIVE |
| View Own Profile | `view_own_profile` | ACTION | Auth session/profile view | IMPLEMENTED | SALES_EXECUTIVE |
| Change Own Password | `change_own_password` | ACTION | `POST /auth/change-password` | IMPLEMENTED | All authenticated |

Proposed count: 62 permission keys (menus and actions), excluding public/system routes.

## 5. Proposed Permission Key Rules

- Use lowercase snake_case only.
- Prefix menu visibility with `menu_`.
- Prefix read page/API access with `view_`.
- Use verbs for mutations: `create_`, `update_`, `delete_`, `reset_`, `convert_`, `check_in_`, `check_out_`, `approve_`, `reject_`, `request_`.
- Menu permissions do not imply action permissions.
- Page permissions do not imply mutation permissions.
- Keep system role, permission, and hierarchy scope separate.

## 6. Valid Landing Pages

Landing-page validation rule: selected landing page must also be permitted by selected permissions.

| Label | Path | Required permission | Roles/layout | Safe default |
|---|---|---|---|---|
| Admin Dashboard | `/admin/dashboard` | `view_admin_dashboard` | SUPER_ADMIN/ADMINISTRATOR, AdminLayout | Yes |
| Account List | `/admin/accounts` | `view_accounts` | SUPER_ADMIN/ADMINISTRATOR, AdminLayout | Yes |
| Role Management | `/admin/role-management` | `view_roles` | SUPER_ADMIN/ADMINISTRATOR, AdminLayout | Yes |
| Sales Structure | `/admin/sales-structure` | `view_sales_structure` | SUPER_ADMIN/ADMINISTRATOR, AdminLayout | Yes |
| Prospect Finder | `/admin/prospect-finder` | `view_prospect_finder` | SUPER_ADMIN/ADMINISTRATOR, AdminLayout | Yes |
| Prospect List | `/admin/prospects/list` | `view_prospect_list` | SUPER_ADMIN/ADMINISTRATOR, AdminLayout | Yes |
| Prospect Pipeline | `/admin/prospects/pipeline` | `view_prospect_pipeline` | SUPER_ADMIN/ADMINISTRATOR, AdminLayout | Yes |
| Customer List | `/admin/customers` | `view_customers` | SUPER_ADMIN/ADMINISTRATOR, AdminLayout | Yes |
| Visit Monitoring | `/admin/visit-monitoring` | `view_visit_monitoring` | SUPER_ADMIN/ADMINISTRATOR, AdminLayout | Yes |
| Reports | `/admin/reports` | `view_reports` | SUPER_ADMIN/ADMINISTRATOR, AdminLayout | Caution: placeholder |
| Sales Dashboard | `/sales/dashboard` | `view_sales_dashboard` | SALES_EXECUTIVE, SalesLayout | Yes |
| My Prospects | `/sales/my-prospects` | `view_my_prospects` | SALES_EXECUTIVE, SalesLayout | Yes |
| Sales Pipeline | `/sales/pipeline` | `view_sales_pipeline` | SALES_EXECUTIVE, SalesLayout | Yes, but no sidebar item |
| My Customers | `/sales/my-customers` | `view_my_customers` | SALES_EXECUTIVE, SalesLayout | Yes |
| Sales History | `/sales/history` | `view_sales_history` | SALES_EXECUTIVE, SalesLayout | Yes |
| Profile | `/sales/profile` | `view_own_profile` | SALES_EXECUTIVE, SalesLayout | Yes |

## 7. Department Status

Department is not implemented as a database table, user field, role field, master data, or hardcoded option in the scoped files. Safest recommendation: A. defer Department until the permission model is stable. If product requires it later, prefer a small master table over frontend-only production data.

## 8. Current Authorization Model

Frontend:

- Route meta uses `role?: UserRole`.
- `/admin` parent route requires `ADMINISTRATOR`, with `roleAllowed()` permitting `SUPER_ADMIN`.
- `/sales` parent route requires `SALES_EXECUTIVE`.
- Sidebar entries are unconditional inside their layout.
- Layout selection is route-based, not permission-based.
- `SALES_MANAGER` has no layout and is redirected to `/forbidden`.

Backend:

- `Authenticate` validates access token and sets principal.
- `RequirePasswordChanged` blocks users with mandatory password changes.
- `RequireRole(...)` protects route groups: admin group uses SUPER_ADMIN/ADMINISTRATOR; sales group uses SALES_EXECUTIVE.
- Admin services repeat admin-role checks with `actor.Role.IsAdminRole()`.
- Sales/prospect/customer services also enforce actor/data ownership rules internally.

Later changes:

- Keep system role for broad app surface and layout family.
- Add permissions for menu visibility, route/page access, and action/API access.
- Keep hierarchy level/data ownership separate for whose records are visible.
- Keep sensitive auth operations such as authentication, refresh, password-change-required, and primary system role invariants system-controlled.

## 9. Role Page Current State

Current Role Management page supports:

- Role list: yes.
- Role detail: API yes, separate frontend route no.
- Create role: yes.
- Edit role: yes.
- Delete role: yes, with frontend default-role disable and backend in-use/default protection.
- Status toggle: yes.
- Description: yes.
- Numeric hierarchy level: yes.

Not implemented:

- Department.
- Initial open menu / landing page.
- Permission mapping.
- Copy/duplicate role.
- Selected permission count.
- Permission tree/search/select visible/clear visible/clear all/select all.
- Backend permission key badges.

Likely implementation files:

- `frontend/src/views/Admin/Roles/AdminRoleManagementView.vue`
- `frontend/src/types/admin.ts`
- `frontend/src/api/admin.ts`
- `frontend/src/stores/admin.ts`
- `frontend/src/router/index.ts`
- `frontend/src/layouts/AdminLayout.vue`
- `frontend/src/layouts/SalesLayout.vue`
- `backend/server/app.go`
- `backend/internal/admin/model/sales_organization.go`
- `backend/internal/admin/handler/sales_organization.go`
- `backend/internal/admin/service/sales_organization.go`
- New backend permission model/repository/service files and migrations, if approved in a future phase.

## 10. Missing Implementation And Recommended Phasing

Missing or unclear:

- Permission storage schema and assignment model.
- Whether permissions attach to sales organizational roles, system roles, or a new custom role concept.
- Department semantics and source of truth.
- Sales Manager layout/menu expectations.
- Whether placeholder Reports should be selectable as a landing page.
- Whether direct Customer Add/Edit pages should remain reachable without matching backend create/update endpoints.
- Whether Sales Pipeline should appear in Sales sidebar.
- Whether move/history frontend should be enabled now that backend endpoints exist.
- Exact split between "view all", "view team", and "view own" for prospects/customers/visits/reports.

Recommended phased implementation:

1. Add permission catalog constants and backend read endpoint for the permission tree.
2. Add role permission persistence and expose it in role create/edit/detail.
3. Add landing page persistence and validate it against selected permissions.
4. Add frontend Role Permission Explorer UI using the catalog.
5. Enforce route/menu visibility on frontend.
6. Enforce backend action permissions at route/service boundary.
7. Add hierarchy-scope rules separately from permissions.
8. Revisit Department only after the permission role model is stable.
