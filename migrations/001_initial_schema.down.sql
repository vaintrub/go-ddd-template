-- Rollback Initial Schema
-- Created: 2025-11-18
-- Purpose: Remove all tables created in 001_initial_schema.up.sql

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS trainings_trainings;
DROP TABLE IF EXISTS trainer_hours;
DROP TABLE IF EXISTS users_users;
