package middleware

import (
	"errors"
	"fmt"
	"time"
)

// TableOperations contains functions to apply to tables before sending to an entity and columns to exclude
type TableOperations struct {
	TableTransforms map[string]map[string]func(interface{}) (interface{}, error)
	ExcludedCols    map[string][]string
}

// NewTableOperations returns a pointer to a TableOperations struct with initialised fields
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

// transforms is a map from privacyGroups to TableOperations
type DataTransforms map[*PrivacyGroup]*TableOperations

// DataPolicy allow us to get a function which must be applied to data before returning for a given identifier
type DataPolicy interface {
	// resolve takes an identifier for an entity and returns the TableOperatoinsg for the entity
	Resolve(string) (*TableOperations, error)
	LastUpdated() time.Time
}

// StaticDataPolicy implements the DataPolicy interface and contains a list of privacyGroups and DataTransforms for them
type StaticDataPolicy struct {
	// privacyGroups is an ordered list of privacyGroups where the policy for the first group we are a member of is applied
	privacyGroups []*PrivacyGroup
	transforms    DataTransforms
	created       time.Time
}

// NewStaticDataPolicy returns a pointer to a StaticDataPolicy with initialised fields
func NewStaticDataPolicy(privacyGroups []*PrivacyGroup, transforms DataTransforms) *StaticDataPolicy {
	return &StaticDataPolicy{
		privacyGroups: privacyGroups,
		transforms:    transforms,
		created:       time.Now(),
	}
}

// Resolve takes an entity ID and returns a pointer to the relevant TableOperations struct based on the privacyGroups
// that the entity ID is in and the associated transforms stored in the StaticDataPolicy
func (sdp *StaticDataPolicy) Resolve(entityID string) (*TableOperations, error) {
	var privacyGroups []*PrivacyGroup
	for _, group := range sdp.privacyGroups {
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
		tableOperations, ok := sdp.transforms[privacyGroup]
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

func (sdp StaticDataPolicy) LastUpdated() time.Time {
	// The static policy is not intended to be updated
	return sdp.created
}
