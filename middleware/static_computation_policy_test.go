package middleware

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

var rawDataHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
var canComputeHandler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

func TestStaticComputationPolicy_Register(t *testing.T) {
	compPol := NewStaticComputationPolicy()
	compPol.Register("/", RawData, rawDataHandler)
	compPol.Register("/", CanCompute, canComputeHandler)

	// We just check the levels here to ensure that a handler is registered
	localLevel, _ := compPol.Resolve("/", Local)
	require.Equal(t, RawData, localLevel)

	remoteLevel, _ := compPol.Resolve("/", Remote)
	require.Equal(t, CanCompute, remoteLevel)
}

func TestStaticComputationPolicy_UnregisterAll(t *testing.T) {
	compPol := NewStaticComputationPolicy()
	compPol.Register("/", RawData, rawDataHandler)
	compPol.Register("/", CanCompute, canComputeHandler)

	// We just check the levels here to ensure that a handler is registered
	localLevel, _ := compPol.Resolve("/", Local)
	require.Equal(t, RawData, localLevel)
	remoteLevel, _ := compPol.Resolve("/", Remote)
	require.Equal(t, CanCompute, remoteLevel)

	compPol.UnregisterAll("/")
	localLevel, _ = compPol.Resolve("/", Local)
	require.Equal(t, NoComputation, localLevel)
	remoteLevel, _ = compPol.Resolve("/", Remote)
	require.Equal(t, NoComputation, remoteLevel)
}

func TestStaticComputationPolicy_UnregisterOne_NoComputation(t *testing.T) {
	compPol := NewStaticComputationPolicy()
	compPol.Register("/", RawData, rawDataHandler)

	// We just check the levels here to ensure that a handler is registered
	localLevel, _ := compPol.Resolve("/", Local)
	require.Equal(t, RawData, localLevel)

	compPol.UnregisterOne("/", RawData)
	localLevel, _ = compPol.Resolve("/", Local)
	require.Equal(t, NoComputation, localLevel)
}

func TestStaticComputationPolicy_UnregisterOne_LeavingOther(t *testing.T) {
	compPol := NewStaticComputationPolicy()
	compPol.Register("/", RawData, rawDataHandler)
	compPol.Register("/", CanCompute, canComputeHandler)

	// We just check the levels here to ensure that a handler is registered
	localLevel, _ := compPol.Resolve("/", Local)
	require.Equal(t, RawData, localLevel)

	compPol.UnregisterOne("/", RawData)
	localLevel, _ = compPol.Resolve("/", Local)
	require.Equal(t, CanCompute, localLevel)

	localLevel, _ = compPol.Resolve("/", Remote)
	require.Equal(t, CanCompute, localLevel)
}

// Resolve is tested in http_test.go via the PolicyAwareHandler privacy_aware_client_test.go via the PolicyAwareClient
