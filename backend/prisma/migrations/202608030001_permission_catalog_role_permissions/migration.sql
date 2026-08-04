ALTER TABLE "sales_roles" ADD COLUMN IF NOT EXISTS "landing_page" VARCHAR(255);

CREATE TABLE IF NOT EXISTS "permissions" (
    "id" UUID NOT NULL,
    "key" VARCHAR(255) NOT NULL,
    "name" VARCHAR(255) NOT NULL,
    "description" TEXT,
    "group_key" VARCHAR(255) NOT NULL,
    "parent_key" VARCHAR(255),
    "node_type" VARCHAR(32) NOT NULL,
    "route_path" VARCHAR(255),
    "is_active" BOOLEAN NOT NULL DEFAULT true,
    "sort_order" INTEGER NOT NULL DEFAULT 0,
    "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMPTZ(6) NOT NULL DEFAULT now(),
    CONSTRAINT "permissions_pkey" PRIMARY KEY ("id"),
    CONSTRAINT "permissions_key_key" UNIQUE ("key"),
    CONSTRAINT "permissions_node_type_check" CHECK ("node_type" IN ('GROUP','MENU','ACTION'))
);

CREATE INDEX IF NOT EXISTS "permissions_group_key_sort_order_idx" ON "permissions"("group_key", "sort_order");
CREATE INDEX IF NOT EXISTS "permissions_parent_key_idx" ON "permissions"("parent_key");

CREATE TABLE IF NOT EXISTS "role_permissions" (
    "sales_role_id" UUID NOT NULL,
    "permission_id" UUID NOT NULL,
    "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT now(),
    CONSTRAINT "role_permissions_pkey" PRIMARY KEY ("sales_role_id", "permission_id")
);

CREATE INDEX IF NOT EXISTS "role_permissions_permission_id_idx" ON "role_permissions"("permission_id");
ALTER TABLE "role_permissions" ADD CONSTRAINT "role_permissions_sales_role_id_fkey" FOREIGN KEY ("sales_role_id") REFERENCES "sales_roles"("id") ON DELETE CASCADE ON UPDATE CASCADE;
ALTER TABLE "role_permissions" ADD CONSTRAINT "role_permissions_permission_id_fkey" FOREIGN KEY ("permission_id") REFERENCES "permissions"("id") ON DELETE CASCADE ON UPDATE CASCADE;

