# Account Data Model

> **Document:** 03-account-data-model.md  
> **Project:** CRM Prospect Simulator  
> **Date:** 2026-07-30  
> **Status:** Draft — Design Phase  

---

## 1. User Fields (Final)

| Field | Type | Required | Unique | Mutable | Notes |
|---|---|---|---|---|---|
| `id` | UUID (PK) | Yes | Yes | No | Auto-generated |
| `employee_id` | String | Yes | Yes | Caution | Format: `ADM-0001`, `SM-0001`, `SE-0001` |
| `email` | String | Yes | Yes | Yes | Lowercase, used for login |
| `password_hash` | String | Yes | No | Yes | bcrypt hash only |
| `full_name` | String | Yes | No | Yes | Display name |
| `phone` | String | Recommended | No | Yes | Indonesia format |
| `role` | Enum | Yes | No | Caution | `ADMINISTRATOR`, `SALES_MANAGER`, `SALES_EXECUTIVE` |
| `manager_id` | UUID (FK→users) | See rules | No | Yes | Required for SE only |
| `status` | Enum | Yes | No | Yes | `ACTIVE`, `INACTIVE`, (`LOCKED` future) |
| `must_change_password` | Boolean | Yes | No | Yes | Default `true` on first creation |
| `token_version` | Int | Yes | No | Yes | Default 1, incremented on password reset |
| `last_login_at` | DateTime? | No | No | Yes | Set on successful login |
| `created_by` | UUID (FK→users) | Yes | No | No | Admin who created the account |
| `updated_by` | UUID (FK→users) | Yes | No | Yes | Last admin who updated |
| `created_at` | DateTime | Yes | No | No | Auto-set |
| `updated_at` | DateTime | Yes | No | No | Auto-updated |

---

## 2. Manager Relationship

### 2.1 Structure

```
Administrator
    │
    └── Sales Manager  (manager_id = NULL  or ADMINISTRATOR ID)
           │
           └── Sales Executive  (manager_id = SALES_MANAGER ID)
```

### 2.2 Rules

| Rule | Condition |
|---|---|
| 1. SE must have a manager | `role = 'SALES_EXECUTIVE'` → `manager_id` is **required** |
| 2. Manager must be SALES_MANAGER | `manager_id` must reference a user with `role = 'SALES_MANAGER'` |
| 3. SM manager_id optional | `role = 'SALES_MANAGER'` → `manager_id` may be `NULL` |
| 4. ADMIN has no manager | `role = 'ADMINISTRATOR'` → `manager_id` must be `NULL` |
| 5. No self-reference | `manager_id` cannot equal `id` |
| 6. Manager must be ACTIVE | Referenced manager must have `status = 'ACTIVE'` |
| 7. No circular relationship | A->B->A is forbidden (validate on write) |
| 8. One manager per SE | Single `manager_id` |
| 9. Multiple SE per manager | No limit on SE count per manager |

### 2.3 Conceptual Schema Change

```prisma
model User {
  // ... existing fields ...
  managerId   String?    @map("manager_id") @db.Uuid
  manager     User?      @relation("UserManager", fields: [managerId], references: [id])
  subordinates User[]    @relation("UserManager")

  createdBy   String?    @map("created_by") @db.Uuid
  createdByUser User?    @relation("UserCreatedBy", fields: [createdBy], references: [id])
  updatedBy   String?    @map("updated_by") @db.Uuid
  updatedByUser User?    @relation("UserUpdatedBy", fields: [updatedBy], references: [id])
}
```

---

## 3. Field Validation Rules

### 3.1 Employee ID

```text
Format:  PREFIX-XXXX
Prefix:  ADM / SM / SE
XXXX:    Zero-padded incrementing number (0001, 0002, ...)

Examples:
  ADM-0001   → Administrator
  SM-0001    → Sales Manager
  SE-0001    → Sales Executive
```

- Generated server-side on account creation
- Admin may override if unassigned
- Immutable after first activation (except by admin override with reason)

### 3.2 Email

- Must be a valid email format
- Lowercased before storage (as currently implemented)
- Must be unique across all users
- Used as login credential

### 3.3 Phone

- Required for SALES_MANAGER and SALES_EXECUTIVE
- Optional for ADMINISTRATOR
- Indonesia format recommended (`+62` or `08xx` prefix)
- Not used as credential

### 3.4 Full Name

- Required, no uniqueness constraint
- Reasonable length limit (e.g., 200 characters)

---

## 4. Account Status

| Status | Can Login | Can Refresh | Can Access | Session Status |
|---|---|---|---|---|
| `ACTIVE` | Yes | Yes | Yes (per role) | Normal |
| `INACTIVE` | No | No | No | Revoked on deactivation |
| `LOCKED` (future) | No | No | No | Revoked on lock |

### 4.1 Status Transitions

```
ACTIVE  ←→  INACTIVE
ACTIVE  ——→  LOCKED (future phase)
LOCKED  ←——  ACTIVE (future: admin unlock)
```

### 4.2 Deactivation Behavior

- Login is rejected with `ACCOUNT_INACTIVE` error
- Token refresh is rejected
- All active sessions are revoked immediately
- Historical data (prospects, customers, visits) is preserved
- Account is not deleted — all data remains associated

---

## 5. Account Creation Flow

### 5.1 Create Account (Admin)

