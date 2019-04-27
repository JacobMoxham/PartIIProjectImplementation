package middleware

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPrivacyGroup_Name(t *testing.T) {
	pg := NewPrivacyGroup("g1")
	require.Equal(t, "g1", pg.name)
}

func TestPrivacyGroup_Add(t *testing.T) {
	pg := NewPrivacyGroup("g1")
	require.False(t, pg.contains("jacob"))

	pg.Add("jacob")
	require.True(t, pg.contains("jacob"))
}

func TestPrivacyGroup_AddMany(t *testing.T) {
	pg := NewPrivacyGroup("g1")
	require.False(t, pg.contains("jacob"))

	pg.AddMany([]string{"jacob1", "jacob2"})
	require.True(t, pg.contains("jacob1"))
	require.True(t, pg.contains("jacob2"))
}

func TestPrivacyGroup_Remove(t *testing.T) {
	pg := NewPrivacyGroup("g1")
	pg.Add("jacob")
	require.True(t, pg.contains("jacob"))

	pg.Remove("jacob")
	require.False(t, pg.contains("jacob"))
}
