package middleware

import (
	"net/http"
)

// TODO: rename to <MwareName>Request once I have decided on an Mware name
type Request struct {
	Policy      *RequestPolicy
	HttpRequest *http.Request
}

// TODO: rename to <MwareName>Request once I have decided on an Mware name
type Response struct {
	PartialResult bool // TODO: Consider extending to a configurable enum
	HttpResponse  http.Response
}

func (r *Request) Send() (*http.Response, error) {
	// TODO: populate client fields - this should maybe be a func of a client rather than a request itself we will see

	// TODO: consider whether to copy requests before sending as we need to edit the body - probably fine without
	//request := *r.HttpRequest

	httpRequest := r.HttpRequest

	// Add the query params from the policy
	params := httpRequest.URL.Query()
	r.Policy.AddToParams(&params)
	httpRequest.URL.RawQuery = params.Encode()

	client := http.Client{}
	resp, err := client.Do(httpRequest)
	if err != nil {
		return nil, err
	}

	// TODO: handle middleware only fields either here or at the next level up

	return resp, nil
}

// Example of a very simple go middleware which takes a DataPolicy and returns its default handler
func PrivacyAwareHandler(policy NodePolicy) http.Handler {
	// TODO: Add the logic to decide how to resolve policies given and whether we will be returning a partial result
	return *policy.DataPolicy.DefaultHandler
}
