package middleware

import (
	"errors"
	"log"
	"net/http"
)

// PamRequest contains a RequestPolicy and a http request
type PamRequest struct {
	Policy      *RequestPolicy
	HttpRequest *http.Request
}

// AddParam adds a parameter to the contained http request with the same semantics as adding to the URL.Values of the
// http request
func (r *PamRequest) AddParam(key, value string) {
	httpRequest := r.HttpRequest
	params := httpRequest.URL.Query()
	params.Add(key, value)
	httpRequest.URL.RawQuery = params.Encode()
}

// AddParam deletes a parameter from the contained http request with the same semantics as adding to the URL.Values of the
// http request
func (r *PamRequest) DelParam(key string) {
	httpRequest := r.HttpRequest
	params := httpRequest.URL.Query()
	params.Del(key)
	httpRequest.URL.RawQuery = params.Encode()
}

// AddParam gets a parameter from the contained http request with the same semantics as adding to the URL.Values of the
// http request
func (r *PamRequest) GetParam(key string) string {
	httpRequest := r.HttpRequest
	params := httpRequest.URL.Query()
	return params.Get(key)
}

// AddParam sets a parameter on the contained http request with the same semantics as adding to the URL.Values of the
// http request
func (r *PamRequest) SetParam(key, value string) {
	httpRequest := r.HttpRequest
	params := httpRequest.URL.Query()
	params.Set(key, value)
	httpRequest.URL.RawQuery = params.Encode()
}

// BuildPamRequest takes a pointer to a http request and returns a PamRequest with policy values taken from
// the parameters of the passed request
func BuildPamRequest(req *http.Request) (PamRequest, error) {
	policy, err := BuildRequestPolicy(req)
	if err != nil {
		return PamRequest{}, err
	}
	pamRequest := PamRequest{
		HttpRequest: req,
		Policy:      policy,
	}
	return pamRequest, nil
}

// PamResponse contains a http response and its associated computation level
type PamResponse struct {
	ComputationLevel ComputationLevel
	HttpResponse     *http.Response
}

// BuildPamResponse takes a pointer to a http response and returns a PamResponse with the ComputationLevel taken from
// the response header
func BuildPamResponse(resp *http.Response) (PamResponse, error) {
	// Query response to see if this is a partial globalResult
	computationLevelString := resp.Header.Get("computation_level")
	if computationLevelString == "" {
		return PamResponse{}, errors.New("the response did not specify a computation level")
	}

	computationLevel, err := ComputationLevelFromString(computationLevelString)
	if err != nil {
		return PamResponse{}, err
	}

	return PamResponse{
		ComputationLevel: computationLevel,
		HttpResponse:     resp,
	}, nil
}

// PrivacyAwareHandler returns a http.Handler based on the passed
// ComputationPolicy. It also performs some basic logging of requests received.
func PrivacyAwareHandler(policy ComputationPolicy) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("PAM: handling path: ", r.URL.Path)

		// Get preferred processing location
		pamRequest, err := BuildPamRequest(r)
		if err != nil {
			log.Println(err.Error())
			http.Error(w, err.Error(), 500)
		}
		preferredLocation := pamRequest.Policy.PreferredProcessingLocation

		// Get the handler the policy specifies for this path and preferred processing location
		computationLevel, handler := policy.Resolve(r.URL.Path, preferredLocation)

		switch computationLevel {
		case NoComputation:
			//// Return 403: FORBIDDEN as we are currently refusing to compute this globalResult
			//// 404: NOT FOUND may be better in some cases as this is what you get for an unregistered path
			//http.Error(w, "Cannot compute globalResult", 403)
			w.Header().Set("computation_level", "NoComputation")
		case CanCompute:
			w.Header().Set("computation_level", "CanCompute")
			handler.ServeHTTP(w, r)
		case RawData:
			w.Header().Set("computation_level", "RawData")
			handler.ServeHTTP(w, r)
		}
		log.Println("PAM: finished serving: ", r.URL.Path)
	})
}
