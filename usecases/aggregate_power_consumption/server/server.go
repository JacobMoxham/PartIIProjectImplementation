package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/JacobMoxham/PartIIProjectImplementation/middleware"
	"github.com/justinas/alice"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"strconv"
)

const DOCKER = false

func createGetAveragePowerConsumptionHandler() (func(http.ResponseWriter, *http.Request), error) {
	policy := middleware.RequestPolicy{
		RequesterID:                 "server",
		PreferredProcessingLocation: middleware.Remote,
		HasAllRequiredData:          false,
	}

	var clients []string
	if DOCKER {
		clients = []string{"data-client"}
	} else {
		clients = []string{"127.0.0.1"}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		pamRequest, err := middleware.BuildPamRequest(r)
		startDate := pamRequest.GetParam("startDate")
		endDate := pamRequest.GetParam("endDate")

		averageActiveEnergyPerMinuteFromAllClients := 0.0
		for _, client := range clients {
			httpRequest, _ := http.NewRequest("GET", fmt.Sprintf("http://%s:3001/", client), nil)
			req := middleware.PamRequest{
				Policy:      &policy,
				HttpRequest: httpRequest,
			}
			req.SetParam("startDate", startDate)
			req.SetParam("endDate", endDate)

			resp, err := req.Send()
			// TODO: check if the database failed to connect and error properly
			if err != nil {
				log.Println("Error:", err)
				http.Error(w, err.Error(), 500)
				return
			}

			// Read response
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Fatal("Error reading body of response.", err)
			}
			resp.Body.Close()

			// logging
			log.Println(fmt.Sprintf("Code: %d Body: %s", resp.StatusCode, body[:100]))

			reader := csv.NewReader(bytes.NewBuffer(body))
			lines, err := reader.ReadAll()
			if err != nil {
				log.Println("Error:", err)
				http.Error(w, err.Error(), 500)
				return
			}
			totalActiveEnergyPerMinute := 0.0
			for i, line := range lines {
				// Skip header
				if i > 0 {
					// Average the active energy per minute from all clients
					activeEnergyPerMinute, err := strconv.ParseFloat(line[1], 64)
					if err != nil {
						log.Println("Error:", err)
						http.Error(w, err.Error(), 500)
						return
					}
					totalActiveEnergyPerMinute += activeEnergyPerMinute
				}
			}
			// Do a division here to keep numbers smaller
			averageActiveEnergyPerMinute := totalActiveEnergyPerMinute / float64(len(lines))
			averageActiveEnergyPerMinuteFromAllClients += averageActiveEnergyPerMinute
		}
		averageActiveEnergyPerMinuteFromAllClients /= float64(len(clients))
		log.Printf("Average Active Energy Per Minute: %.2f\n", averageActiveEnergyPerMinuteFromAllClients)
		_, err = w.Write([]byte(fmt.Sprintf("%.2f", averageActiveEnergyPerMinuteFromAllClients)))
		if err != nil {
			log.Println("Error:", err)
			http.Error(w, err.Error(), 500)
			return
		}
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
