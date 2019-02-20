package middleware

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestStaticDataPolicy_Resolve_Success(t *testing.T) {
	group1 := PrivacyGroup{
		Name:    "Group1",
		Members: map[string]bool{"jacob": true},
	}

	group2 := PrivacyGroup{
		Name:    "Group2",
		Members: map[string]bool{"jacob": true},
	}

	transforms := DataTransforms{
		&group1: {
			ExcludedCols: map[string][]string{"table1": {"col1", "col2", "col3"}},
			TableTransforms: map[string]map[string]func(interface{}) (interface{}, error){
				"table1": {"col1": func(i interface{}) (interface{}, error) { return i, nil }},
			},
		},
		&group2: {
			ExcludedCols:    map[string][]string{"table1": {"col1", "col3", "col4", "col5"}},
			TableTransforms: map[string]map[string]func(interface{}) (interface{}, error){},
		},
	}

	dataPolicy := StaticDataPolicy{
		PrivacyGroups: []*PrivacyGroup{
			&group1,
			&group2,
		},
		Transforms: transforms,
	}

	policy, err := dataPolicy.Resolve("jacob")
	require.NoError(t, err)

	assert.True(t, reflect.DeepEqual(policy.ExcludedCols["table1"], []string{"col1", "col2", "col3", "col4", "col5"}))

	tableTransform, ok := policy.TableTransforms["table1"]
	assert.True(t, ok)

	transform, ok := tableTransform["col1"]
	assert.True(t, ok)

	transform1, err := transform(1)
	require.NoError(t, err)
	assert.True(t, transform1 == 1)
}

func TestStaticDataPolicy_Resolve_Fail(t *testing.T) {
	group1 := PrivacyGroup{
		Name:    "Group1",
		Members: map[string]bool{"jacob": true},
	}

	group2 := PrivacyGroup{
		Name:    "Group2",
		Members: map[string]bool{"jacob": true},
	}

	transforms := DataTransforms{
		&group1: {
			ExcludedCols: map[string][]string{"table1": {"col1", "col2", "col3"}},
			TableTransforms: map[string]map[string]func(interface{}) (interface{}, error){
				"table1": {"col1": func(i interface{}) (interface{}, error) { return i, nil }},
			},
		},
		&group2: {
			ExcludedCols: map[string][]string{"table1": {"col1", "col3", "col4", "col5"}},
			TableTransforms: map[string]map[string]func(interface{}) (interface{}, error){
				"table1": {"col1": func(i interface{}) (interface{}, error) { return i, nil }},
			},
		},
	}

	dataPolicy := StaticDataPolicy{
		PrivacyGroups: []*PrivacyGroup{
			&group1,
			&group2,
		},
		Transforms: transforms,
	}

	_, err := dataPolicy.Resolve("jacob")
	require.EqualError(t, err, "multiple data policies with different transforms for the same table apply, cannot resolve")
}

// TODO: add test for isTransformValid
