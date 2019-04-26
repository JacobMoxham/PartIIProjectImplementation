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

// TODO: concurrency protection test?
