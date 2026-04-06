package resource

import "time"

type ResourceJSONOutput struct {
	ID                *string           `json:"id,omitempty"`
	FolderParentID    *string           `json:"folder_parent_id,omitempty"`
	Name              *string           `json:"name,omitempty"`
	Username          *string           `json:"username,omitempty"`
	URI               *string           `json:"uri,omitempty"`
	Password          *string           `json:"password,omitempty"`
	Description       *string           `json:"description,omitempty"`
	CustomFields      []CustomFieldJSON `json:"custom_fields,omitempty"`
	CreatedTimestamp  *time.Time        `json:"created_timestamp,omitempty"`
	ModifiedTimestamp *time.Time        `json:"modified_timestamp,omitempty"`
}

type CustomFieldJSON struct {
	ID          string      `json:"id"`
	Type        string      `json:"type"`
	Key         string      `json:"key,omitempty"`
	SecretValue interface{} `json:"secret_value,omitempty"`
}
