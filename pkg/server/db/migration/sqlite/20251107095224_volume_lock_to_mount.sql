-- +goose Up
-- rename a column from "lock_id" to "mount_id"
ALTER TABLE `volume` RENAME COLUMN `lock_id` TO `mount_id`;
-- rename a column from "lock_time" to "mount_time"
ALTER TABLE `volume` RENAME COLUMN `lock_time` TO `mount_time`;
-- rename a column from "lock_box_id" to "mount_box_id"
ALTER TABLE `volume` RENAME COLUMN `lock_box_id` TO `mount_box_id`;
-- rename a column from "lock_id" to "mount_id"
ALTER TABLE `volume_snapshot` RENAME COLUMN `lock_id` TO `mount_id`;

-- +goose Down
-- reverse: rename a column from "lock_id" to "mount_id"
ALTER TABLE `volume_snapshot` RENAME COLUMN `mount_id` TO `lock_id`;
-- reverse: rename a column from "lock_box_id" to "mount_box_id"
ALTER TABLE `volume` RENAME COLUMN `mount_box_id` TO `lock_box_id`;
-- reverse: rename a column from "lock_time" to "mount_time"
ALTER TABLE `volume` RENAME COLUMN `mount_time` TO `lock_time`;
-- reverse: rename a column from "lock_id" to "mount_id"
ALTER TABLE `volume` RENAME COLUMN `mount_id` TO `lock_id`;
