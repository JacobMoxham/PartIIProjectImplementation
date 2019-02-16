package middleware

import "net/http"

type PrivacyAwareClient struct {
	client            *http.Client
	computationPolicy *ComputationPolicy
}

func MakePrivacyAwareClient(policy ComputationPolicy) PrivacyAwareClient {
	return PrivacyAwareClient{
		client:            &http.Client{},
		computationPolicy: &policy,
	}
}

func (c PrivacyAwareClient) Send(req PamRequest) (PamResponse, error) {
	// TODO: consider whether to copy requests before sending as we need to edit the body - probably fine without
	httpRequest := req.HttpRequest

	// Add the query params from the policy
	params := httpRequest.URL.Query()
	req.Policy.AddToParams(&params)
	httpRequest.URL.RawQuery = params.Encode()

	// TODO: consider whether we can handle this if we have all of the data and can process locally.
	//  This may require us to specify a dependency on "has all data" when registering handlers, I'll have a think

	resp, err := c.client.Do(httpRequest)
	if err != nil {
		return PamResponse{}, err
	}

	// TODO: consider whether there should be sender config which says whether or not we can use partial results/full
	// results, it could possibly fit into the same framework
	return BuildPamResponse(resp)
}
