# Role Access Matrix

> **Document:** 02-role-access-matrix.md  
> **Project:** CRM Prospect Simulator  
> **Date:** 2026-07-30  
> **Status:** Draft — Design Phase  

---

## 1. Role Definitions

### 1.1 ADMINISTRATOR

| Property | Value |
|---|---|
| Technical Code | `ADMINISTRATOR` |
| Display Name | Administrator |
| Level | Top |

**Responsibilities:**

- Manage user accounts (create, update, activate, deactivate)
- Create Sales Manager and Sales Executive accounts
- Assign manager for Sales Executive
- Reset passwords and terminate active sessions
- Manage Prospect Finder (search, create, assign prospects)
- View all prospects, customers, visits, pipeline data
- Review Won prospects and convert to customers
- View user activity and reports
- No operational restrictions (ALL scope)

**Cannot:**

- Perform Check In / Check Out (operational field activity)

### 1.2 SALES_MANAGER

| Property | Value |
|---|---|
| Technical Code | `SALES_MANAGER` |
| Display Name | Sales Manager |
| Level | Middle — Supervisory |

**Responsibilities:**

- View team dashboard
- View Sales Executives under management
- View team prospects, customers, pipeline, visits
- View team performance and activity
- View prospect and customer details
- Report on team data (read-only)

**Cannot:**

- Create or manage user accounts
- Check In or Check Out
- Transition pipeline status for team members
- Convert prospects to customers
- Delete customers
- Manage roles or system settings
- Access Admin pages or Admin API endpoints

### 1.3 SALES_EXECUTIVE

| Property | Value |
|---|---|
| Technical Code | `SALES_EXECUTIVE` |
| Display Name | Sales Executive |
| Level | Base — Operational |

**Responsibilities:**

- View personal dashboard
- View assigned prospects and owned customers
- Perform Check In and Check Out
- Update own pipeline status
- View own visit history
- View profile and personal statistics
- Request prospect deletion (existing flow)

**Cannot:**

- View other Sales Executive data
- Access Admin or Manager pages
- Manage accounts
- Assign prospects to other users
- Convert prospects
- View system audit logs

---

## 2. Page Access Matrix

### 2.1 Route Inventory (Current)

| Module | Page | Route | Status |
|---|---|---|---|
| Public | Login | `/login` | Existing |
| Public | Not Found | `/:pathMatch(.*)*` | Existing |
| Admin | Dashboard | `/admin/dashboard` | Existing |
| Admin | Sales Executive | `/admin/sales-executives` | Placeholder |
| Admin | Customer | `/admin/customers` | Existing |
| Admin | Customer Add | `/admin/customers/add` | Existing |
| Admin | Customer Detail | `/admin/customers/:id` | Existing |
| Admin | Customer Edit | `/admin/customers/:id/edit` | Existing |
| Admin | Customer Assignment | `/admin/customer-assignment` | Placeholder |
| Admin | Visit Monitoring | `/admin/visit-monitoring` | Existing |
| Admin | Prospect Finder | `/admin/prospect-finder` | Existing |
| Admin | Prospect List | `/admin/prospects/list` | Existing |
| Admin | Prospect Pipeline | `/admin/prospects/pipeline` | Existing |
| Admin | Prospect Review | `/admin/prospects/:id/review` | Existing |
| Admin | Prospect Convert | `/admin/prospects/:id/convert` | Existing |
| Admin | Won Prospects | `/admin/prospects/won` | Existing |
| Admin | Prospect Assignment | `/admin/prospect-assignment` | Placeholder |
| Admin | Companies | `/admin/companies` | Existing |
| Admin | Company Add | `/admin/companies/add` | Existing |
| Admin | Company Detail | `/admin/companies/:id` | Existing (under construction) |
| Admin | Company Edit | `/admin/companies/:id/edit` | Existing |
| Admin | Reports | `/admin/reports` | Placeholder |
| SE | Dashboard | `/sales/dashboard` | Existing |
| SE | My Prospects | `/sales/my-prospects` | Existing |
| SE | Prospect Detail | `/sales/my-prospects/:id` | Existing |
| SE | Check In | `/sales/my-prospects/:id/check-in` | Existing |
| SE | Check In Success | `/sales/my-prospects/:id/check-in/success` | Existing |
| SE | Visit Result | `/sales/my-prospects/:id/visit-result` | Existing |
| SE | Check Out | `/sales/my-prospects/:id/check-out` | Existing |
| SE | Check Out Success | `/sales/my-prospects/:id/check-out/success` | Existing |
| SE | Pipeline | `/sales/pipeline` | Existing |
| SE | My Customers | `/sales/my-customers` | Existing |
| SE | Customer Detail | `/sales/my-customers/:id` | Existing |
| SE | Customer Check In | `/sales/my-customers/:id/check-in` | Existing |
| SE | Customer Visit Result | `/sales/my-customers/:id/visit-result` | Existing |
| SE | Customer Check Out | `/sales/my-customers/:id/check-out` | Existing |
| SE | History | `/sales/history` | Existing |
| SE | Profile | `/sales/profile` | Existing |

