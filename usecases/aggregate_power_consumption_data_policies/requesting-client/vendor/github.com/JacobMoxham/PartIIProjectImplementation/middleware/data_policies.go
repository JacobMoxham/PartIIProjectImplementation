package middleware

import (
	"fmt"
	"log"
)

// TableOperations contains functions to apply to tables before sending to an entity and columns to exclude
type TableOperations struct {
	TableTransforms map[string]func(interface{}) (interface{}, error)
	ExcludedCols    map[string][]string
}

// Transforms is a map from PrivacyGroups to TableOperations
type DataTransforms map[*PrivacyGroup]*TableOperations

// DataPolicy allow us to get a function which must be applied to data before returning for a given identifier
type DataPolicy interface {
	// resolve takes an identifier for an entity and returns the TableOperatoinsg for the entity
	Resolve(string) (*TableOperations, error)
}

type StaticDataPolicy struct {
	// PrivacyGroups is an ordered list of PrivacyGroups where the policy for the first group we are a member of is applied
	// TODO: talk to Jat about whether to make this more complicated and try to incorporate all policies
	PrivacyGroups []*PrivacyGroup
	Transforms    DataTransforms
}

func (stp *StaticDataPolicy) Resolve(entityID string) (*TableOperations, error) {
	var privacyGroup *PrivacyGroup
	for _, group := range stp.PrivacyGroups {
		if group.contains(entityID) {
			privacyGroup = group
			break
		}
	}
	if privacyGroup == nil {
		return nil, fmt.Errorf("the entity %s is not part of any privacy group", entityID)
	}
	tableOperations, ok := stp.Transforms[privacyGroup]
	if !ok {
		// If the group has no table operations then allow access to the full table
		log.Printf("Group %s has no data transforms, allowing full access", privacyGroup.Name)
		return new(TableOperations), nil
	}
	return tableOperations, nil
}
