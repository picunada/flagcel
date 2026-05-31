package core

import "time"

type ContextSchema struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Fields      []ContextField `json:"fields"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	CreatedBy   *string        `json:"created_by,omitempty"`
	DeletedBy   *string        `json:"deleted_by,omitempty"`
}

type ContextField struct {
	Path string      `json:"path"`
	Type ContextType `json:"type"`
}

type ContextType string

const (
	ContextTypeString    ContextType = "string"
	ContextTypeInt       ContextType = "int"
	ContextTypeDouble    ContextType = "double"
	ContextTypeBool      ContextType = "bool"
	ContextTypeTimestamp ContextType = "timestamp"
	ContextTypeList      ContextType = "list"
	ContextTypeMap       ContextType = "map"
)

func (t ContextType) Valid() bool {
	switch t {
	case ContextTypeString, ContextTypeInt, ContextTypeDouble,
		ContextTypeBool, ContextTypeTimestamp, ContextTypeList, ContextTypeMap:
		return true
	}
	return false
}
