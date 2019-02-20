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
	"os"
	"strconv"
)

const DOCKER = true

func getAverageForClient(client string, httpClient middleware.PrivacyAwareClient, startDate, endDate string,
	policy middleware.RequestPolicy, result chan float64) error {
	log.Printf("Requesting data from %s", client)
	httpRequest, err := http.NewRequest("GET", fmt.Sprintf("http://%s:3001/", client), nil)
	if err != nil {
		return err
	}

	req := middleware.PamRequest{
		Policy:      &policy,
		HttpRequest: httpRequest,
	}

	req.SetParam("startDate", startDate)
	req.SetParam("endDate", endDate)

	pamResp, err := httpClient.Send(req)
	if err != nil {
		log.Printf("Request to %s produced an error: %s", client, err.Error())
		return err
	}
	if pamResp.HttpResponse.StatusCode < 200 || pamResp.HttpResponse.StatusCode >= 300 {
		// Read response
		resp := pamResp.HttpResponse
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal("Error reading body of response.", err)
		}
		resp.Body.Close()

		log.Printf("Request to %s produced a none 2xx status code: %s, %s", client, pamResp.HttpResponse.Status, body)
		return err
	}

	resp := pamResp.HttpResponse

	// Read response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading body of response:", err)
		resp.Body.Close()
		return err
	}

	switch pamResp.ComputationLevel {
	case middleware.CanCompute:
		averageActiveEnergyPerMinute, err := strconv.ParseFloat(string(body), 64)
		if err != nil {
			log.Printf("%s returned a value which could not be parsed as a float, error: %s", client, err)
			return err
		}
		log.Printf("%s gave value %f", client, averageActiveEnergyPerMinute)

		result <- averageActiveEnergyPerMinute
	case middleware.RawData:
		reader := csv.NewReader(bytes.NewBuffer(body))
		lines, err := reader.ReadAll()
		if err != nil {
			return err
		}

		totalActiveEnergyPerMinute := 0.0
		for i, line := range lines {
			// Skip header
			if i > 0 {
				// Average the active energy per minute from all clients
				activeEnergyPerMinute, err := strconv.ParseFloat(line[1], 64)
				if err != nil {
					return err
				}
				totalActiveEnergyPerMinute += activeEnergyPerMinute
			}
		}
		// Do a division here to keep numbers smaller
		averageActiveEnergyPerMinute := totalActiveEnergyPerMinute / float64(len(lines))
		log.Printf("%s gave value %f", client, averageActiveEnergyPerMinute)

		result <- averageActiveEnergyPerMinute
	case middleware.NoComputation:
		log.Printf("client: %s could not compute a result", client)
	}

	return nil
}

func createGetAveragePowerConsumptionHandler(computationPolicy middleware.ComputationPolicy) (func(http.ResponseWriter, *http.Request), error) {
	policy := middleware.RequestPolicy{
		RequesterID:                 "server",
		PreferredProcessingLocation: middleware.Local,
		HasAllRequiredData:          false,
	}

	var clients []string
	if DOCKER {
		// Support docker compose command line arguments
		if len(os.Args) > 1 {
			log.Printf("Clients: %v", os.Args[1:])
			clients = os.Args[1:]
		} else {
			clients = []string{"data-client-compute", "data-client-raw-data", "data-client-both", "data-client-none"}
		}
	} else {
		clients = []string{"127.0.0.1"}
	}

	httpClient := middleware.MakePrivacyAwareClient(computationPolicy)

	return func(w http.ResponseWriter, r *http.Request) {
		pamRequest, err := middleware.BuildPamRequest(r)
		startDate := pamRequest.GetParam("startDate")
		endDate := pamRequest.GetParam("endDate")

		averageActiveEnergyPerMinuteFromAllClients := 0.0

		// Channel for synchronisation
		done := make(chan bool, len(clients))
		results := make(chan float64, len(clients))

		for _, client := range clients {
			// Need to take a copy as we close over a reference to the client which will change
			clientCopy := client

			// Request an average (and compute one if we get raw data) in parallel
			go func() {
				err := getAverageForClient(clientCopy, httpClient, startDate, endDate, policy, results)
				if err != nil {
					log.Printf("Could not get an average for %s, there was an error: %s", client, err.Error())
				}
				done <- true
			}()
		}

		// Wait for all clients to finish
		respondingClientCount := 0
		for i := 0; i < len(clients); i++ {
			<-done
		}

		// Accumulate all the results clients provided
		close(results)
		for averageActiveEnergyPerMinute := range results {
			respondingClientCount += 1
			averageActiveEnergyPerMinuteFromAllClients += averageActiveEnergyPerMinute
		}

		// Don't divide by 0
		if respondingClientCount > 0 {
			averageActiveEnergyPerMinuteFromAllClients /= float64(respondingClientCount)
		} else {
			http.Error(w, "No data could be found", 500)
			return
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
	http.Handle("/get-average-power-consumption", middleware.PrivacyAwareHandler(computationPolicy))
	log.Println("Listening...")
	err = http.ListenAndServe(":3002", nil)
	if err != nil {
		log.Fatal(err)
	}
}
