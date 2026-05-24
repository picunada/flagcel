package core

type ContextSchema struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Fields      []ContextField `json:"fields"`
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
