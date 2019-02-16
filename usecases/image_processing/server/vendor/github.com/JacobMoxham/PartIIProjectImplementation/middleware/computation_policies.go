package middleware

import (
	"errors"
	"fmt"
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

func (c ComputationLevel) ToString() string {
	switch c {
	case NoComputation:
		return "NoComputation"
	case RawData:
		return "RawData"
	case CanCompute:
		return "CanCompute"
	}

	// This will never occur but a return is needed
	return ""
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
	preferredProcessingLocationEnum, err := ProcessingLocationFromString(preferredProcessingLocation)
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

// ComputationPolicy stores computation capabilities of a node
type ComputationPolicy interface {
	Register(string, ComputationLevel, http.Handler)
	UnregisterAll(string)
	UnregisterOne(string, ComputationLevel)
	Resolve(string, ProcessingLocation) (ComputationLevel, http.Handler)
}
