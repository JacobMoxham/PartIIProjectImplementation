package middleware

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func privHandler(local, remote bool) http.HandlerFunc {
	policy := NewStaticComputationPolicy()
	if local {
		policy.Register("/", RawData, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("local"))
		}))
	}
	if remote {
		policy.Register("/", CanCompute, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("remote"))
		}))
	}
	return PrivacyAwareHandler(policy)
}

func TestPrivacyAwareHandler(t *testing.T) {
	testCases := []struct {
		localHandler      bool
		remoteHandler     bool
		preferredLocation ProcessingLocation
		output            string
		computationLevel  ComputationLevel
	}{
		{true, true, Local, "local", RawData},
		{true, false, Local, "local", RawData},
		{true, true, Remote, "remote", CanCompute},
		{false, true, Remote, "remote", CanCompute},
		{false, true, Local, "remote", CanCompute},
		{true, false, Remote, "local", RawData},
		{false, false, Local, "", NoComputation},
		{false, false, Remote, "", NoComputation},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Local: %t Remote: %t Preferred: %s", tc.localHandler, tc.remoteHandler,
			tc.preferredLocation.ToString()), func(t *testing.T) {

			privHandler := privHandler(tc.localHandler, tc.remoteHandler)
			responseRecorder := httptest.NewRecorder()
			request, err := http.NewRequest("GET", "http://127.0.0.1:3007/", nil)
			require.NoError(t, err)

			pamReq := PamRequest{
				HttpRequest: request,
				Policy: &RequestPolicy{
					PreferredProcessingLocation: tc.preferredLocation,
					RequesterID:                 "Jacob",
				},
			}

			// Add the query params from the policy
			params := request.URL.Query()
			pamReq.Policy.AddToParams(&params)
			request.URL.RawQuery = params.Encode()

			privHandler.ServeHTTP(responseRecorder, request)

			resp, err := BuildPamResponse(responseRecorder.Result())
			require.NoError(t, err)

			require.Equal(t, tc.computationLevel, resp.ComputationLevel)
			if tc.computationLevel != NoComputation {
				body, err := ioutil.ReadAll(resp.HttpResponse.Body)
				require.NoError(t, err)
				require.Equal(t, tc.output, string(body))
			}
		})
	}
}
