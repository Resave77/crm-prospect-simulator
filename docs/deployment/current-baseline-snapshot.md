# Current Baseline Snapshot

Source of truth: local database `crm_prospect_simulator_dev`
Captured: 2026-08-05 (after cleaning duplicate/obsolete August assignments)
Purpose: exact snapshot that `backend/cmd/seed_baseline` reproduces on a fresh database after `npx prisma migrate deploy`.

## 1. Users (all current non-deleted users)

| id | employee_id | full_name | email | role | status | token_version | must_change_password | manager_id | sales_role_id |
|---|---|---|---|---|---|---|---|---|---|
| c67edd75-23fe-4a91-b3b6-4c1181cc9d8d | ADM-0001 | Yummy Super Admin | admin@yummy.test | SUPER_ADMIN | ACTIVE | 2 | false | – | – |
| 3f01df1b-950d-43f1-bce2-d8cc658d2101 | EMP-2026-SAN-8100 | Sandy Ramadhan | sandy@yummy.test | SUPER_ADMIN | ACTIVE | 3 | false | – | – |
| 097e9036-0920-4f4a-a5fb-a01012e5fc70 | EMP-2026-DEL-9732 | tes delete | delete@yummy.test | SUPER_ADMIN | ACTIVE | 1 | false | – | – |
| 1391e74f-f2a4-405f-b924-49bf0617bb5d | EMP-2026-LEV-3075 | Dudung Mujaer | level1@yummy.test | SALES_MANAGER | ACTIVE | 1 | false | – | 00000000-0000-0000-0000-000000000101 |
| e540034d-e1ca-444d-990e-6a55a734e852 | EMP-2026-LEV-6765 | Dobleh Dongdong | level2@yummy.test | SALES_MANAGER | ACTIVE | 1 | false | – | 00000000-0000-0000-0000-000000000102 |
| 113de61e-99f4-41d0-bf9c-db76333783be | EMP-2026-LEV-5264 | Jamal Kabur | level3@yummy.test | SALES_EXECUTIVE | ACTIVE | 1 | false | – | 00000000-0000-0000-0000-000000000104 |
| e83fc480-f228-4f80-bca6-55f9b42736ca | SE-0001 | Nurdin Pratama | sales@yummy.test | SALES_EXECUTIVE | ACTIVE | 5 | false | – | 00000000-0000-0000-0000-000000000104 |
| f0fc9903-cc99-43e5-b08f-7cbc1cb9402f | SE-0002 | Alicia Ramadhan | sales2@yummy.test | SALES_EXECUTIVE | ACTIVE | 1 | false | – | 00000000-0000-0000-0000-000000000104 |
| 4ef6f759-c0e9-45ea-996e-8f07c38ab201 | SE-0003 | Rizky Ananda | sales3@yummy.test | SALES_EXECUTIVE | ACTIVE | 1 | false | – | – |

> `manager@yummy.test` (b6681f56-f389-4504-9f5c-eb2006c35597) is soft-deleted and no longer referenced by any preserved FK (its assignment and all `manager_id` references were removed), so it is NOT seeded. Other soft-deleted users (tes@yummt.test, tes@yummy.test, atmin@yummy.test, level1@yummy.crm) are also NOT seeded.

## 2. Sales roles (all currently active)

| id | name | level | landing_page | description |
|---|---|---|---|---|
| 00000000-0000-0000-0000-000000000101 | Sales Level 1 | 1 | /admin/dashboard | Default editable sales organization role |
| 00000000-0000-0000-0000-000000000102 | Sales Level 2 | 2 | /sales/dashboard | Default editable sales organization role |
| 00000000-0000-0000-0000-000000000103 | Sales Regional Supervisor | 3 | /sales/dashboard | Supervises Level 4 sales in one team |
| 00000000-0000-0000-0000-000000000104 | Sales Level 4 | 4 | /sales/dashboard | Default editable sales organization role |

The inactive `Retired Super Admin` (984cc9a7-3877-52dd-a110-9d33c42c5da7) is NOT seeded.

## 3. Permission mappings per active role

| sales_role_id | permission count |
|---|---|
| 00000000-0000-0000-0000-000000000101 (Sales Level 1) | 71 |
| 00000000-0000-0000-0000-000000000102 (Sales Level 2) | 13 |
| 00000000-0000-0000-0000-000000000103 (Sales Regional Supervisor) | 13 |
| 00000000-0000-0000-0000-000000000104 (Sales Level 4) | 18 |

Total permission grants for active roles: 115.

## 4. Sales structure assignments (all currently present, all 2026-08-01 → 2026-08-31)

| id | user_id | sales_role_id | parent_user_id | created_by |
|---|---|---|---|---|
| 945c73a2-cedb-42bf-88c2-9e78388967e5 | 1391e74f-f2a4-405f-b924-49bf0617bb5d (level1) | 00000000-0000-0000-0000-000000000101 (Sales Level 1) | – | c67edd75-23fe-4a91-b3b6-4c1181cc9d8d |
| 6d0b6630-f256-41bd-8a37-acbff3ab542f | e540034d-e1ca-444d-990e-6a55a734e852 (level2) | 00000000-0000-0000-0000-000000000102 (Sales Level 2) | 1391e74f-f2a4-405f-b924-49bf0617bb5d | c67edd75-23fe-4a91-b3b6-4c1181cc9d8d |
| 989db891-e70e-4f2a-a414-6b5f24a3a6c7 | 113de61e-99f4-41d0-bf9c-db76333783be (level3) | 00000000-0000-0000-0000-000000000103 (Sales Regional Supervisor) | e540034d-e1ca-444d-990e-6a55a734e852 | c67edd75-23fe-4a91-b3b6-4c1181cc9d8d |
| a57c54e4-c2f8-4d10-aa4a-a1fbd6e3324e | e83fc480-f228-4f80-bca6-55f9b42736ca (sales) | 00000000-0000-0000-0000-000000000104 (Sales Level 4) | 113de61e-99f4-41d0-bf9c-db76333783be | c67edd75-23fe-4a91-b3b6-4c1181cc9d8d |
| 18209b51-ddb5-4990-b8f3-e7dc68998d51 | f0fc9903-cc99-43e5-b08f-7cbc1cb9402f (sales2) | 00000000-0000-0000-0000-000000000104 (Sales Level 4) | 113de61e-99f4-41d0-bf9c-db76333783be | c67edd75-23fe-4a91-b3b6-4c1181cc9d8d |
| f93b718f-d2ad-5189-b75a-b6224252750e | 4ef6f759-c0e9-45ea-996e-8f07c38ab201 (sales3) | 00000000-0000-0000-0000-000000000104 (Sales Level 4) | 113de61e-99f4-41d0-bf9c-db76333783be | c67edd75-23fe-4a91-b3b6-4c1181cc9d8d |

## 5. Desired August 2026 hierarchy

```
level1@yummy.test        Sales Level 1             (parent: null)
└── level2@yummy.test     Sales Level 2             (parent: level1)
    └── level3@yummy.test Sales Regional Supervisor (parent: level2)
        ├── sales@yummy.test    Sales Level 4       (parent: level3)
        ├── sales2@yummy.test   Sales Level 4       (parent: level3)
        └── sales3@yummy.test   Sales Level 4       (parent: level3)
```

`admin@yummy.test` is SUPER_ADMIN only: no `sales_role_id`, no sales structure assignment.

## 6. Snapshot totals

| item | count |
|---|---|
| users seeded | 9 |
| active sales roles seeded | 4 |
| permission grants for active roles | 115 |
| sales structure assignments | 6 |