### 2.2 Access Matrix (Final Proposed)

| Module | Page | Route | Admin | Sales Manager | Sales Executive | Data Scope | Page Status |
|---|---|---|---|---|---|---|---|
| Public | Login | `/login` | Yes | Yes | Yes | — | Existing |
| Public | Not Found | `/:pathMatch(.*)*` | Yes | Yes | Yes | — | Existing |
| Public | Forbidden | `/forbidden` | Yes | Yes | Yes | — | Planned |
| Admin | Dashboard | `/admin/dashboard` | **Yes** | No | No | ALL | Existing |
| Admin | Account Management | `/admin/accounts` | **Yes** | No | No | ALL | Planned |
| Admin | Account Detail | `/admin/accounts/:id` | **Yes** | No | No | ALL | Planned |
| Admin | Sales Executive | `/admin/sales-executives` | **Yes** | No | No | ALL | Placeholder |
| Admin | Customer | `/admin/customers` | **Yes** | No | No | ALL | Existing |
| Admin | Customer Add | `/admin/customers/add` | **Yes** | No | No | — | Existing |
| Admin | Customer Detail | `/admin/customers/:id` | **Yes** | No | No | ALL | Existing |
| Admin | Customer Edit | `/admin/customers/:id/edit` | **Yes** | No | No | — | Existing |
| Admin | Customer Assignment | `/admin/customer-assignment` | **Yes** | No | No | ALL | Placeholder |
| Admin | Visit Monitoring | `/admin/visit-monitoring` | **Yes** | No | No | ALL | Existing |
| Admin | Prospect Finder | `/admin/prospect-finder` | **Yes** | No | No | — | Existing |
| Admin | Prospect List | `/admin/prospects/list` | **Yes** | No | No | ALL | Existing |
| Admin | Prospect Pipeline | `/admin/prospects/pipeline` | **Yes** | No | No | ALL | Existing |
| Admin | Prospect Review | `/admin/prospects/:id/review` | **Yes** | No | No | ALL | Existing |
| Admin | Prospect Convert | `/admin/prospects/:id/convert` | **Yes** | No | No | — | Existing |
| Admin | Won Prospects | `/admin/prospects/won` | **Yes** | No | No | ALL | Existing |
| Admin | Prospect Assignment | `/admin/prospect-assignment` | **Yes** | No | No | ALL | Placeholder |
| Admin | Companies | `/admin/companies` | **Yes** | No | No | ALL | Existing |
| Admin | Company Add | `/admin/companies/add` | **Yes** | No | No | — | Existing |
| Admin | Company Detail | `/admin/companies/:id` | **Yes** | No | No | ALL | Existing |
| Admin | Company Edit | `/admin/companies/:id/edit` | **Yes** | No | No | — | Existing |
| Admin | Reports | `/admin/reports` | **Yes** | No | No | ALL | Placeholder |
| Admin | Activity Log | `/admin/activity-log` | **Yes** | No | No | ALL | Planned |
| Manager | Dashboard | `/manager/dashboard` | No | **Yes** | No | TEAM | Planned |
| Manager | Team Members | `/manager/team-members` | No | **Yes** | No | TEAM | Planned |
| Manager | Team Prospects | `/manager/prospects` | No | **Yes** | No | TEAM | Planned |
| Manager | Team Customers | `/manager/customers` | No | **Yes** | No | TEAM | Planned |
| Manager | Team Pipeline | `/manager/pipeline` | No | **Yes** | No | TEAM | Planned |
| Manager | Team Visits | `/manager/visits` | No | **Yes** | No | TEAM | Planned |
| Manager | Team Reports | `/manager/reports` | No | **Yes** | No | TEAM | Planned |
| Manager | Profile | `/manager/profile` | No | **Yes** | Yes | OWN | Shared (planned) |
| SE | Dashboard | `/sales/dashboard` | No | No | **Yes** | OWN | Existing |
| SE | My Prospects | `/sales/my-prospects` | No | No | **Yes** | OWN | Existing |
| SE | Prospect Detail | `/sales/my-prospects/:id` | No | No | **Yes** | OWN | Existing |
| SE | Pipeline | `/sales/pipeline` | No | No | **Yes** | OWN | Existing |
| SE | My Customers | `/sales/my-customers` | No | No | **Yes** | OWN | Existing |
| SE | Customer Detail | `/sales/my-customers/:id` | No | No | **Yes** | OWN | Existing |
| SE | Check In | `/sales/*/check-in` | No | No | **Yes** | OWN | Existing |
| SE | Check In Success | `/sales/*/check-in/success` | No | No | **Yes** | OWN | Existing |
| SE | Visit Result | `/sales/*/visit-result` | No | No | **Yes** | OWN | Existing |
| SE | Check Out | `/sales/*/check-out` | No | No | **Yes** | OWN | Existing |
| SE | Check Out Success | `/sales/*/check-out/success` | No | No | **Yes** | OWN | Existing |
| SE | History | `/sales/history` | No | No | **Yes** | OWN | Existing |
| SE | Profile | `/sales/profile` | No | **Yes** | **Yes** | OWN | Shared (planned) |

