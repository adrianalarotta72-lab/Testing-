package api

// Request represents a KV store request
type Request struct {
	Operation string `json:"operation"` // SET, UPDATE, GET, DELETE, EXISTS, SIZE, CLEAR
	Key       string `json:"key"`
	Value     string `json:"value,omitempty"` // JSON string que puede contener cualquier tipo
}

// Response represents a KV store response
type Response struct {
	Success bool   `json:"success"`
	Value   string `json:"value,omitempty"` // JSON string
	Error   string `json:"error,omitempty"`
	Size    int    `json:"size,omitempty"`
}

// Operation constants
const (
	OpSet    = "SET"
	OpUpdate = "UPDATE"
	OpGet    = "GET"
	OpDelete = "DELETE"
	OpExists = "EXISTS"
	OpSize   = "SIZE"
	OpClear  = "CLEAR"
)
