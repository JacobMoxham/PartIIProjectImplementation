package main

import (
	_ "expvar"
	"fmt"
	"github.com/JacobMoxham/PartIIProjectImplementation/middleware"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
)

const DOCKER = true

func createMakeRequestHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Request Received")
		policy := middleware.RequestPolicy{
			RequesterID:                 "client1",
			PreferredProcessingLocation: middleware.Remote,
			HasAllRequiredData:          false,
		}

		var httpRequest *http.Request
		if DOCKER {
			httpRequest, _ = http.NewRequest("GET", "http://server:3002/get-average-power-consumption", nil)
		} else {
			httpRequest, _ = http.NewRequest("GET", "http://127.0.0.1:3002/get-average-power-consumption", nil)
		}

		req := middleware.PamRequest{
			Policy:      &policy,
			HttpRequest: httpRequest,
		}

		// Get date interval from received request
		requestParams := r.URL.Query()
		startDate := requestParams.Get("startDate")
		endDate := requestParams.Get("endDate")

		// Add date interval to request to send
		req.SetParam("startDate", startDate)
		req.SetParam("endDate", endDate)

		pamResp, err := req.Send()
		if err != nil {
			log.Println("Error:", err)
			return
		}
		if pamResp.HttpResponse.StatusCode < 200 || pamResp.HttpResponse.StatusCode >= 300 {
			log.Printf("Request produced a none 2xx status code: %s", pamResp.HttpResponse.Status)
			return
		}

		// We ignore the computation level for this example

		resp := pamResp.HttpResponse

		// Read response
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("Error reading body of response.", err)
		}
		resp.Body.Close()

		log.Println(fmt.Sprintf("Code: %d Result: %s", resp.StatusCode, body))
		_, err = w.Write([]byte(body))
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}
}

func main() {
	// Logging for performance analysis
	go func() {
		log.Println(http.ListenAndServe("0.0.0.0:6060", nil))
	}()

	// Listen on 4000 for request to start example
	http.Handle("/request", http.HandlerFunc(createMakeRequestHandler()))
	log.Println("Listening...")
	err := http.ListenAndServe(":4000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
