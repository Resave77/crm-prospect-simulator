# Database Architecture

Status: `APPROVED — AUTHENTICATION FOUNDATION IMPLEMENTED; CRM DOMAIN SCHEMA PENDING`

## 1. Ownership and Tooling

PostgreSQL is the system of record. Go repositories own all runtime database access. Prisma is used only for schema definition, migration generation, migration review, and migration execution.

Prohibited:

- Prisma Client as a Go runtime dependency.
- Direct SQL from Fiber handlers.
- Business logic in migration scripts.
- Storing selfie binaries in PostgreSQL.
- Hard deletion of audit-relevant business records.

## 2. Conventions

- Table and column names use `snake_case`.
- Primary keys are UUIDs.
- All business tables contain `created_at` and `updated_at` in `timestamptz`.
- Mutable aggregates use an integer `version` for optimistic concurrency where required.
- Status fields use constrained string values documented below.
- Email comparison uses a normalized lowercase value.
- Coordinates use fixed-precision numeric columns suitable for latitude/longitude.
- Distances and radii use integer meters.
- Flexible JSON is allowed only for external snapshots or non-queryable metadata.
- Foreign keys and essential uniqueness rules are enforced by PostgreSQL, not only services.

## 3. Conceptual Tables

### users

| Column | Purpose |
|---|---|
| `id` | UUID identity |
| `email` | Unique normalized login identifier |
| `password_hash` | bcrypt password hash |
| `full_name` | Display name |
| `role` | `ADMINISTRATOR` or `SALES_EXECUTIVE` |
| `status` | `ACTIVE` or `INACTIVE` |
| `token_version` | Invalidates outstanding sessions after security events |
| `last_login_at` | Successful login audit value |
| timestamps | Lifecycle metadata |

### refresh_sessions

Stores refresh-token sessions rather than raw refresh tokens.

| Column | Purpose |
|---|---|
| `id` | Session UUID and JWT session identifier |
| `user_id` | Session owner |
| `token_hash` | Hash of the refresh-token secret |
| `user_agent`, `ip_address` | Security context |
| `expires_at` | Absolute expiry |
| `revoked_at`, `revoke_reason` | Revocation audit |
| `replaced_by_session_id` | Refresh rotation chain |
| timestamps | Lifecycle metadata |

### prospects

| Column | Purpose |
|---|---|
| `id` | CRM prospect UUID |
| `google_place_id` | Unique provider identity |
| `name`, `category`, `business_status` | Saved business snapshot |
| `address`, `phone`, `website_url`, `google_maps_url` | Saved contact snapshot |
| `latitude`, `longitude` | Saved location |
| `rating`, `user_rating_count` | Saved discovery metadata |
| `place_snapshot` | Versioned JSON snapshot of approved provider data |
| `pipeline_status` | Canonical prospect state |
| `saved_by_user_id` | Administrator who saved it |
| `won_at`, `lost_at`, `converted_at` | Terminal milestone timestamps |
| `version` | Concurrency control |
| timestamps | Lifecycle metadata |

Unique constraint: `google_place_id`.

### prospect_assignments

| Column | Purpose |
|---|---|
| `id` | Assignment UUID |
| `prospect_id` | Assigned prospect |
| `sales_executive_id` | Owner |
| `assigned_by_user_id` | Administrator actor |
| `status` | `ACTIVE`, `REASSIGNED`, or `COMPLETED` |
| `start_date`, `end_date` | Optional assignment window |
| `ended_at`, `end_reason` | Closure audit |
| timestamps | Lifecycle metadata |

A partial unique index permits only one `ACTIVE` assignment per prospect.

### prospect_visits

| Column | Purpose |
|---|---|
| `id` | Visit UUID |
| `prospect_id`, `sales_executive_id` | Subject and actor |
| `status` | `CHECKED_IN` or `CHECKED_OUT` |
| `check_in_at`, `check_out_at` | Server timestamps |
| check-in GPS fields | Latitude, longitude, accuracy |
| check-out GPS fields | Latitude, longitude, accuracy |
| `check_in_selfie_asset_id` | Required evidence reference |
| `notes` | Visit notes |
| `check_in_idempotency_key` | Duplicate-request protection |
| `check_out_idempotency_key` | Duplicate-request protection |
| timestamps | Lifecycle metadata |

