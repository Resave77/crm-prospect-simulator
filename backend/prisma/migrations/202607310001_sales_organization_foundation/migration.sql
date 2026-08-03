CREATE TABLE "sales_roles" (
    "id" UUID NOT NULL,
    "name" VARCHAR(255) NOT NULL,
    "normalized_name" VARCHAR(255) NOT NULL,
    "level" SMALLINT NOT NULL,
    "description" TEXT,
    "is_active" BOOLEAN NOT NULL DEFAULT true,
    "created_by" UUID,
    "updated_by" UUID,
    "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMPTZ(6) NOT NULL DEFAULT now(),
    CONSTRAINT "sales_roles_pkey" PRIMARY KEY ("id"),
    CONSTRAINT "sales_roles_level_check" CHECK ("level" BETWEEN 1 AND 4),
    CONSTRAINT "sales_roles_name_check" CHECK (length(btrim("name")) > 0)
);

CREATE UNIQUE INDEX "sales_roles_active_normalized_name_key"
    ON "sales_roles"("normalized_name") WHERE "is_active" = true;

CREATE TABLE "sales_structure_assignments" (
    "id" UUID NOT NULL,
    "user_id" UUID NOT NULL,
    "sales_role_id" UUID NOT NULL,
    "parent_user_id" UUID,
    "effective_from" DATE NOT NULL,
    "effective_to" DATE,
    "created_by" UUID,
    "created_at" TIMESTAMPTZ(6) NOT NULL DEFAULT now(),
    "updated_at" TIMESTAMPTZ(6) NOT NULL DEFAULT now(),
    CONSTRAINT "sales_structure_assignments_pkey" PRIMARY KEY ("id"),
    CONSTRAINT "sales_structure_assignments_period_check" CHECK ("effective_to" IS NULL OR "effective_to" >= "effective_from"),
    CONSTRAINT "sales_structure_assignments_self_parent_check" CHECK ("parent_user_id" IS NULL OR "user_id" <> "parent_user_id")
);

CREATE INDEX "sales_structure_assignments_user_id_idx" ON "sales_structure_assignments"("user_id");
CREATE INDEX "sales_structure_assignments_parent_user_id_idx" ON "sales_structure_assignments"("parent_user_id");
CREATE INDEX "sales_structure_assignments_sales_role_id_idx" ON "sales_structure_assignments"("sales_role_id");
CREATE INDEX "sales_structure_assignments_effective_from_idx" ON "sales_structure_assignments"("effective_from");
CREATE INDEX "sales_structure_assignments_effective_to_idx" ON "sales_structure_assignments"("effective_to");

ALTER TABLE "sales_roles" ADD CONSTRAINT "sales_roles_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "users"("id") ON DELETE SET NULL ON UPDATE CASCADE;
ALTER TABLE "sales_roles" ADD CONSTRAINT "sales_roles_updated_by_fkey" FOREIGN KEY ("updated_by") REFERENCES "users"("id") ON DELETE SET NULL ON UPDATE CASCADE;
ALTER TABLE "sales_structure_assignments" ADD CONSTRAINT "sales_structure_assignments_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "users"("id") ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE "sales_structure_assignments" ADD CONSTRAINT "sales_structure_assignments_parent_user_id_fkey" FOREIGN KEY ("parent_user_id") REFERENCES "users"("id") ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE "sales_structure_assignments" ADD CONSTRAINT "sales_structure_assignments_sales_role_id_fkey" FOREIGN KEY ("sales_role_id") REFERENCES "sales_roles"("id") ON DELETE RESTRICT ON UPDATE CASCADE;
ALTER TABLE "sales_structure_assignments" ADD CONSTRAINT "sales_structure_assignments_created_by_fkey" FOREIGN KEY ("created_by") REFERENCES "users"("id") ON DELETE SET NULL ON UPDATE CASCADE;

INSERT INTO "sales_roles" ("id", "name", "normalized_name", "level", "description")
VALUES
    ('00000000-0000-0000-0000-000000000101', 'Sales Level 1', 'sales level 1', 1, 'Default editable sales organization role'),
    ('00000000-0000-0000-0000-000000000102', 'Sales Level 2', 'sales level 2', 2, 'Default editable sales organization role'),
    ('00000000-0000-0000-0000-000000000103', 'Sales Level 3', 'sales level 3', 3, 'Default editable sales organization role'),
    ('00000000-0000-0000-0000-000000000104', 'Sales Level 4', 'sales level 4', 4, 'Default editable sales organization role')
ON CONFLICT ("id") DO NOTHING;
