package domain

import (
	"errors"
)

// Role represents a user role for access control.
type Role struct {
	id          string
	name        string
	description string
}

// NewRole creates a new Role instance.
func NewRole(id, name, description string) (*Role, error) {
	if id == "" {
		return nil, errors.New("role id must not be empty")
	}
	if name == "" {
		return nil, errors.New("role name must not be empty")
	}
	if !isValidRoleName(name) {
		return nil, errors.New("invalid role name")
	}

	return &Role{
		id:          id,
		name:        name,
		description: description,
	}, nil
}

// ID returns the role's unique identifier.
func (r *Role) ID() string {
	return r.id
}

// Name returns the role's name.
func (r *Role) Name() string {
	return r.name
}

// Description returns the role's description.
func (r *Role) Description() string {
	return r.description
}

// isValidRoleName checks if the role name is one of the predefined valid roles.
func isValidRoleName(name string) bool {
	validRoles := map[string]bool{
		"ADMIN":     true,
		"MODERATOR": true,
		"STREAMER":  true,
		"USER":      true,
	}
	return validRoles[name]
}