```
Admin clicks "Create Account"
         │
         ▼
    [Account Identity Section]
         │
         ├── Employee ID          (auto-generated, editable)
         ├── Full Name            (required)
         ├── Email                (required, unique)
         └── Phone                (recommended)
         │
         ▼
    [Role & Reporting Section]
         │
         ├── Role                (dropdown: ADMINISTRATOR / SALES_MANAGER / SALES_EXECUTIVE)
         │
         ├── If SALES_EXECUTIVE:
         │   └── Manager         (dropdown of ACTIVE SALES_MANAGER users)
         │
         └── If SALES_MANAGER or ADMINISTRATOR:
             └── Manager         (hidden, auto-set to NULL)
         │
         ▼
    [Role Access Summary]         (read-only, updates on role change)
         │
         ▼
    [Security Section]
         │
         ├── Status              (ACTIVE / INACTIVE)
         ├── Password Mode       (dropdown: Auto-generate / Manual)
         │   ├── Auto:           system generates random password (shown once)
         │   └── Manual:         Admin enters temporary password
         └── Must Change Password (auto-set true)
         │
         ▼
    [Summary & Confirmation]
         │
         ▼
    Submit → Backend Validation → Create → Show Temporary Password (once)
```

### 5.2 Validation Rules (Backend)

| Rule | Error Code | HTTP Status |
|---|---|---|
| Email already exists | `EMAIL_TAKEN` | 409 |
| Employee ID already exists | `EMPLOYEE_ID_TAKEN` | 409 |
| Role required | `ROLE_REQUIRED` | 422 |
| Manager required for SE | `MANAGER_REQUIRED` | 422 |
| Manager must be SALES_MANAGER | `INVALID_MANAGER_ROLE` | 422 |
| Manager must be ACTIVE | `INACTIVE_MANAGER` | 422 |
| Manager cannot be self | `SELF_MANAGER` | 422 |
| ADMIN cannot have manager | `ADMIN_NO_MANAGER` | 422 |
| Invalid email format | `INVALID_EMAIL` | 422 |
| Password too short (if manual) | `WEAK_PASSWORD` | 422 |

---

## 6. Edit Account Flow

### 6.1 Editable Fields

| Field | Editable? | Notes |
|---|---|---|
| `full_name` | Yes | |
| `email` | Yes | Must remain unique |
| `phone` | Yes | |
| `role` | Caution | Clear implications (see below) |
| `manager_id` | Yes | Only for SE |
| `status` | Yes | ACTIVE ↔ INACTIVE |

### 6.2 Role Change Rules

- Cannot change role if it would orphan records (SE with prospects/customers being changed to ADMIN)
- Role change may require reassignment of subordinates (if SM → SE)
- Audit log must record the change

### 6.3 Manager Change Rules

- Previous manager loses access to this SE's data
- New manager gains access immediately
- No retrospective data migration needed (scoped by `manager_id`)

---

## 7. Relationship Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         users (self-referencing)                 │
│                                                                   │
│  ┌──────────┐       ┌──────────────┐       ┌──────────────────┐  │
│  │ADMINISTRA│       │SALES_MANAGER │       │ SALES_EXECUTIVE  │  │
│  │  TOR     │       │              │       │                  │  │
│  │          │       │              │       │                  │  │
│  │manager_id│──NULL │manager_id│───│NULL   │manager_id│───│FK  │  │
│  │          │       │              │       │         │        │  │
│  └──────────┘       └──────┬───────┘       └─────────┼────────┘  │
│                            │                        │           │
│                            │  has many SEs          │ has one SM│
│                            └────────────────────────┘           │
│                                                                   │
│  ┌─────────────────────────────────────────────────────────┐     │
│  │  prospects                                              │     │
│  │  assigned_sales_executive_id ────────────────────────FK──┼──→ │
│  └─────────────────────────────────────────────────────────┘     │
│  ┌─────────────────────────────────────────────────────────┐     │
│  │  customer_sites                                          │     │
│  │  sales_executive_id ────────────────────────────────FK──┼──→ │
│  └─────────────────────────────────────────────────────────┘     │
│  ┌─────────────────────────────────────────────────────────┐     │
│  │  refresh_sessions                                        │     │
│  │  user_id ───────────────────────────────────────────FK──┼──→ │
│  └─────────────────────────────────────────────────────────┘     │
└─────────────────────────────────────────────────────────────────┘
```

---

## 8. Seed Data Plan

Current hardcoded accounts in `cmd/seed/main.go` must be extended:

```go
users := []model.User{
  {FullName: "Yummy Administrator", Role: model.RoleAdministrator, ...},
  {FullName: "Nurdin Pratama",      Role: model.RoleSalesExecutive, ...},
  {FullName: "Alicia Ramadhan",     Role: model.RoleSalesExecutive, ...},
  {FullName: "Rizky Ananda",        Role: model.RoleSalesExecutive, ...},
}
```

**Proposed seed accounts after Phase 3A:**

| # | Full Name | Role | Manager | Status |
|---|---|---|---|---|
| 1 | Yummy Administrator | ADMINISTRATOR | — | ACTIVE |
| 2 | Budi Santoso | SALES_MANAGER | — | ACTIVE |
| 3 | Nurdin Pratama | SALES_EXECUTIVE | Budi Santoso (SM) | ACTIVE |
| 4 | Alicia Ramadhan | SALES_EXECUTIVE | Budi Santoso (SM) | ACTIVE |
| 5 | Rizky Ananda | SALES_EXECUTIVE | Budi Santoso (SM) | ACTIVE |

This gives the system one manager with three SEs for immediate testing.
