package middleware

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"reflect"
	"testing"
)

func TestStaticDataPolicy_Resolve_Success(t *testing.T) {
	group1 := PrivacyGroup{
		name:    "Group1",
		members: map[string]bool{"alice": true},
	}

	group2 := PrivacyGroup{
		name:    "Group2",
		members: map[string]bool{"alice": true},
	}

	transforms := DataTransforms{
		&group1: {
			ExcludedCols: map[string][]string{"table1": {"col1", "col2", "col3"}},
			TableTransforms: map[string]TableTransform{
				"table1": {"col1": func(i interface{}) (interface{}, bool, error) { return i, true, nil }},
			},
		},
		&group2: {
			ExcludedCols:    map[string][]string{"table1": {"col1", "col3", "col4", "col5"}},
			TableTransforms: map[string]TableTransform{},
		},
	}

	dataPolicy := StaticDataPolicy{
		privacyGroups: []*PrivacyGroup{
			&group1,
			&group2,
		},
		transforms: transforms,
	}

	policy, err := dataPolicy.Resolve("alice")
	require.NoError(t, err)

	assert.True(t, reflect.DeepEqual(policy.ExcludedCols["table1"], []string{"col1", "col2", "col3", "col4", "col5"}))

	tableTransform, ok := policy.TableTransforms["table1"]
	assert.True(t, ok)

	transform, ok := tableTransform["col1"]
	assert.True(t, ok)

	transform1, _, err := transform(1)
	require.NoError(t, err)
	assert.True(t, transform1 == 1)
}

func TestStaticDataPolicy_Resolve_Fail(t *testing.T) {
	group1 := PrivacyGroup{
		name:    "Group1",
		members: map[string]bool{"alice": true},
	}

	group2 := PrivacyGroup{
		name:    "Group2",
		members: map[string]bool{"alice": true},
	}

	transforms := DataTransforms{
		&group1: {
			ExcludedCols: map[string][]string{"table1": {"col1", "col2", "col3"}},
			TableTransforms: map[string]TableTransform{
				"table1": {"col1": func(i interface{}) (interface{}, bool, error) { return i, true, nil }},
			},
		},
		&group2: {
			ExcludedCols: map[string][]string{"table1": {"col1", "col3", "col4", "col5"}},
			TableTransforms: map[string]TableTransform{
				"table1": {"col1": func(i interface{}) (interface{}, bool, error) { return i, true, nil }},
			},
		},
	}

	dataPolicy := StaticDataPolicy{
		privacyGroups: []*PrivacyGroup{
			&group1,
			&group2,
		},
		transforms: transforms,
	}

	_, err := dataPolicy.Resolve("alice")
	require.EqualError(t, err, "multiple data policies with different transforms for the same table apply, cannot resolve")
}
