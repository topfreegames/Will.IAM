CREATE UNIQUE INDEX permissions_requests_open_unique ON permissions_requests (service, ownership_level, action, resource_hierarchy, service_account_id) WHERE state = 'open';
