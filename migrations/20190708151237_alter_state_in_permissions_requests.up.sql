DO $$ BEGIN
  CREATE TYPE permission_request_state AS ENUM ('open', 'granted', 'denied');
EXCEPTION WHEN duplicate_object THEN null;
END $$;
ALTER TABLE permissions_requests DROP COLUMN state;
ALTER TABLE permissions_requests ADD COLUMN state permission_request_state NOT NULL DEFAULT 'open';