WITH catalog(key, name, group_key, parent_key, node_type, route_path, sort_order) AS (
  VALUES
    ('menu_admin_dashboard', 'Admin Dashboard', 'dashboard', NULL, 'MENU', '/admin/dashboard', 10),
    ('view_admin_dashboard', 'View Admin Dashboard', 'dashboard', 'menu_admin_dashboard', 'ACTION', '/admin/dashboard', 11),
    ('menu_sales_dashboard', 'Sales Dashboard', 'dashboard', NULL, 'MENU', '/sales/dashboard', 12),
    ('view_sales_dashboard', 'View Sales Dashboard', 'dashboard', 'menu_sales_dashboard', 'ACTION', '/sales/dashboard', 13),
    ('menu_accounts', 'Accounts', 'accounts', NULL, 'MENU', '/admin/accounts', 20),
    ('view_accounts', 'View Accounts', 'accounts', 'menu_accounts', 'ACTION', '/admin/accounts', 21),
    ('view_account_detail', 'View Account Detail', 'accounts', 'menu_accounts', 'ACTION', NULL, 22),
    ('create_account', 'Create Account', 'accounts', 'menu_accounts', 'ACTION', NULL, 23),
    ('update_account', 'Update Account', 'accounts', 'menu_accounts', 'ACTION', NULL, 24),
    ('update_account_status', 'Update Account Status', 'accounts', 'menu_accounts', 'ACTION', NULL, 25),
    ('reset_account_password', 'Reset Account Password', 'accounts', 'menu_accounts', 'ACTION', NULL, 26),
    ('menu_roles', 'Roles', 'roles', NULL, 'MENU', '/admin/role-management', 30),
    ('view_roles', 'View Roles', 'roles', 'menu_roles', 'ACTION', '/admin/role-management', 31),
    ('view_role_detail', 'View Role Detail', 'roles', 'menu_roles', 'ACTION', NULL, 32),
    ('create_role', 'Create Role', 'roles', 'menu_roles', 'ACTION', NULL, 33),
    ('update_role', 'Update Role', 'roles', 'menu_roles', 'ACTION', NULL, 34),
    ('update_role_status', 'Update Role Status', 'roles', 'menu_roles', 'ACTION', NULL, 35),
    ('delete_role', 'Delete Role', 'roles', 'menu_roles', 'ACTION', NULL, 36),
    ('manage_role_permissions', 'Manage Role Permissions', 'roles', 'menu_roles', 'ACTION', NULL, 37),
    ('menu_sales_structure', 'Sales Structure', 'sales_structure', NULL, 'MENU', '/admin/sales-structure', 40),
    ('view_sales_structure', 'View Sales Structure', 'sales_structure', 'menu_sales_structure', 'ACTION', '/admin/sales-structure', 41),
    ('create_sales_assignment', 'Create Assignment', 'sales_structure', 'menu_sales_structure', 'ACTION', NULL, 42),
    ('move_sales_assignment', 'Move Assignment', 'sales_structure', 'menu_sales_structure', 'ACTION', NULL, 43),
    ('view_sales_assignment_history', 'View Assignment History', 'sales_structure', 'menu_sales_structure', 'ACTION', NULL, 44),
    ('menu_prospect_finder', 'Prospect Finder', 'prospects', NULL, 'MENU', '/admin/prospect-finder', 50),
    ('menu_prospect_list', 'Prospect List', 'prospects', NULL, 'MENU', '/admin/prospects/list', 51),
    ('menu_prospect_pipeline', 'Prospect Pipeline', 'prospects', NULL, 'MENU', '/admin/prospects/pipeline', 52),
    ('menu_my_prospects', 'My Prospects', 'prospects', NULL, 'MENU', '/sales/my-prospects', 53),
    ('menu_sales_pipeline', 'Sales Pipeline', 'prospects', NULL, 'MENU', '/sales/pipeline', 54),
    ('view_prospect_finder', 'View Prospect Finder', 'prospects', 'menu_prospect_finder', 'ACTION', '/admin/prospect-finder', 55),
    ('view_prospect_list', 'View Prospect List', 'prospects', 'menu_prospect_list', 'ACTION', '/admin/prospects/list', 56),
    ('view_prospect_pipeline', 'View Prospect Pipeline', 'prospects', 'menu_prospect_pipeline', 'ACTION', '/admin/prospects/pipeline', 57),
    ('view_prospect_detail', 'View Prospect Detail', 'prospects', 'menu_prospect_list', 'ACTION', NULL, 58),
    ('view_my_prospects', 'View My Prospects', 'prospects', 'menu_my_prospects', 'ACTION', '/sales/my-prospects', 59),
    ('view_my_prospect_detail', 'View My Prospect Detail', 'prospects', 'menu_my_prospects', 'ACTION', NULL, 60),
    ('create_prospect', 'Create Prospect', 'prospects', 'menu_prospect_finder', 'ACTION', NULL, 61),
    ('update_prospect_pipeline', 'Update Prospect Pipeline', 'prospects', 'menu_sales_pipeline', 'ACTION', NULL, 62),
    ('delete_prospect', 'Delete Prospect', 'prospects', 'menu_prospect_list', 'ACTION', NULL, 63),
    ('request_prospect_deletion', 'Request Prospect Deletion', 'prospects', 'menu_my_prospects', 'ACTION', NULL, 64),
    ('approve_prospect_deletion', 'Approve Prospect Deletion', 'prospects', 'menu_prospect_list', 'ACTION', NULL, 65),
    ('reject_prospect_deletion', 'Reject Prospect Deletion', 'prospects', 'menu_prospect_list', 'ACTION', NULL, 66),
    ('convert_prospect', 'Convert Prospect', 'prospects', 'menu_prospect_list', 'ACTION', NULL, 67),
    ('manage_prospect_comments', 'Comments', 'prospects', 'menu_prospect_list', 'ACTION', NULL, 68),
    ('menu_customers', 'Customers', 'customers', NULL, 'MENU', '/admin/customers', 70),
    ('menu_my_customers', 'My Customers', 'customers', NULL, 'MENU', '/sales/my-customers', 71),
    ('view_customers', 'View Customers', 'customers', 'menu_customers', 'ACTION', '/admin/customers', 72),
    ('view_customer_detail', 'View Customer Detail', 'customers', 'menu_customers', 'ACTION', NULL, 73),
    ('view_my_customers', 'View My Customers', 'customers', 'menu_my_customers', 'ACTION', '/sales/my-customers', 74),
    ('view_my_customer_detail', 'View My Customer Detail', 'customers', 'menu_my_customers', 'ACTION', NULL, 75),
    ('create_customer', 'Create Customer', 'customers', 'menu_customers', 'ACTION', NULL, 76),
    ('update_customer', 'Update Customer', 'customers', 'menu_customers', 'ACTION', NULL, 77),
    ('delete_customer', 'Delete Customer', 'customers', 'menu_customers', 'ACTION', NULL, 78),
    ('view_company_detail', 'View Company Detail', 'customers', 'menu_customers', 'ACTION', NULL, 79),
    ('update_company', 'Update Company', 'customers', 'menu_customers', 'ACTION', NULL, 80),
    ('menu_visit_monitoring', 'Visit Monitoring', 'visits', NULL, 'MENU', '/admin/visit-monitoring', 90),
    ('view_visit_monitoring', 'View Visit Monitoring', 'visits', 'menu_visit_monitoring', 'ACTION', '/admin/visit-monitoring', 91),
    ('view_own_visits', 'View Own Visits', 'visits', NULL, 'ACTION', NULL, 92),
    ('check_in_prospect', 'Check In Prospect', 'visits', 'menu_my_prospects', 'ACTION', NULL, 93),
    ('check_out_prospect', 'Check Out Prospect', 'visits', 'menu_my_prospects', 'ACTION', NULL, 94),
    ('check_in_customer', 'Check In Customer', 'visits', 'menu_my_customers', 'ACTION', NULL, 95),
    ('check_out_customer', 'Check Out Customer', 'visits', 'menu_my_customers', 'ACTION', NULL, 96),
    ('update_visit_result', 'Update Visit Result', 'visits', NULL, 'ACTION', NULL, 97),
    ('view_visit_evidence', 'View Visit Evidence', 'visits', 'menu_visit_monitoring', 'ACTION', NULL, 98),
    ('delete_visit', 'Delete Visit', 'visits', NULL, 'ACTION', NULL, 99),
    ('menu_reports', 'Reports', 'reports', NULL, 'MENU', '/admin/reports', 100),
    ('view_reports', 'View Reports', 'reports', 'menu_reports', 'ACTION', '/admin/reports', 101),
    ('menu_sales_history', 'Sales History', 'profile', NULL, 'MENU', '/sales/history', 110),
    ('view_sales_history', 'View Sales History', 'profile', 'menu_sales_history', 'ACTION', '/sales/history', 111),
    ('menu_profile', 'Profile', 'profile', NULL, 'MENU', '/sales/profile', 112),
    ('view_own_profile', 'View Own Profile', 'profile', 'menu_profile', 'ACTION', '/sales/profile', 113),
    ('change_own_password', 'Change Own Password', 'profile', NULL, 'ACTION', NULL, 114)
)
INSERT INTO "permissions" ("id", "key", "name", "description", "group_key", "parent_key", "node_type", "route_path", "sort_order", "updated_at")
SELECT (
    substr(md5(key), 1, 8) || '-' || substr(md5(key), 9, 4) || '-' || substr(md5(key), 13, 4) || '-' || substr(md5(key), 17, 4) || '-' || substr(md5(key), 21, 12)
  )::uuid,
  key, name, '', group_key, parent_key, node_type, route_path, sort_order, now()
