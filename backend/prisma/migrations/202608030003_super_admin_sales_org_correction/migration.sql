-- Promote the primary administrator account to SUPER_ADMIN with the official
-- display name. This migration runs in a separate transaction from the enum
-- migration above so the new enum value is committed before it is used.
UPDATE "users"
SET "full_name" = 'Yummy Super Admin',
    "role" = 'SUPER_ADMIN',
    "updated_at" = now()
WHERE "email" = 'admin@yummy.test';

-- Deactivate obsolete sales organization roles. Historical assignments that
-- reference these roles are preserved; the sales structure queries do not
-- filter on is_active, so deactivation is safe and reversible.
UPDATE "sales_roles"
SET "is_active" = false,
    "updated_at" = now()
WHERE "normalized_name" IN (
    'sales level 1',
    'sales level 2',
    'sales level 3',
    'sales level 3 + billing',
    'admin sales',
    'sales regional supervisor'
);

-- Rename the canonical demo Level 1 role to the official "Super Admin" name.
-- The demo seed later reconciles the remaining roles idempotently by id.
UPDATE "sales_roles"
SET "name" = 'Super Admin',
    "normalized_name" = 'super admin',
    "level" = 1,
    "description" = 'Top-level sales organization role for Super Admin.',
    "is_active" = true,
    "updated_at" = now()
WHERE "id" = '984cc9a7-3877-52dd-a110-9d33c42c5da7';

-- Give the primary administrator the single current Level 1 assignment,
-- effective from August 2026, with no parent (root of the tree).
INSERT INTO "sales_structure_assignments" ("id", "user_id", "sales_role_id", "parent_user_id", "effective_from", "updated_at")
SELECT '00000000-0000-0000-0000-0000000001a1', u."id", '984cc9a7-3877-52dd-a110-9d33c42c5da7', NULL, DATE '2026-08-01', now()
FROM "users" u
WHERE u."email" = 'admin@yummy.test'
ON CONFLICT ("id") DO UPDATE SET
    "user_id" = EXCLUDED."user_id",
    "sales_role_id" = EXCLUDED."sales_role_id",
    "parent_user_id" = NULL,
    "effective_from" = EXCLUDED."effective_from",
    "effective_to" = NULL,
    "updated_at" = now();

-- Re-parent current Level 2 assignments under the administrator.
UPDATE "sales_structure_assignments" child
SET "parent_user_id" = super_admin."id",
    "updated_at" = now()
FROM "users" super_admin,
     "sales_roles" child_role
WHERE super_admin."email" = 'admin@yummy.test'
  AND child."sales_role_id" = child_role."id"
  AND child_role."level" = 2
  AND child."effective_to" IS NULL;
