package middleware

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type ComputationLevel int

const (
	NoComputation ComputationLevel = iota
	RawData       ComputationLevel = iota
	CanCompute    ComputationLevel = iota
)

func ComputationLevelFromString(level string) (ComputationLevel, error) {
	switch strings.ToLower(level) {
	case strings.ToLower("NoComputation"):
		return NoComputation, nil
	case strings.ToLower("RawData"):
		return RawData, nil
	case strings.ToLower("CanCompute"):
		return CanCompute, nil
	default:
		return 0, fmt.Errorf("cannot parse %s as a computation level", level)
	}
}

type ComputationCapability map[ComputationLevel]http.Handler

// ComputationPolicy stores computation capabilities of a node
type ComputationPolicy interface {
	Register(string, ComputationLevel, http.Handler)
	UnregisterAll(string)
	UnregisterOne(string, ComputationLevel)
	Resolve(string, ProcessingLocation) (ComputationLevel, http.Handler)
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

func (p *StaticComputationPolicy) Register(path string, level ComputationLevel, handler http.Handler) {
	if p.capabilities[path] == nil {
		p.capabilities[path] = make(ComputationCapability)
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
		canComputehandler, hasCanComputeHandler := capability[CanCompute]

		// TODO: in dynamic version we may have a "valid" tag?
		if hasCanComputeHandler {
			if hasRawDataHandler {
				if preferredLocation == Remote {
					log.Println("Serving request as preferred location is remote and we can compute")
					return CanCompute, canComputehandler
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
				return CanCompute, canComputehandler
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

type ProcessingLocation string

const (
	// Specified if the request should ideally be executed locally and never leave the device
	Local ProcessingLocation = "local"
	// Specified if the returned data would ideally be unprocessed
	Remote ProcessingLocation = "remote"
)

func ProcessingLocationFromString(loc string) (ProcessingLocation, error) {
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