---

## 3. Action Access Matrix

| Action | Administrator | Sales Manager | Sales Executive | Scope | Notes |
|---|---|---|---|---|---|
| Create Prospect | **Yes** | No | No | ALL | Via Prospect Finder (Admin only) |
| View Prospect | **Yes** | **Yes** (team) | **Yes** (own) | ALL / TEAM / OWN | Scope varies by role |
| Update Prospect | **Yes** | No | **Yes** (own) | ALL / OWN | Manager is read-only |
| Assign Prospect | **Yes** | No | No | ALL | Admin assigns at creation |
| Reassign Prospect | **Yes** | Planned | No | ALL | Future bulk reassignment |
| Transition Pipeline | **Yes** | No | **Yes** (own) | ALL / OWN | Manager cannot transition |
| Mark Won | **Yes** | No | **Yes** (own) | ALL / OWN | Via transition |
| Mark Lost | **Yes** | No | **Yes** (own) | ALL / OWN | Via transition |
| Request Deletion | No | No | **Yes** (own) | OWN | Existing flow |
| Approve Deletion | **Yes** | No | No | ALL | Admin only |
| Reject Deletion | **Yes** | No | No | ALL | Admin only |
| Review Won | **Yes** | **Yes** (view only) | No | ALL / TEAM | Manager read-only |
| Convert Customer | **Yes** | No | No | — | Admin only |
| View Customer | **Yes** | **Yes** (team) | **Yes** (own) | ALL / TEAM / OWN | Scope varies |
| Update Customer | **Yes** | No | No | ALL | Admin only |
| Delete Customer | **Yes** | No | No | ALL | Admin only |
| Check In | No | No | **Yes** (own) | OWN | Operational field activity |
| Check Out | No | No | **Yes** (own) | OWN | Operational field activity |
| Add Comment | **Yes** | **Yes** (team) | **Yes** (own) | ALL / TEAM / OWN | Scope varies |
| View Visit | **Yes** | **Yes** (team) | **Yes** (own) | ALL / TEAM / OWN | Scope varies |
| Delete Visit | **Yes** | No | No | ALL | Admin only |
| Create Account | **Yes** | No | No | — | Admin only |
| Update Account | **Yes** | No | No | — | Admin only |
| Activate Account | **Yes** | No | No | — | Admin only |
| Deactivate Account | **Yes** | No | No | — | Admin only |
| Reset Password | **Yes** | No | No | — | Admin only |
| End Session | **Yes** | No | No | — | Admin only |
| View Activity Log | **Yes** | No | No | ALL | Admin only |

