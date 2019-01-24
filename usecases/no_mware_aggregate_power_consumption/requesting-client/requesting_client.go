package main

import (
	_ "expvar"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
)

func createMakeRequestHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		httpRequest, _ := http.NewRequest("GET", "http://no-mware-server:3002/get-average-power-consumption", nil)
		//httpRequest, _ := http.NewRequest("GET", "http://127.0.0.1:3002/get-average-power-consumption", nil)

		// Get date interval from received request
		requestParams := r.URL.Query()
		startDate := requestParams.Get("startDate")
		endDate := requestParams.Get("endDate")

		// Add date interval to request to send
		params := httpRequest.URL.Query()
		params.Set("startDate", startDate)
		httpRequest.URL.RawQuery = params.Encode()
		params.Set("endDate", endDate)
		httpRequest.URL.RawQuery = params.Encode()

		client := http.Client{}
		resp, err := client.Do(httpRequest)
		if err != nil {
			log.Println("Error:", err)
			return
		}

		// Read response
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("Error reading body of response.", err)
		}
		resp.Body.Close()

		log.Println(fmt.Sprintf("%s", body))
		_, err = w.Write([]byte(body))
		if err != nil {
			http.Error(w, err.Error(), 200)
			return
		}
		log.Println("Handler Complete")
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