### pipeline_history

| Column | Purpose |
|---|---|
| `id` | History UUID |
| `prospect_id` | Aggregate |
| `from_status`, `to_status` | Transition |
| `reason_code`, `notes` | Decision evidence |
| `changed_by_user_id` | Actor |
| `source` | `ADMIN`, `SALES`, or `SYSTEM` |
| `occurred_at` | Immutable event time |

History rows are append-only.

### customers

| Column | Purpose |
|---|---|
| `id` | Customer UUID |
| `customer_code` | Unique business identifier |
| `source_prospect_id` | Unique converted prospect reference |
| `name`, `category`, `address`, contact fields | Existing-customer profile |
| `latitude`, `longitude` | Attendance center |
| `attendance_radius_meters` | Allowed check-in radius |
| `sales_executive_id` | Current owner |
| `status` | `ACTIVE` or `INACTIVE` |
| `additional_data` | Approved non-queryable extension fields |
| `created_by_user_id` | Converting Administrator |
| `version` | Concurrency control |
| timestamps | Lifecycle metadata |

Unique constraints: `customer_code`, `source_prospect_id`.

### customer_attendances

Contains the same check-in/out evidence pattern as prospect visits plus:

- Customer location snapshot.
- Radius snapshot.
- Calculated check-in distance.
- Inside-radius flag.
- Required check-in selfie reference.

Only one open attendance per Sales Executive is allowed by partial unique index.

### media_assets

| Column | Purpose |
|---|---|
| `id` | Asset UUID |
| `owner_user_id` | Uploading user |
| `purpose` | `PROSPECT_CHECK_IN_SELFIE` or `CUSTOMER_ATTENDANCE_SELFIE` |
| `storage_provider`, `object_key` | Object-storage reference |
| `content_type`, `size_bytes`, `checksum` | Validation metadata |
| `status` | `PENDING`, `READY`, `REJECTED`, or `DELETED` |
| timestamps | Lifecycle metadata |

### audit_logs

Append-only audit records include actor, action, entity type/ID, request ID, safe before/after metadata, IP address, user agent, and timestamp. Password/token/media contents are never included.

### settings

Stores controlled operational settings with key, typed value, description, version, and updating Administrator. Secrets belong in deployment environment variables, not this table.

## 4. Relationship Summary

```text
users 1---* refresh_sessions
users 1---* prospect_assignments *---1 prospects
prospects 1---* prospect_visits
prospects 1---* pipeline_history
prospects 1---0..1 customers
users 1---* customers
customers 1---* customer_attendances
media_assets 1---0..1 prospect_visits/customer_attendances
```

## 5. Transaction Boundaries

Mandatory transactions:

- Prospect save and initial history entry.
- Assignment/reassignment and pipeline transition.
- Visit checkout and automatic `VISITED` transition.
- Won/Lost decision and history.
- Prospect-to-customer conversion.
- Refresh-token rotation.

External Google or object-storage calls must not execute inside an open PostgreSQL transaction.

## 6. Index Strategy

Planned indexes support:

- Prospect deduplication by Google Place ID.
- Prospect filtering by pipeline status and creation time.
- Active assignment lookup by Sales Executive.
- Won Review queue.
- Customer lookup by owner/status.
- Open visit and attendance lookup.
- Pipeline/audit history by aggregate and occurrence time.
- Refresh-session lookup and expiry cleanup.

Indexes will be confirmed with actual query plans before production; speculative indexes are avoided.

## 7. Migration Policy

- Every schema change is represented by a reviewed Prisma migration.
- Applied migrations are immutable.
- Destructive changes use expand/migrate/contract phases.
- Production migrations are executed separately from application startup.
- Backups and a rollback/forward-fix procedure are mandatory before risky changes.
- Seed data is limited to controlled initial Administrator provisioning and reference codes; test/demo data is never part of production migrations.

## 8. Retention and Privacy

- Business and audit retention periods must be approved before production.
- Selfie access uses time-limited signed URLs.
- Object keys are never publicly enumerable.
- Location and selfie data are sensitive and visible only to authorized users.
- User deactivation does not delete historical business records.