---

## 4. API Access Matrix

| Method | Path | Current Guard | Proposed Role(s) | Scope | Change Required | Notes |
|---|---|---|---|---|---|---|
| POST | `/api/v1/auth/login` | Public | Public | — | No | |
| POST | `/api/v1/auth/refresh` | Public | Public | — | No | |
| POST | `/api/v1/auth/logout` | Cookie | Public | — | No | Cookie-based |
| GET | `/api/v1/auth/me` | Auth | Auth (all roles) | OWN | No | |
| POST | `/api/v1/auth/logout-all` | Auth | Auth (all roles) | OWN | No | |
| GET | `/api/v1/sales/prospects` | SE | SE | OWN | No | |
| GET | `/api/v1/sales/prospects/:id` | SE | SE | OWN | No | |
| PATCH | `/api/v1/sales/prospects/:id/transition` | SE | SE | OWN | No | |
| POST | `/api/v1/sales/prospects/:id/visits/check-in` | SE | SE | OWN | No | |
| PATCH | `/api/v1/sales/prospects/:id/visits/:visitId/check-out` | SE | SE | OWN | No | |
| GET | `/api/v1/sales/prospects/:id/comments` | SE | SE | OWN | No | |
| POST | `/api/v1/sales/prospects/:id/comments` | SE | SE | OWN | No | |
| POST | `/api/v1/sales/prospects/:id/request-deletion` | SE | SE | OWN | No | |
| GET | `/api/v1/sales/visits` | SE | SE | OWN | No | |
| GET | `/api/v1/sales/customers` | SE | SE | OWN | No | |
| GET | `/api/v1/sales/customers/:id` | SE | SE | OWN | No | |
| GET | `/api/v1/sales/customers/:id/place-details` | SE | SE | OWN | No | |
| GET | `/api/v1/sales/prospects/:id/place-details` | SE | SE | OWN | No | |
| GET | `/api/v1/sales/pipeline` | SE | SE | OWN | No | |
| GET | `/api/v1/admin/dashboard` | Admin | Admin | ALL | No | |
| GET | `/api/v1/admin/prospects/pipeline` | Admin | Admin | ALL | No | |
| GET | `/api/v1/admin/prospects/won` | Admin | Admin | ALL | No | |
| GET | `/api/v1/admin/prospects/:id` | Admin | Admin, SM (separate) | ALL / TEAM | Yes | SM needs team-scoped endpoint |
| GET | `/api/v1/admin/prospects/:id/comments` | Admin | Admin, SM (separate) | ALL / TEAM | Yes | |
| POST | `/api/v1/admin/prospects/:id/comments` | Admin | Admin, SM (separate) | ALL / TEAM | Yes | |
| GET | `/api/v1/admin/prospects/:id/place-details` | Admin | Admin, SM (separate) | ALL / TEAM | Yes | |
| DELETE | `/api/v1/admin/prospects/:id` | Admin | Admin | ALL | No | |
| POST | `/api/v1/admin/prospects` | Admin | Admin | ALL | No | |
| GET | `/api/v1/admin/sales-executives` | Admin | Admin | ALL | No | |
| GET | `/api/v1/admin/prospect-finder/search` | Admin | Admin | — | No | |
| GET | `/api/v1/admin/prospect-finder/place-details/:googlePlaceId` | Admin | Admin | — | No | |
| POST | `/api/v1/admin/prospects/:id/approve-deletion` | Admin | Admin | ALL | No | |
| POST | `/api/v1/admin/prospects/:id/reject-deletion` | Admin | Admin | ALL | No | |
| GET | `/api/v1/admin/prospects/:id/conversion-form` | Admin | Admin | — | No | |
| POST | `/api/v1/admin/prospects/:id/convert` | Admin | Admin | — | No | |
| GET | `/api/v1/admin/customers` | Admin | Admin | ALL | No | |
| GET | `/api/v1/admin/customers/list` | Admin | Admin | ALL | No | |
| GET | `/api/v1/admin/customers/filter-options` | Admin | Admin | ALL | No | |
| GET | `/api/v1/admin/customers/:id` | Admin | Admin | ALL | No | |
| DELETE | `/api/v1/admin/customers/:id` | Admin | Admin | ALL | No | |
| GET | `/api/v1/admin/visits` | Admin | Admin | ALL | No | |
| POST | `/api/v1/admin/visits/:visitId/delete` | Admin | Admin | ALL | No | |
| GET | `/api/v1/admin/parent-companies` | Admin | Admin | ALL | No | |
| GET | `/api/v1/admin/companies/:id` | Admin | Admin | ALL | No | |
| PATCH | `/api/v1/admin/companies/:id` | Admin | Admin | ALL | No | |
| — | `/api/v1/manager/*` | — | **SM** | TEAM | **New group** | To be created |

