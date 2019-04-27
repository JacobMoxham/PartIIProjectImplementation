package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/JacobMoxham/PartIIProjectImplementation/middleware"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"strconv"
)

const DOCKER = true

func createGetAveragePowerConsumptionHandler(computationPolicy middleware.ComputationPolicy) (func(http.ResponseWriter, *http.Request), error) {
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

	httpClient := middleware.MakePolicyAwareClient(computationPolicy)

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

			pamResp, err := httpClient.Send(req)
			if err != nil {
				log.Println("Error:", err)
				http.Error(w, err.Error(), 500)
				return
			}
			if pamResp.HttpResponse.StatusCode < 200 || pamResp.HttpResponse.StatusCode >= 300 {
				log.Printf("Request to %s produced a none 2xx status code: %s", client, pamResp.HttpResponse.Status)
				continue
			}

			// We ignore the computation level for this example

			resp := pamResp.HttpResponse

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
			if len(lines) > 0 {
				averageActiveEnergyPerMinute := totalActiveEnergyPerMinute / float64(len(lines))
				averageActiveEnergyPerMinuteFromAllClients += averageActiveEnergyPerMinute
			}
		}
		if len(clients) > 0 {
			averageActiveEnergyPerMinuteFromAllClients /= float64(len(clients))
		}
		_, err = w.Write([]byte(fmt.Sprintf("%.2f", averageActiveEnergyPerMinuteFromAllClients)))
		if err != nil {
			log.Println("Error:", err)
			http.Error(w, err.Error(), 500)
			return
		}
	}, nil
}

func main() {
	// Define computation policy
	computationPolicy := middleware.NewStaticComputationPolicy()

	// Create actual function to run
	getAveragePowerConsumptionHandler, err := createGetAveragePowerConsumptionHandler(computationPolicy)
	if err != nil {
		log.Fatal(err)
	}
	handler := http.HandlerFunc(getAveragePowerConsumptionHandler)

	computationPolicy.Register("/get-average-power-consumption", middleware.CanCompute, handler)

	// Register the composite handler at '/get-average-power-consumption' on port 3002
	http.Handle("/get-average-power-consumption", middleware.PolicyAwareHandler(computationPolicy))
	log.Println("Listening...")
	err = http.ListenAndServe(":3002", nil)
	if err != nil {
		log.Fatal(err)
	}
}
