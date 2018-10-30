package middleware

import (
	"net/http"
	"net/url"
)

// DataPolicy stores the preferences for the location of the data and any associated handlers for requests which they
// may wish to run
type DataPolicy struct {
	DefaultHandler *http.Handler
}

// ComputationPolicy stores computation capabilities of a node
type ComputationPolicy struct {
}

// NodePolicy combines an nodes data and computation policies
type NodePolicy struct {
	DataPolicy        *DataPolicy
	ComputationPolicy *ComputationPolicy
}

// RequestPolicy stores the preferred location for processing of a request (and the identity of the requester?)
type RequestPolicy struct {
}

// Resolves a node and request policy to give the appropriate handler for a request
type PolicyResolver interface {
	// Thoughts: maybe pass a request which contains a policy rather than the request policy
	// Do we get back the response or some policy dictating how we will then go and get the response
	resolve(nodePolicy DataPolicy, requestPolicy RequestPolicy) http.Handler
}

// AddToParams adds each of its fields as a parameter in the passed Values struct
func (p *RequestPolicy) AddToParams(params *url.Values) {
	return
}
