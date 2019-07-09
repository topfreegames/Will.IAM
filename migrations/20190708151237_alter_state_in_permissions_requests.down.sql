DROP TYPE permission_request_state;
ALTER TABLE permissions_requests DROP COLUMN state;
ALTER TABLE permissions_requests ADD COLUMN state smallint NOT NULL DEFAULT 0;
