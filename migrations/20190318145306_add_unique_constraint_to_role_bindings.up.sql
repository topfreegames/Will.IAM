CREATE UNIQUE INDEX IF NOT EXISTS role_bindings_unique ON role_bindings (service_account_id, role_id);