### 4.1 Proposed Manager Endpoints

| Method | Path | Scope | Notes |
|---|---|---|---|
| GET | `/api/v1/manager/dashboard` | TEAM | Team summary |
| GET | `/api/v1/manager/team-members` | TEAM | Sales Executives under this manager |
| GET | `/api/v1/manager/prospects` | TEAM | Team prospects list |
| GET | `/api/v1/manager/prospects/:id` | TEAM | Team prospect detail |
| GET | `/api/v1/manager/prospects/:id/comments` | TEAM | Team prospect comments |
| GET | `/api/v1/manager/prospects/:id/place-details` | TEAM | Team prospect place details |
| GET | `/api/v1/manager/customers` | TEAM | Team customers list |
| GET | `/api/v1/manager/customers/:id` | TEAM | Team customer detail |
| GET | `/api/v1/manager/pipeline` | TEAM | Team pipeline |
| GET | `/api/v1/manager/visits` | TEAM | Team visits |

---

## 5. Menu Design

### 5.1 Administrator Menu

```
Dashboard          → /admin/dashboard
Accounts           → /admin/accounts           [Planned]
Sales Executives   → /admin/sales-executives    [Placeholder]
Customers          → /admin/customers
Visit Monitoring   → /admin/visit-monitoring
Prospect Finder    → /admin/prospect-finder
Prospect List      → /admin/prospects/list
Prospect Assignment→ /admin/prospect-assignment [Placeholder]
Reports            → /admin/reports             [Placeholder]
Activity Log       → /admin/activity-log        [Planned]
```

### 5.2 Sales Manager Menu

```
Dashboard          → /manager/dashboard         [Planned]
Team Members       → /manager/team-members       [Planned]
Team Prospects     → /manager/prospects          [Planned]
Team Customers     → /manager/customers          [Planned]
Team Pipeline      → /manager/pipeline           [Planned]
Team Visits        → /manager/visits             [Planned]
Reports            → /manager/reports            [Planned]
Profile            → /manager/profile            [Shared]
```

### 5.3 Sales Executive Menu

