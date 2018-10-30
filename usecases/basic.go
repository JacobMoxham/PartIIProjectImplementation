package main

import (
	"fmt"
	"github.com/JacobMoxham/partiiproject/middleware"
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

func server() {
	// Create actual function to run
	finalHandler := http.HandlerFunc(final)

	// Chain together "other" middlewares
	regularHandlers := alice.New(enforceGETHandler).Then(finalHandler)

	// Initialise data policy
	nodePolicy := middleware.NodePolicy{
		DataPolicy: &middleware.DataPolicy{
			DefaultHandler: &regularHandlers,
		},
		ComputationPolicy: &middleware.ComputationPolicy{},
	}

	// Add in the privacy aware middleware
	privacyAwareHandler := middleware.PrivacyAwareHandler(nodePolicy)

	// Register the composite handler at '/' on port 3000
	http.Handle("/", privacyAwareHandler)
	log.Println("Listening...")
	http.ListenAndServe(":3000", nil)
}

func client() {
	for i := 0; i < 10; i++ {
		policy := middleware.RequestPolicy{}
		httpRequest, _ := http.NewRequest("GET", "http://127.0.0.1:3000/", nil)
		req := middleware.Request{
			Policy:      &policy,
			HttpRequest: httpRequest,
		}
		resp, err := req.Send()
		if err != nil {
			log.Println("Error:", err)
			continue
		}

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
	go server()
	time.Sleep(50)
	go client()
	wg.Wait()
}

// Simply write OK back
func final(w http.ResponseWriter, r *http.Request) {
	log.Println("Request Received")
	w.Write([]byte("Hello World"))
}
