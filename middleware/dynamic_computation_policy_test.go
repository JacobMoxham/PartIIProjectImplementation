package middleware

import (
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func handler(w http.ResponseWriter, r *http.Request) {
}

func TestDynamicComputationPolicy_Register(t *testing.T) {
	dynamicComputationPolicy := NewDynamicComputationPolicy()
	handlerFunc := http.HandlerFunc(handler)
	dynamicComputationPolicy.Register("/", CanCompute, handlerFunc)

	computationLevel, _ := dynamicComputationPolicy.Resolve("/", Local)
	require.Equal(t, computationLevel, CanCompute)

}

func TestDynamicComputationPolicy_Deactivate(t *testing.T) {
	dynamicComputationPolicy := NewDynamicComputationPolicy()
	handlerFunc := http.HandlerFunc(handler)
	dynamicComputationPolicy.Register("/", CanCompute, handlerFunc)
	err := dynamicComputationPolicy.Deactivate("/", CanCompute)
	require.NoError(t, err)

	computationLevel, handler := dynamicComputationPolicy.Resolve("/", Local)
	require.Equal(t, computationLevel, NoComputation)
	require.Equal(t, handler, nil)
}

func TestDynamicComputationPolicy_Activate(t *testing.T) {
	dynamicComputationPolicy := NewDynamicComputationPolicy()
	handlerFunc := http.HandlerFunc(handler)
	dynamicComputationPolicy.Register("/", CanCompute, handlerFunc)
	dynamicComputationPolicy.Deactivate("/", CanCompute)
	err := dynamicComputationPolicy.Activate("/", CanCompute)
	require.NoError(t, err)

	computationLevel, _ := dynamicComputationPolicy.Resolve("/", Local)
	require.Equal(t, computationLevel, CanCompute)
}

func TestDynamicComputationPolicy_UnregisterAll(t *testing.T) {
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

func TestDynamicComputationPolicy_UnregisterOne_NoComputation(t *testing.T) {
	compPol := NewDynamicComputationPolicy()
	compPol.Register("/", RawData, rawDataHandler)

	// We just check the levels here to ensure that a handler is registered
	localLevel, _ := compPol.Resolve("/", Local)
	require.Equal(t, RawData, localLevel)

	compPol.UnregisterOne("/", RawData)
	localLevel, _ = compPol.Resolve("/", Local)
	require.Equal(t, NoComputation, localLevel)
}

func TestDynamicComputationPolicy_UnregisterOne_LeavingOther(t *testing.T) {
	compPol := NewDynamicComputationPolicy()
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
