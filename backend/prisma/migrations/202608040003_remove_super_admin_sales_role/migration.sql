WITH super_admin_sales_roles AS (
    SELECT id
    FROM "sales_roles"
    WHERE lower(btrim("name")) = 'super admin'
       OR lower(btrim("normalized_name")) = 'super admin'
)
DELETE FROM "sales_structure_assignments"
WHERE "sales_role_id" IN (SELECT id FROM super_admin_sales_roles);

WITH super_admin_sales_roles AS (
    SELECT id
    FROM "sales_roles"
    WHERE lower(btrim("name")) = 'super admin'
       OR lower(btrim("normalized_name")) = 'super admin'
)
UPDATE "users"
SET "sales_role_id" = NULL,
    "manager_id" = CASE WHEN "role" = 'SUPER_ADMIN' THEN NULL ELSE "manager_id" END,
    "updated_at" = now()
WHERE "sales_role_id" IN (SELECT id FROM super_admin_sales_roles);

UPDATE "users"
SET "role" = 'SUPER_ADMIN',
    "sales_role_id" = NULL,
    "manager_id" = NULL,
    "updated_at" = now()
WHERE lower("email") = 'admin@yummy.test';

UPDATE "sales_roles"
SET "name" = CASE
        WHEN "name" = 'Super Admin' THEN 'Retired Super Admin'
        ELSE "name"
    END,
    "normalized_name" = CASE
        WHEN lower(btrim("normalized_name")) = 'super admin' THEN 'retired super admin'
        ELSE "normalized_name"
    END,
    "is_active" = false,
    "updated_at" = now()
WHERE lower(btrim("name")) = 'super admin'
   OR lower(btrim("normalized_name")) = 'super admin';
