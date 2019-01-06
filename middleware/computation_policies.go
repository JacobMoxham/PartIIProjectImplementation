package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type ComputationCapability int

const (
	// TODO: consider keep a list of these or adding in CanComputeOrRawData - seems worse though
	NoComputation ComputationCapability = iota
	RawData       ComputationCapability = iota
	CanCompute    ComputationCapability = iota
)

func computationCapabilityFromString(capability string) (ComputationCapability, error) {
	switch strings.ToLower(capability) {
	case strings.ToLower("NoComputation"):
		return NoComputation, nil
	case strings.ToLower("RawData"):
		return RawData, nil
	case strings.ToLower("CanCompute"):
		return CanCompute, nil
	default:
		return 0, fmt.Errorf("cannot parse %s as a computation capability", capability)
	}
}

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

func processingLocationFromString(loc string) (ProcessingLocation, error) {
	switch strings.ToLower(loc) {
	case "local":
		return Local, nil
	case "remote":
		return Remote, nil
	default:
		return "", fmt.Errorf("cannot parse %s as a processing location", loc)
	}
}

// RequestPolicy stores the preferred location for processing of a request (and the identity of the requester?)
type RequestPolicy struct {
	RequesterID                 string
	PreferredProcessingLocation ProcessingLocation
	HasAllRequiredData          bool
}

// AddToParams adds each of its fields as a parameter in the passed Values struct
func (p *RequestPolicy) AddToParams(params *url.Values) {
	preferredProcessingLocation := string(p.PreferredProcessingLocation)
	hasAllRequiredData := strconv.FormatBool(p.HasAllRequiredData)
	params.Set("requester_id", p.RequesterID)
	params.Set("preferred_processing_location", preferredProcessingLocation)
	params.Set("has_all_required_data", hasAllRequiredData)

	return
}

func BuildRequestPolicy(req *http.Request) (*RequestPolicy, error) {
	params := req.URL.Query()
	requesterID := params.Get("requester_id")
	if requesterID == "" {
		return nil, errors.New("a context cannot be parsed from the request as there is no requester ID")
	}

	preferredProcessingLocation := params.Get("preferred_processing_location")
	if requesterID == "" {
		return nil, errors.New("a context cannot be parsed from the request as there is no preferred processing location")
	}
	preferredProcessingLocationEnum, err := processingLocationFromString(preferredProcessingLocation)
	if err != nil {
		return nil, fmt.Errorf("%s cannot be parsed as a processing location", preferredProcessingLocation)
	}

	hasAllRequiredData := params.Get("has_all_required_data")
	if requesterID == "" {
		return nil, errors.New("a context cannot be parsed from the request as there is no \"has all required data\" field")
	}
	hasAllRequiredDataBool, err := strconv.ParseBool(hasAllRequiredData)
	if err != nil {
		return nil, fmt.Errorf("%s cannot be parsed as a bool", hasAllRequiredData)
	}
	return &RequestPolicy{RequesterID: requesterID,
		PreferredProcessingLocation: preferredProcessingLocationEnum,
		HasAllRequiredData:          hasAllRequiredDataBool,
	}, nil
}
