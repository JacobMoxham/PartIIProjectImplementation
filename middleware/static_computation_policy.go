package middleware

import (
	"net/http"
)

type computationCapability map[ComputationLevel]http.Handler

// StaticComputationPolicy holds a set of computation capabilities for paths, these must be set manually
type StaticComputationPolicy struct {
	capabilities map[string]computationCapability
}

func NewStaticComputationPolicy() *StaticComputationPolicy {
	return &StaticComputationPolicy{
		make(map[string]computationCapability),
	}
}

func (p *StaticComputationPolicy) Register(path string, level ComputationLevel, handler http.Handler) {
	if p.capabilities[path] == nil {
		p.capabilities[path] = make(computationCapability)
	}
	p.capabilities[path][level] = handler
}

func (p *StaticComputationPolicy) UnregisterAll(path string) {
	delete(p.capabilities, path)
}

func (p *StaticComputationPolicy) UnregisterOne(path string, level ComputationLevel) {
	capability, ok := p.capabilities[path]
	if ok {
		delete(capability, level)
	}
}

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
