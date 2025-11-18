-- Initial Schema for go-ddd-template
-- Created: 2025-11-18
-- Purpose: Initialize database schema for trainings, users, and trainer bounded contexts

-- ============================================================================
-- Trainings Context
-- ============================================================================

CREATE TABLE trainings_trainings (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    user_name TEXT NOT NULL,
    training_time TIMESTAMP WITH TIME ZONE NOT NULL,
    notes TEXT,
    proposed_new_time TIMESTAMP WITH TIME ZONE,
    move_proposed_by TEXT,
    canceled BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT notes_length_check CHECK (LENGTH(notes) <= 1000),
    CONSTRAINT move_proposed_by_check CHECK (move_proposed_by IN ('trainer', 'attendee', NULL))
);

-- Indexes for common query patterns
CREATE INDEX trainings_trainings_user_id_idx ON trainings_trainings(user_id);
CREATE INDEX trainings_trainings_time_idx ON trainings_trainings(training_time);
CREATE INDEX trainings_trainings_created_at_id_idx ON trainings_trainings(created_at, id);

-- Comments for documentation
COMMENT ON TABLE trainings_trainings IS 'Training sessions scheduled by users with trainers';
COMMENT ON COLUMN trainings_trainings.id IS 'Training unique identifier (domain UUID)';
COMMENT ON COLUMN trainings_trainings.user_id IS 'User who booked the training';
COMMENT ON COLUMN trainings_trainings.user_name IS 'User name snapshot at booking time';
COMMENT ON COLUMN trainings_trainings.notes IS 'Training notes, max 1000 characters';
COMMENT ON COLUMN trainings_trainings.move_proposed_by IS 'Who proposed reschedule: trainer or attendee';
COMMENT ON COLUMN trainings_trainings.created_at IS 'Record creation timestamp for auditing and pagination';

-- ============================================================================
-- Trainer Context
-- ============================================================================

CREATE TABLE trainer_hours (
    id UUID PRIMARY KEY,
    hour_time TIMESTAMP WITH TIME ZONE NOT NULL UNIQUE,
    availability TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT availability_check CHECK (availability IN ('available', 'not_available', 'training_scheduled'))
);

-- Indexes for common query patterns
CREATE INDEX trainer_hours_time_idx ON trainer_hours(hour_time);
CREATE INDEX trainer_hours_created_at_id_idx ON trainer_hours(created_at, id);

-- Comments for documentation
COMMENT ON TABLE trainer_hours IS 'Trainer availability hours for scheduling';
COMMENT ON COLUMN trainer_hours.id IS 'Hour unique identifier';
COMMENT ON COLUMN trainer_hours.hour_time IS 'The specific hour (truncated to hour boundary)';
COMMENT ON COLUMN trainer_hours.availability IS 'Availability status: available, not_available, or training_scheduled';
COMMENT ON COLUMN trainer_hours.created_at IS 'Record creation timestamp for auditing and pagination';

-- ============================================================================
-- Users Context
-- ============================================================================

CREATE TABLE users_users (
    id UUID PRIMARY KEY,
    user_type TEXT NOT NULL,
    name TEXT NOT NULL,
    email TEXT UNIQUE,
    balance INTEGER NOT NULL DEFAULT 0,
    last_ip TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    -- Constraints
    CONSTRAINT user_type_check CHECK (user_type IN ('trainer', 'attendee'))
);

-- Indexes for common query patterns
CREATE INDEX users_users_email_idx ON users_users(email) WHERE email IS NOT NULL;
CREATE INDEX users_users_user_type_idx ON users_users(user_type);
CREATE INDEX users_users_created_at_id_idx ON users_users(created_at, id);

-- Comments for documentation
COMMENT ON TABLE users_users IS 'User accounts with roles (trainer or attendee)';
COMMENT ON COLUMN users_users.id IS 'User unique identifier';
COMMENT ON COLUMN users_users.user_type IS 'User role: trainer or attendee';
COMMENT ON COLUMN users_users.name IS 'User full name';
COMMENT ON COLUMN users_users.email IS 'User email address (optional, for future authentication)';
COMMENT ON COLUMN users_users.created_at IS 'Record creation timestamp for auditing and pagination';
