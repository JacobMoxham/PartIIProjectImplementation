package middleware

import (
	"net/url"
)

// DataTransform is a map from PrivacyGroups to functions which must be applied to data before returning
type DataTransform map[PrivacyGroup]func(interface{}) interface{}

// PrivacyGroup is an interface for structs which contain a data structure of ID's which we can add to and remove from
type PrivacyGroup interface {
	Add(string)
	Remove(string) error
}

// DataPolicy allow us to get a function which must be applied to data before returning for a given identifier
type DataPolicy interface {
	// resolve takes an indentififer and returns the function to be applied to data before returning
	resolve(string) func(interface{}) interface{}
}

type ComputationCapability int

const (
	NoComputation ComputationCapability = iota
	RawData       ComputationCapability = iota
	CanCompute    ComputationCapability = iota
)

// ComputationPolicy stores computation capabilities of a node
type ComputationPolicy interface {
	Register(string, ComputationCapability)
	Unregister(string)
	Resolve(string) ComputationCapability
}

// StaticComputationPolicy holds a set of computation capabilities for paths, these must be set manually
// TODO: make safe for concurrent use
type StaticComputationPolicy struct {
	capabilities map[string]ComputationCapability
}

func NewStaticComputationPolicy() *StaticComputationPolicy {
	return &StaticComputationPolicy{
		make(map[string]ComputationCapability),
	}
}

func (p *StaticComputationPolicy) Register(handler string, capability ComputationCapability) {
	p.capabilities[handler] = capability
}

func (p *StaticComputationPolicy) Unregister(handler string) {
	delete(p.capabilities, handler)
}

func (p *StaticComputationPolicy) Resolve(handler string) ComputationCapability {
	c, ok := p.capabilities[handler]
	if ok {
		return c
	} else {
		// Default to no capabilities
		return NoComputation
	}
}

type ProcessingLocation string

const (
	// Specified if the request should ideally be executed locally and never leave the device
	Local ProcessingLocation = "local"
	// Specified if the returned data would ideally be unprocessed
	Remote ProcessingLocation = "remote"
)

// RequestPolicy stores the preferred location for processing of a request (and the identity of the requester?)
type RequestPolicy struct {
	ID                          string
	PreferredProcessingLocation ProcessingLocation
	HasAllRequiredData          bool
}

// AddToParams adds each of its fields as a parameter in the passed Values struct
func (p *RequestPolicy) AddToParams(params *url.Values) {
	// TODO
	preferredProcessingLocation := string(p.PreferredProcessingLocation)
	params.Set("preferred_processing_location", preferredProcessingLocation)
	return
}
