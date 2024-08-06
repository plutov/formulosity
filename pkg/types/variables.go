package types

import (
	"errors"
	"fmt"
)

type VariableType string

const (
	VariableType_List VariableType = "list"
)

var supportedVariableTypes = map[VariableType]bool{
	VariableType_List: true,
}

type Variables struct {
	Variables []Variable `json:"variables" yaml:"variables"`
}

type Variable struct {
	ID      string       `json:"id" yaml:"id"`
	Type    VariableType `json:"type" yaml:"type"`
	Options []string     `json:"options,omitempty" yaml:"options,omitempty"`
}

func (v *Variables) Validate() error {
	uniqueIDs := make(map[string]bool)
	for _, variable := range v.Variables {
		if variable.ID == "" {
			return errors.New("variables[].id is required")
		}
		if _, ok := supportedVariableTypes[variable.Type]; !ok {
			return fmt.Errorf("variables[].type is invalid: %s", variable.Type)
		}

		if _, ok := uniqueIDs[variable.ID]; ok {
			return fmt.Errorf("variables[].id is duplicated: %s", variable.ID)
		}
		uniqueIDs[variable.ID] = true
	}

	return nil
}
