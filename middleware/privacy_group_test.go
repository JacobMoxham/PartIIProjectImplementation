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
	require.False(t, pg.contains("alice"))

	pg.Add("alice")
	require.True(t, pg.contains("alice"))
}

func TestPrivacyGroup_AddMany(t *testing.T) {
	pg := NewPrivacyGroup("g1")
	require.False(t, pg.contains("alice"))

	pg.AddMany([]string{"alice1", "alice2"})
	require.True(t, pg.contains("alice1"))
	require.True(t, pg.contains("alice2"))
}

func TestPrivacyGroup_Remove(t *testing.T) {
	pg := NewPrivacyGroup("g1")
	pg.Add("alice")
	require.True(t, pg.contains("alice"))

	pg.Remove("alice")
	require.False(t, pg.contains("alice"))
}
