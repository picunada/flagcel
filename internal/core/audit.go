package core

import "time"

// Audit resource types. Only flags are recorded for now, but the audit_logs
// table is generic so other resources can be added without a migration.
const ResourceTypeFlag = "flag"

// Audit actions.
const (
	AuditActionCreated = "created"
	AuditActionUpdated = "updated"
	AuditActionDeleted = "deleted"
)

// AuditEntry is a single version in a resource's change history. Each entry
// stores the full snapshot of the resource as it existed after the change
// (nil for a delete), so consecutive entries can be diffed to show exactly
// what changed.
type AuditEntry struct {
	ID            string
	EnvironmentID string
	ResourceType  string
	ResourceID    string
	Action        string
	Version       int
	Snapshot      *FlagConfig
	ActorID       *string
	ActorLabel    *string
	CreatedAt     time.Time
}