FROM catalog
ON CONFLICT ("key") DO UPDATE SET
    "name" = EXCLUDED."name",
    "description" = EXCLUDED."description",
    "group_key" = EXCLUDED."group_key",
    "parent_key" = EXCLUDED."parent_key",
    "node_type" = EXCLUDED."node_type",
    "route_path" = EXCLUDED."route_path",
    "sort_order" = EXCLUDED."sort_order",
    "is_active" = true,
    "updated_at" = now();

UPDATE "sales_roles" SET "landing_page" = '/admin/dashboard' WHERE "id" = '00000000-0000-0000-0000-000000000101' AND "landing_page" IS NULL;
UPDATE "sales_roles" SET "landing_page" = '/sales/dashboard' WHERE "id" IN ('00000000-0000-0000-0000-000000000102','00000000-0000-0000-0000-000000000103','00000000-0000-0000-0000-000000000104') AND "landing_page" IS NULL;

WITH default_keys(sales_role_id, key) AS (
    SELECT '00000000-0000-0000-0000-000000000101'::uuid, key FROM "permissions" WHERE is_active = true
    UNION ALL SELECT '00000000-0000-0000-0000-000000000102'::uuid, unnest(ARRAY['menu_sales_dashboard','view_sales_dashboard','menu_sales_structure','view_sales_structure','menu_my_prospects','view_my_prospects','menu_my_customers','view_my_customers','menu_sales_history','view_sales_history','menu_profile','view_own_profile','change_own_password'])
    UNION ALL SELECT '00000000-0000-0000-0000-000000000103'::uuid, unnest(ARRAY['menu_sales_dashboard','view_sales_dashboard','menu_my_prospects','view_my_prospects','view_my_prospect_detail','menu_my_customers','view_my_customers','view_my_customer_detail','menu_sales_history','view_sales_history','menu_profile','view_own_profile','change_own_password'])
    UNION ALL SELECT '00000000-0000-0000-0000-000000000104'::uuid, unnest(ARRAY['menu_sales_dashboard','view_sales_dashboard','menu_my_prospects','view_my_prospects','view_my_prospect_detail','check_in_prospect','check_out_prospect','update_visit_result','menu_my_customers','view_my_customers','view_my_customer_detail','check_in_customer','check_out_customer','menu_sales_history','view_sales_history','menu_profile','view_own_profile','change_own_password'])
)
INSERT INTO "role_permissions" ("sales_role_id", "permission_id")
SELECT dk.sales_role_id, p.id
FROM default_keys dk
JOIN "permissions" p ON p.key = dk.key
JOIN "sales_roles" r ON r.id = dk.sales_role_id
ON CONFLICT DO NOTHING;

