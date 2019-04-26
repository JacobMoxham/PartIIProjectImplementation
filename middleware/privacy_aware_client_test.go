package middleware

import (
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestPrivacyAwareClient_Send_RunLocally(t *testing.T) {
	returnValue := []byte("Computed Locally")

	computationPolicy := NewStaticComputationPolicy()
	computationPolicy.Register("/", CanCompute, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(returnValue)
	}))

	// Add a raw data handler to ensure we resolve to get the correct handler
	computationPolicy.Register("/", RawData, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("raw data"))
	}))

	client := MakePrivacyAwareClient(computationPolicy)

	request, err := http.NewRequest("GET", "http://ip/", nil)
	require.NoError(t, err)

	pamResp, err := client.Send(PamRequest{
		&RequestPolicy{
			PreferredProcessingLocation: Local,
			HasAllRequiredData:          true,
			RequesterID:                 "client1",
		},
		request,
	})
	require.NoError(t, err)

	resp := pamResp.HttpResponse

	body, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	err = resp.Body.Close()
	require.NoError(t, err)

	require.Equal(t, returnValue, body)
}