```
Home               → /sales/dashboard
Customer           → /sales/my-customers
Prospect           → /sales/my-prospects
Pipeline           → /sales/pipeline
History            → /sales/history
Profile            → /sales/profile              [Shared]
```

---

## 6. Data Scope Definitions

### 6.1 ALL

The user can access every record of a given entity regardless of ownership or team.

**Applies to:** ADMINISTRATOR

**Example queries (conceptual):**
```sql
-- All prospects
SELECT * FROM prospects;

-- All customers
SELECT * FROM customer_sites;

-- All visits
SELECT * FROM prospect_visits;
```

### 6.2 TEAM

The user can access records owned by any Sales Executive whose `manager_id` matches the current user's ID.

**Applies to:** SALES_MANAGER

**Example queries (conceptual):**
```sql
-- Team prospects
SELECT p.*
FROM prospects p
JOIN users u ON u.id = p.assigned_sales_executive_id
WHERE u.manager_id = current_user_id
  AND u.role = 'SALES_EXECUTIVE';

-- Team customers
SELECT cs.*
FROM customer_sites cs
JOIN users u ON u.id = cs.sales_executive_id
WHERE u.manager_id = current_user_id
  AND u.role = 'SALES_EXECUTIVE';

-- Team members
SELECT id, email, full_name, role, status
FROM users
WHERE manager_id = current_user_id
  AND role = 'SALES_EXECUTIVE';
```

### 6.3 OWN

The user can only access records where they are the direct owner or assignee.

**Applies to:** SALES_EXECUTIVE

**Example queries (conceptual):**
```sql
-- Own prospects
SELECT * FROM prospects WHERE assigned_sales_executive_id = current_user_id;

-- Own customers
SELECT * FROM customer_sites WHERE sales_executive_id = current_user_id;

-- Own visits
SELECT * FROM prospect_visits WHERE sales_executive_id = current_user_id;
```

---

## 7. Route Guard Design

### 7.1 Current Limitation

The current route guard uses `meta.role` which only accepts a single role string:

```ts
// Current (limited)
meta: { role: 'ADMINISTRATOR' }
```

### 7.2 Proposed Change

Use a `roles` array instead:

```ts
// Proposed
meta: {
  requiresAuth: true,
  roles: ['ADMINISTRATOR', 'SALES_MANAGER']
}
```

### 7.3 Route Group Design

```ts
// Admin — only ADMINISTRATOR
{
  path: '/admin',
  meta: { requiresAuth: true, roles: ['ADMINISTRATOR'] },
  component: AdminLayout,
  children: [ ... ]
}

// Manager — only SALES_MANAGER
{
  path: '/manager',
  meta: { requiresAuth: true, roles: ['SALES_MANAGER'] },
  component: ManagerLayout,
  children: [ ... ]
}

// Sales — only SALES_EXECUTIVE
{
  path: '/sales',
  meta: { requiresAuth: true, roles: ['SALES_EXECUTIVE'] },
  component: SalesLayout,
  children: [ ... ]
}

// Shared profile — any authenticated role
{
  path: '/profile',
  meta: { requiresAuth: true, roles: ['ADMINISTRATOR', 'SALES_MANAGER', 'SALES_EXECUTIVE'] },
  component: ProfileLayout,
  children: [ ... ]
}
```

### 7.4 Forbidden Route

A `/forbidden` route should be added:

```ts
{
  path: '/forbidden',
  name: 'Forbidden',
  meta: { public: true },
  component: () => import('../views/ForbiddenView.vue')
}
```

And the `beforeEach` guard should redirect to `/forbidden` (not the user's home) when role mismatch occurs.

### Phase 7A-Lite Sales Organization Note
System roles (`ADMINISTRATOR`, `SALES_MANAGER`, `SALES_EXECUTIVE`) remain the authorization keys. Sales organizational roles are configurable labels with numeric Levels 1-4; future team visibility is level-based and team-descendant scoped, not name-based.
