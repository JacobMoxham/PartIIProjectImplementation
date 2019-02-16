package middleware

import (
	"log"
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

		// TODO: in dynamic version we may have a "valid" tag?
		if hasCanComputeHandler {
			if hasRawDataHandler {
				if preferredLocation == Remote {
					log.Println("Serving request as preferred location is remote and we can compute")
					return CanCompute, canComputeHandler
				} else {
					log.Println("Partially serving request as preferred location is local and we can compute")
					return RawData, rawDataHandler
				}
			} else {
				if preferredLocation == Local {
					log.Println("Preferred location is local but we can only compute full result")
				} else {
					log.Println("Serving request as we can compute full result")
				}
				return CanCompute, canComputeHandler
			}
		} else {
			if hasRawDataHandler {
				if preferredLocation == Local {
					log.Println("Partially serving request as preferred location is local and we can compute")
				} else {
					// TODO: ensure receivers handle this correctly
					log.Println("Preferred location is remote but we can only partially compute, returning anyway")
				}
				return RawData, rawDataHandler
			}
		}

	}
	// Default to no capabilities (and so nil function reference)
	return NoComputation, nil
}
