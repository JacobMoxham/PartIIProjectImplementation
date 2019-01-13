package main

import (
	"fmt"
	"github.com/JacobMoxham/PartIIProjectImplementation/middleware"
	"github.com/justinas/alice"
	"io/ioutil"
	"log"
	"net/http"
)

var clients = []string{"data-client"}

func createGetAveragePowerConsumptionHandler() (func(http.ResponseWriter, *http.Request), error) {
	policy := middleware.RequestPolicy{
		RequesterID:                 "server",
		PreferredProcessingLocation: middleware.Remote,
		HasAllRequiredData:          false,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("PamRequest Received")
		returnString := ""
		for _, client := range clients {
			httpRequest, _ := http.NewRequest("GET", fmt.Sprintf("http://%s:3001/", client), nil)
			req := middleware.PamRequest{
				Policy:      &policy,
				HttpRequest: httpRequest,
			}
			resp, err := req.Send()
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
			// TODO: parse a better format and do some averaging
			returnString += string(body)
		}
		_, err := w.Write([]byte(returnString))
		if err != nil {
			http.Error(w, err.Error(), 200)
			return
		}
		log.Println("Handler Complete")
	}, nil
}

func main() {
	// Create actual function to run
	getAveragePowerConsumptionHandler, err := createGetAveragePowerConsumptionHandler()
	if err != nil {
		log.Fatal(err)
	}
	finalHandler := http.HandlerFunc(getAveragePowerConsumptionHandler)

	// Define computation policy
	computationPolicy := middleware.NewStaticComputationPolicy()
	computationPolicy.Register("/get-average-power-consumption", middleware.CanCompute)

	// Chain together "other" middlewares
	handlers := alice.New(middleware.PrivacyAwareHandler(computationPolicy)).Then(finalHandler)

	// Register the composite handler at '/get-average-power-consumption' on port 3002
	http.Handle("/get-average-power-consumption", handlers)
	log.Println("Listening...")
	err = http.ListenAndServe(":3002", nil)
	if err != nil {
		log.Fatal(err)
	}
}
