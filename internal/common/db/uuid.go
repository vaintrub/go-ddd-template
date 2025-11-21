package db

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// UUIDToPgtype converts a github.com/google/uuid.UUID to pgtype.UUID
func UUIDToPgtype(u uuid.UUID) pgtype.UUID {
	return pgtype.UUID{
		Bytes: u,
		Valid: true,
	}
}

// PgtypeToUUID converts a pgtype.UUID to github.com/google/uuid.UUID
func PgtypeToUUID(u pgtype.UUID) uuid.UUID {
	if !u.Valid {
		return uuid.Nil
	}
	return u.Bytes
}

// StringToPgtypeUUID converts a UUID string to pgtype.UUID
func StringToPgtypeUUID(s string) (pgtype.UUID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return UUIDToPgtype(u), nil
}

// MustParsePgtypeUUID converts a UUID string to pgtype.UUID and panics on error.
// Use this only when you are certain the string is a valid UUID.
func MustParsePgtypeUUID(s string) pgtype.UUID {
	u, err := StringToPgtypeUUID(s)
	if err != nil {
		panic("invalid UUID: " + s + " - " + err.Error())
	}
	return u
}
