package actions

import "simulator/models/capabilities"

type ActionDefinition struct {
	Type          ActionType        `json:"type"`
	Name          string            `json:"name"`
	Description   string            `json:"description"`
	Parameters    []ActionParameter `json:"parameters"`
	Tags          []string          `json:"tags,omitempty"`
	AllowedLabels capabilities.Set  `json:"allowed_labels,omitempty"`
	DeniedLabels  capabilities.Set  `json:"denied_labels,omitempty"`
}

type ActionParameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}
