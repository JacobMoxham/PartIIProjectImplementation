package usecases

import (
	"fmt"
	"github.com/JacobMoxham/PartIIProjectImplementation/middleware"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

func imageProcessingHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("PamRequest Received")
	w.Write([]byte("Processed Image Goes Here"))
}

func imServer() {
	// Create actual function to run
	finalHandler := http.HandlerFunc(imageProcessingHandler)

	// Define computation policy
	computationPolicy := middleware.NewStaticComputationPolicy()
	computationPolicy.Register("/", middleware.CanCompute, finalHandler)

	// Chain together "other" middlewares
	handler := middleware.PrivacyAwareHandler(computationPolicy)

	// Register the composite handler at '/' on port 3000
	http.Handle("/", handler)
	log.Println("Listening...")
	http.ListenAndServe(":3000", nil)
}

func imClient() {
	for i := 0; i < 10; i++ {
		policy := middleware.RequestPolicy{
			RequesterID:                 "client1",
			PreferredProcessingLocation: middleware.Remote,
			HasAllRequiredData:          true,
		}
		httpRequest, _ := http.NewRequest("GET", "http://127.0.0.1:3000/", nil)
		req := middleware.PamRequest{
			Policy:      &policy,
			HttpRequest: httpRequest,
		}
		resp, err := req.Send()
		if err != nil {
			log.Println("Error:", err)
			continue
		}
		if resp.HttpResponse.StatusCode < 200 || resp.HttpResponse.StatusCode >= 300 {
			log.Printf("Request produced a none 2xx status code: %s", resp.HttpResponse.Status)
			continue
		}

		// Read response
		body, err := ioutil.ReadAll(resp.HttpResponse.Body)
		if err != nil {
			log.Fatal("Error reading body of response.", err)
		}
		resp.HttpResponse.Body.Close()

		log.Println(fmt.Sprintf("%d: %s", i, body))
	}
}

func processing_main() {
	var wg sync.WaitGroup
	wg.Add(2)
	go imServer()
	time.Sleep(50)
	go imClient()
	wg.Wait()
}
