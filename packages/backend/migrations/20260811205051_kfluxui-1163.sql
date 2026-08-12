-- Modify "issue_scopes" table
ALTER TABLE "issue_scopes" ALTER COLUMN "id" DROP DEFAULT;
-- Modify "issues" table
ALTER TABLE "issues" ALTER COLUMN "id" DROP DEFAULT, ADD COLUMN "resolved_by_id" character varying(40) NULL DEFAULT NULL::character varying;
-- Modify "links" table
ALTER TABLE "links" ALTER COLUMN "id" DROP DEFAULT;
-- Modify "related_issues" table
ALTER TABLE "related_issues" ALTER COLUMN "id" DROP DEFAULT;
