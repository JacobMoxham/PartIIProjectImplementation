package main

import (
	"fmt"
	"github.com/JacobMoxham/PartIIProjectImplementation/middleware"
	"github.com/justinas/alice"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

// Simple example of another middleware to show this can build on top of other things
func enforceGETHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that request has method GET
		if r.Method != "GET" {
			http.Error(w, "Must be a GET request", 200)
		}

		next.ServeHTTP(w, r)
	})
}

func basicServer() {
	// Create actual function to run
	finalHandler := http.HandlerFunc(basicFinal)

	// Define computation policy
	computationPolicy := middleware.NewStaticComputationPolicy()

	handlers := alice.New(enforceGETHandler).Then(finalHandler)

	computationPolicy.Register("/", middleware.CanCompute, handlers)

	// Register the composite handler at '/' on port 3000
	http.Handle("/", middleware.PrivacyAwareHandler(computationPolicy))
	log.Println("Listening...")
	http.ListenAndServe(":3000", nil)
}

func basicClient() {
	client := middleware.MakePrivacyAwareClient(middleware.NewStaticComputationPolicy())

	for i := 0; i < 10; i++ {
		policy := middleware.RequestPolicy{
			RequesterID:                 "basic",
			HasAllRequiredData:          false,
			PreferredProcessingLocation: middleware.Remote,
		}
		httpRequest, _ := http.NewRequest("GET", "http://127.0.0.1:3000/", nil)
		req := middleware.PamRequest{
			Policy:      &policy,
			HttpRequest: httpRequest,
		}

		pamResp, err := client.Send(req)
		if err != nil {
			log.Println("Error:", err)
			continue
		}
		if pamResp.HttpResponse.StatusCode < 200 || pamResp.HttpResponse.StatusCode >= 300 {
			log.Printf("Request produced a none 2xx status code: %s", pamResp.HttpResponse.Status)
			continue
		}

		resp := pamResp.HttpResponse

		// Read response
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("Error reading body of response.", err)
		}
		resp.Body.Close()

		log.Println(fmt.Sprintf("%d: %s", i, body))
	}
}

func main() {
	var wg sync.WaitGroup
	wg.Add(2)
	go basicServer()
	time.Sleep(50)
	go basicClient()
	wg.Wait()
}

func basicFinal(w http.ResponseWriter, r *http.Request) {
	log.Println("PamRequest Received")
	w.Write([]byte("Hello World"))
}
