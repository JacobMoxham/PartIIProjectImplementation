package policies

import "net/http"

type DataPolicy interface {
	defaultHandler(w http.ResponseWriter, r *http.Request)
}

type ComputationPolicy interface {
}

type NodePolicy struct {
	dataPolicy DataPolicy
	nodePolicy NodePolicy
}

type RequestPolicy interface {
}

type PolicyResolver interface {
	// Thoughts: maybe pass a request which contains a policy rather than the request policy
	// Do we get back the response or some policy dictating how we will then go and get the response
	resolve(nodePolicy DataPolicy, requestPolicy RequestPolicy)
}
