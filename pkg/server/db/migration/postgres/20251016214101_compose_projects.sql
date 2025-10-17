-- +goose Up
-- modify "box" table
ALTER TABLE "box" DROP COLUMN "box_spec";
-- create "box_compose_project" table
CREATE TABLE "box_compose_project" (
  "box_id" bigint NOT NULL,
  "name" text NOT NULL,
  "compose_project" text NOT NULL,
  PRIMARY KEY ("box_id", "name"),
  CONSTRAINT "box_compose_project_box_id_fkey" FOREIGN KEY ("box_id") REFERENCES "box" ("id") ON UPDATE NO ACTION ON DELETE CASCADE
);

-- +goose Down
-- reverse: create "box_compose_project" table
DROP TABLE "box_compose_project";
-- reverse: modify "box" table
ALTER TABLE "box" ADD COLUMN "box_spec" bytea NOT NULL;
