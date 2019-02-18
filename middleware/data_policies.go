package middleware

import (
	"errors"
	"fmt"
)

// TableOperations contains functions to apply to tables before sending to an entity and columns to exclude
type TableOperations struct {
	TableTransforms map[string]map[string]func(interface{}) (interface{}, error)
	ExcludedCols    map[string][]string
}

func NewTableOperations() *TableOperations {
	return &TableOperations{
		TableTransforms: make(map[string]map[string]func(interface{}) (interface{}, error)),
		ExcludedCols:    make(map[string][]string),
	}
}

func (t *TableOperations) merge(tableOperations *TableOperations) error {
	// Try to merge transforms, error if we have a clash
	for id, transforms := range tableOperations.TableTransforms {
		// Check for a clash
		_, ok := t.TableTransforms[id]
		if ok {
			return errors.New("multiple data policies with different transforms for the same table apply, cannot resolve")
		}
		t.TableTransforms[id] = transforms
	}

	// Merge excluded columns
	for id, excludedCols := range tableOperations.ExcludedCols {
		t.ExcludedCols[id] = mergeStringSlice(t.ExcludedCols[id], excludedCols)
	}

	return nil
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
	PrivacyGroups []*PrivacyGroup
	Transforms    DataTransforms
}

func (stp *StaticDataPolicy) Resolve(entityID string) (*TableOperations, error) {
	var privacyGroups []*PrivacyGroup
	for _, group := range stp.PrivacyGroups {
		if group.contains(entityID) {
			privacyGroups = append(privacyGroups, group)
		}
	}
	if privacyGroups == nil {
		return nil, fmt.Errorf("the entity %s is not part of any privacy group", entityID)
	}

	// Make sure we only have one set of transforms but concatenate removed columns
	allTableOperations := NewTableOperations()
	for _, privacyGroup := range privacyGroups {
		tableOperations, ok := stp.Transforms[privacyGroup]
		if ok {
			err := allTableOperations.merge(tableOperations)
			if err != nil {
				if err != nil {
					return nil, err
				}
			}
		}
	}

	return allTableOperations, nil
}
