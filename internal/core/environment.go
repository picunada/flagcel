package core

import "time"

const DefaultEnvironmentKey = "production"

type Environment struct {
	ID          string
	Key         string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	CreatedBy   *string
	DeletedBy   *string
}
