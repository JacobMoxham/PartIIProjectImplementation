package middleware

import (
	"net/http"
	"net/http/httptest"
)

// PrivacyAwareClient wraps a http client with a ComputationPolicy
type PrivacyAwareClient struct {
	client            *http.Client
	computationPolicy ComputationPolicy
}

// MakePrivacyAwareClient returns an PrivacyAwareClient with initialised fields
func MakePrivacyAwareClient(policy ComputationPolicy) PrivacyAwareClient {
	return PrivacyAwareClient{
		client:            &http.Client{},
		computationPolicy: policy,
	}
}

// Send takes a PamRequest and martials the RequestPolicy into the http request parameters before sending it using the
// contained http client. If the ComputationPolicy has a local handler for the requested path, and the preferred
// location is local, and all of the data required for a globalResult is contained within the request then the request will
// instead be handled locally.
func (c PrivacyAwareClient) Send(req PamRequest) (PamResponse, error) {
	httpRequest := req.HttpRequest

	// Add the query params from the policy
	params := httpRequest.URL.Query()
	req.Policy.AddToParams(&params)
	httpRequest.URL.RawQuery = params.Encode()

	// Check if we would prefer to process locally
	policy := req.Policy
	preferLocal := policy.PreferredProcessingLocation == Local

	// Check if we can process this request locally
	requestPath := req.HttpRequest.URL.Path
	computationPolicy := c.computationPolicy
	// TODO: check if this should be Local or Remote
	// Pass Remote to resolve so that we get a CanCompute handler if it is available
	computationLevel, localHandler := computationPolicy.Resolve(requestPath, Remote)

	// Check if we have all of the required data to process this locally within the request
	allRequiredData := req.Policy.HasAllRequiredData

	if preferLocal && computationLevel != NoComputation && allRequiredData {
		// TODO: consider whether its good practice to use a testing library in this way
		// Use the httptest ResponseRecorder to get the globalResult locally
		responseRecorder := httptest.NewRecorder()
		localHandler.ServeHTTP(responseRecorder, httpRequest)

		// Copy from the recorded response into a normal response
		resp := responseRecorder.Result()
		resp.Header.Set("computation_level", computationLevel.ToString())
		return BuildPamResponse(resp)
	}

	resp, err := c.client.Do(httpRequest)
	if err != nil {
		return PamResponse{}, err
	}

	return BuildPamResponse(resp)
}
