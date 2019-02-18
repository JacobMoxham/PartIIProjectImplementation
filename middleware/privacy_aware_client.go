package middleware

import (
	"net/http"
	"net/http/httptest"
)

type PrivacyAwareClient struct {
	client            *http.Client
	computationPolicy ComputationPolicy
}

func MakePrivacyAwareClient(policy ComputationPolicy) PrivacyAwareClient {
	return PrivacyAwareClient{
		client:            &http.Client{},
		computationPolicy: policy,
	}
}

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
	computationLevel, localHandler := computationPolicy.Resolve(requestPath, Local)

	// Check if we have all of the required data to process this locally within the request
	allRequiredData := req.Policy.HasAllRequiredData

	if preferLocal && computationLevel != NoComputation && allRequiredData {
		// TODO: consider whether its good practice to use a testing library in this way
		// Use the httptest ResponseRecorder to get the result locally
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
