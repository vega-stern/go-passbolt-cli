package util

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/passbolt/go-passbolt/api"
	"github.com/passbolt/go-passbolt/helper"
)

// CustomField is a merged view of a v5-custom-fields entry.
// MetadataKey is the human label (resource metadata), SecretValue is the actual secret value.
type CustomField struct {
	ID          string
	Type        string
	MetadataKey string
	SecretKey   string
	SecretValue interface{}
}

type v5CustomFieldsMetadata struct {
	Name         string            `json:"name,omitempty"`
	URIs         []string          `json:"uris,omitempty"`
	Description  string            `json:"description,omitempty"`
	CustomFields []v5CustomFieldMD `json:"custom_fields,omitempty"`
}

type v5CustomFieldMD struct {
	ID          string  `json:"id"`
	Type        string  `json:"type"`
	MetadataKey *string `json:"metadata_key,omitempty"`
}

type v5CustomFieldsSecret struct {
	CustomFields []v5CustomFieldSecret `json:"custom_fields,omitempty"`
}

type v5CustomFieldSecret struct {
	ID          string      `json:"id"`
	Type        string      `json:"type"`
	SecretKey   *string     `json:"secret_key,omitempty"`
	SecretValue interface{} `json:"secret_value,omitempty"`
}

// DecryptCustomFieldsResource decrypts a v5-custom-fields resource into best-effort top-level
// fields plus the full list of custom fields.
func DecryptCustomFieldsResource(ctx context.Context, c *api.Client, resource api.Resource, secret api.Secret, rType api.ResourceType) (name, username, uri, password, description string, fields []CustomField, err error) {
	if rType.Slug != "v5-custom-fields" {
		return "", "", "", "", "", nil, fmt.Errorf("not a v5-custom-fields resource")
	}

	rawMetadata, err := helper.GetResourceMetadata(ctx, c, &resource, &rType)
	if err != nil {
		return "", "", "", "", "", nil, fmt.Errorf("getting metadata: %w", err)
	}
	var md v5CustomFieldsMetadata
	if err := json.Unmarshal([]byte(rawMetadata), &md); err != nil {
		return "", "", "", "", "", nil, fmt.Errorf("parsing metadata: %w", err)
	}

	name = md.Name
	description = md.Description
	if len(md.URIs) > 0 {
		uri = md.URIs[0]
	}

	// Merge field labels (metadata) with secret values.
	idx := map[string]*CustomField{}
	fields = make([]CustomField, 0, len(md.CustomFields))
	for _, f := range md.CustomFields {
		cf := CustomField{ID: f.ID, Type: f.Type}
		if f.MetadataKey != nil {
			cf.MetadataKey = strings.TrimSpace(*f.MetadataKey)
		}
		fields = append(fields, cf)
		idx[f.ID] = &fields[len(fields)-1]
	}

	if secret.Data != "" {
		rawSecret, err := c.DecryptSecretWithResourceID(resource.ID, secret.Data)
		if err != nil {
			return "", "", "", "", "", nil, fmt.Errorf("decrypting secret: %w", err)
		}
		var sd v5CustomFieldsSecret
		if err := json.Unmarshal([]byte(rawSecret), &sd); err != nil {
			return "", "", "", "", "", nil, fmt.Errorf("parsing secret: %w", err)
		}
		for _, s := range sd.CustomFields {
			cf, ok := idx[s.ID]
			if !ok {
				fields = append(fields, CustomField{ID: s.ID, Type: s.Type})
				cf = &fields[len(fields)-1]
				idx[s.ID] = cf
			}
			cf.Type = s.Type
			if s.SecretKey != nil {
				cf.SecretKey = strings.TrimSpace(*s.SecretKey)
			}
			cf.SecretValue = s.SecretValue
		}
	}

	// Best-effort extraction for convenience (used by exec).
	username = findCustomFieldString(fields, []string{"user", "username", "login", "email"})
	password = findCustomFieldString(fields, []string{"password", "pass", "passwd"})

	return name, username, uri, password, description, fields, nil
}

func findCustomFieldString(fields []CustomField, keys []string) string {
	keyset := map[string]struct{}{}
	for _, k := range keys {
		keyset[k] = struct{}{}
	}
	for _, f := range fields {
		k := strings.ToLower(strings.TrimSpace(f.MetadataKey))
		if k == "" {
			k = strings.ToLower(strings.TrimSpace(f.SecretKey))
		}
		if _, ok := keyset[k]; !ok {
			continue
		}
		if s, ok := f.SecretValue.(string); ok {
			return s
		}
	}
	return ""
}
