-- Critical Seed Data for Attendance Management System

-- 1. Roles
INSERT INTO roles (name, slug, created_at, updated_at) VALUES 
('Admin', 'admin', NOW(), NOW()),
('Manager', 'manager', NOW(), NOW()),
('Employee', 'employee', NOW(), NOW()),
('Supervisor', 'supervisor', NOW(), NOW());

-- 2. Initial Admin User
-- Credentials from .env: admin / admin@email.com / adminpassword
INSERT INTO users (name, email, password, role_id, created_at, updated_at) VALUES 
('admin', 'admin@email.com', '$2y$10$WC69pWT6li1GFznHK6lNcecosXOapH/9Eye108Tw5xmZWTU2NfYp2', (SELECT id FROM roles WHERE slug = 'admin'), NOW(), NOW());

-- =========================================================
-- RE-SYNCHRONIZE SEQUENCES (PREVENTS PKEY 23505 COLLISION ERRORS)
-- Run these after any manual INSERTs in your DB or seeders.
-- =========================================================
SELECT setval('roles_id_seq', COALESCE((SELECT MAX(id)+1 FROM roles), 1), false);
SELECT setval('users_id_seq', COALESCE((SELECT MAX(id)+1 FROM users), 1), false);
SELECT setval('work_centers_id_seq', COALESCE((SELECT MAX(id)+1 FROM work_centers), 1), false);
SELECT setval('positions_id_seq', COALESCE((SELECT MAX(id)+1 FROM positions), 1), false);
SELECT setval('work_shifts_id_seq', COALESCE((SELECT MAX(id)+1 FROM work_shifts), 1), false);
SELECT setval('employees_id_seq', COALESCE((SELECT MAX(id)+1 FROM employees), 1), false);
SELECT setval('attendances_id_seq', COALESCE((SELECT MAX(id)+1 FROM attendances), 1), false);
SELECT setval('incidents_id_seq', COALESCE((SELECT MAX(id)+1 FROM incidents), 1), false);
SELECT setval('reports_id_seq', COALESCE((SELECT MAX(id)+1 FROM reports), 1), false);
SELECT setval('holidays_id_seq', COALESCE((SELECT MAX(id)+1 FROM holidays), 1), false);
