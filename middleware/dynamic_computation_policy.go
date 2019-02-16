package middleware

import (
	"fmt"
	"net/http"
	"sync"
)

type dynamicHandler struct {
	sync.Mutex
	handler http.Handler
	active  bool
}

func newDynamicHandler(handler http.Handler) *dynamicHandler {
	return &dynamicHandler{
		handler: handler,
		active:  true,
	}
}

type dynamicComputationCapability map[ComputationLevel]*dynamicHandler

func (dcc dynamicComputationCapability) Get(computationLevel ComputationLevel) (http.Handler, bool) {
	dynamicHandler, ok := dcc[computationLevel]

	// Check there was a handler registered
	if ok {
		// Check the handler is active
		if dynamicHandler.active {
			return dynamicHandler.handler, ok
		}
	}

	return nil, false
}

// DynamicComputationPolicy holds a set of computation capabilities for paths, these must be set manually
type DynamicComputationPolicy struct {
	capabilities map[string]dynamicComputationCapability
}

// NewDynamicComputationPolicy returns a pointer to a DynamicComputationPolicy with an empty, initialised internal map
func NewDynamicComputationPolicy() *DynamicComputationPolicy {
	return &DynamicComputationPolicy{
		make(map[string]dynamicComputationCapability),
	}
}

func (p *DynamicComputationPolicy) Register(path string, level ComputationLevel, handler http.Handler) {
	if p.capabilities[path] == nil {
		p.capabilities[path] = make(dynamicComputationCapability)
	}
	p.capabilities[path][level] = newDynamicHandler(handler)
}

func (p *DynamicComputationPolicy) UnregisterAll(path string) {
	delete(p.capabilities, path)
}

func (p *DynamicComputationPolicy) UnregisterOne(path string, level ComputationLevel) {
	capability, ok := p.capabilities[path]
	if ok {
		delete(capability, level)
	}
}

func (p *DynamicComputationPolicy) Deactivate(path string, level ComputationLevel) error {
	capability := p.capabilities[path]
	if capability == nil {
		return fmt.Errorf("no capability was registered for path %s", path)
	}

	dynamicCapacity, ok := capability[level]
	if !ok {
		return fmt.Errorf("no capacity was registered for path %s at level %s", path, level.ToString())
	}
	dynamicCapacity.Lock()
	dynamicCapacity.active = false
	dynamicCapacity.Unlock()

	return nil
}

func (p *DynamicComputationPolicy) Activate(path string, level ComputationLevel) error {
	capability := p.capabilities[path]
	if capability == nil {
		return fmt.Errorf("no capability was registered for path %s", path)
	}

	dynamicCapacity, ok := capability[level]
	if !ok {
		return fmt.Errorf("no capacity was registered for path %s at level %s", path, level.ToString())
	}
	dynamicCapacity.Lock()
	dynamicCapacity.active = true
	dynamicCapacity.Unlock()

	return nil
}

func (p *DynamicComputationPolicy) Resolve(path string, preferredLocation ProcessingLocation) (ComputationLevel, http.Handler) {
	capability, ok := p.capabilities[path]
	if ok {
		rawDataHandler, hasRawDataHandler := capability.Get(RawData)
		canComputeHandler, hasCanComputeHandler := capability.Get(CanCompute)

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
