package middleware

import (
	"net/http"
)

type computationCapability map[ComputationLevel]http.Handler

// StaticComputationPolicy holds a map from http request paths to computation capabilities which dictate which handlers
// can be used for the request. A handler can be specified for returning a full globalResult (CanCompute) or just the raw data
// (RawData)
type StaticComputationPolicy struct {
	capabilities map[string]computationCapability
}

// NewStaticComputationPolicy returns a pointer to an initialised StaticComputationPolicy
func NewStaticComputationPolicy() *StaticComputationPolicy {
	return &StaticComputationPolicy{
		make(map[string]computationCapability),
	}
}

// Register adds a capability for a path at a specific ComputationLevel
func (p *StaticComputationPolicy) Register(path string, level ComputationLevel, handler http.Handler) {
	if p.capabilities[path] == nil {
		p.capabilities[path] = make(computationCapability)
	}
	p.capabilities[path][level] = handler
}

// UnregisterAll removes all capabilities for a path
func (p *StaticComputationPolicy) UnregisterAll(path string) {
	delete(p.capabilities, path)
}

// UnregisterOne removes a capability for a path at a specific computation level
func (p *StaticComputationPolicy) UnregisterOne(path string, level ComputationLevel) {
	capability, ok := p.capabilities[path]
	if ok {
		delete(capability, level)
	}
}

// Resolve takes a path and preferred processing location and returns a handler and the computation level which that
// handler provides. It does this based on the capabilities for this path registered with the StaticComputationPolicy.
// The preferred processing location is used to break ties when we can offer full computation and raw data.
func (p *StaticComputationPolicy) Resolve(path string, preferredLocation ProcessingLocation) (ComputationLevel, http.Handler) {
	capability, ok := p.capabilities[path]
	if ok {
		rawDataHandler, hasRawDataHandler := capability[RawData]
		canComputeHandler, hasCanComputeHandler := capability[CanCompute]

		if hasCanComputeHandler {
			if hasRawDataHandler {
				if preferredLocation == Remote {
					return CanCompute, canComputeHandler
				} else {
					return RawData, rawDataHandler
				}
			} else {
				return CanCompute, canComputeHandler
			}
		} else {
			if hasRawDataHandler {
				return RawData, rawDataHandler
			}
		}

	}
	// Default to no capabilities (and so nil function reference)
	return NoComputation, nil
}
