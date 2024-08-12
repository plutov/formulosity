package types

import (
	"fmt"
)

type DuplicateProtectionType string

const (
	DuplicateProtectionType_Cookie DuplicateProtectionType = "cookie"
	DuplicateProtectionType_Ip     DuplicateProtectionType = "ip"
)

var supportedDuplicateProtectionTypes = map[DuplicateProtectionType]bool{
	DuplicateProtectionType_Cookie: true,
	DuplicateProtectionType_Ip:     true,
}

type Security struct {
	DuplicateProtection DuplicateProtectionType `json:"duplicateProtection" yaml:"duplicateProtection"`
}

func (s *Security) Validate() error {
	if _, ok := supportedDuplicateProtectionTypes[s.DuplicateProtection]; !ok {
		return fmt.Errorf("security.duplicateProtection is invalid: %s", s.DuplicateProtection)
	}

	return nil
}
