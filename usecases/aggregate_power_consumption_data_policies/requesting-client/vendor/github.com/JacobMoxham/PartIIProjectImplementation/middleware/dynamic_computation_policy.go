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
		dynamicHandler.Lock()
		defer dynamicHandler.Unlock()
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

// Register adds a capability for a path at a specific ComputationLevel
func (p *DynamicComputationPolicy) Register(path string, level ComputationLevel, handler http.Handler) {
	if p.capabilities[path] == nil {
		p.capabilities[path] = make(dynamicComputationCapability)
	}
	p.capabilities[path][level] = newDynamicHandler(handler)
}

// UnregisterAll removes all capabilities for a path
func (p *DynamicComputationPolicy) UnregisterAll(path string) {
	delete(p.capabilities, path)
}

// UnregisterOne removes a capability for a path at a specific computation level
func (p *DynamicComputationPolicy) UnregisterOne(path string, level ComputationLevel) {
	capability, ok := p.capabilities[path]
	if ok {
		delete(capability, level)
	}
}

// Deactivate marks the handler for a specific request path and computation level as deactivated which means it will
// appear not be registered but can easily be re-activated with a call to Activate
func (p *DynamicComputationPolicy) Deactivate(path string, level ComputationLevel) error {
	capability := p.capabilities[path]
	if capability == nil {
		return fmt.Errorf("no capability was registered for path %s", path)
	}

	dynamicCapacity, ok := capability[level]
	if !ok {
		return fmt.Errorf("no handler was registered for path %s at level %s", path, level.ToString())
	}
	dynamicCapacity.Lock()
	dynamicCapacity.active = false
	dynamicCapacity.Unlock()

	return nil
}

// Activate marks a handler for a specific request path and computation level as active and hence it will appear as
// registered
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

// Resolve takes a path and preferred processing location and returns a handler and the computation level which that
// handler provides. It does this based on the capabilities for this path registered with the StaticComputationPolicy.
// The preferred processing location is used to break ties when we can offer full computation and raw data.
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
