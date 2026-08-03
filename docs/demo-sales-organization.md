# Demo Sales Organization Dataset

This local-development dataset populates Role Management and Sales Structure with an enterprise-style FMCG sales organization for demos and manual testing.

Do not use this seed in production or against a shared database. The commands refuse database URLs that are not recognized as local.

## Dataset

- Users: 52 demo users
- Current assigned hierarchy: 1 Level 1, 3 Level 2, 8 Level 3, 36 Level 4
- Active unassigned users: 4
- Regions represented: Jakarta, Bandung, Semarang, Surabaya, Malang, Medan, Makassar
- Demo email domain: `@demo.yummy.local`
- Demo employee ID prefix: `DEMO-`

## Commands

From `backend`:

```bash
go run ./cmd/seed_demo
```

Cleanup only demo-marked records:

```bash
go run ./cmd/cleanup_demo
```

Cleanup deletes demo sales assignments, demo sessions, demo users, and demo roles only when no non-demo assignments still reference those roles.

## Demo Credentials

All demo accounts use the same local temporary password:

```text
DemoYummy2026!
```

The password is stored only as a bcrypt hash in the database. Demo accounts are seeded with `must_change_password=false` to simplify demos.

## Position Limitation

The current schema has no dedicated Position field. Position-like demo context is included in display-compatible names such as "Regional Manager Jakarta" and role labels such as "Sales Level 4 + Merchandising". Hierarchy logic still depends only on numeric role levels 1-4.

## Safety

The seed is idempotent and uses deterministic IDs. Rerunning updates safe demo fields and skips unchanged rows.

Do not run cleanup automatically. Cleanup is intended for local demo reset only.
