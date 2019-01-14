package middleware

import (
	"log"
	"net/http"
)

type PamRequest struct {
	Policy      *RequestPolicy
	HttpRequest *http.Request
}

type PamResponse struct {
	PartialResult bool // TODO: Consider extending to a configurable enum
	HttpResponse  http.Response
}

func (r *PamRequest) Send() (*http.Response, error) {
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

func (r *PamRequest) AddParam(key string, value string) {
	httpRequest := r.HttpRequest
	params := httpRequest.URL.Query()
	params.Add(key, value)
}

// Example of a very simple go middleware which takes a Transforms and returns its default handler
// TODO: see if we can get this to fit the Handler interface
func PrivacyAwareHandler(policy ComputationPolicy) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Print("Handling path: ", r.URL.Path)
			capability := policy.Resolve(r.URL.Path)
			switch capability {
			case NoComputation:
				// TODO: correct error codes
				http.Error(w, "Cannot compute result", 200)
			case RawData:
				// TODO handler choosing correct handler
				log.Print("Partially serving request as we can only provide raw data")
				next.ServeHTTP(w, r)
			case CanCompute:
				// TODO: use the PamRequest build command
				preferredLocation := r.URL.Query().Get("preferred_processing_location")
				if preferredLocation == "remote" {
					log.Printf("Serving request as preferred location is %s and we can compute", preferredLocation)
					next.ServeHTTP(w, r)
				} else {
					// TODO: do as RawData if preferred is local (or allow user to specify)
					log.Printf("Partially serving request as preferred location is %s and we can compute", preferredLocation)
					next.ServeHTTP(w, r)
				}
			default:
				next.ServeHTTP(w, r)
			}
		})
	}
}
