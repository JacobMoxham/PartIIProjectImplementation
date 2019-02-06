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

func createGetAveragePowerConsumptionHandler() (func(http.ResponseWriter, *http.Request), error) {
	policy := middleware.RequestPolicy{
		RequesterID:                 "server",
		PreferredProcessingLocation: middleware.Local,
		HasAllRequiredData:          false,
	}

	var clients []string
	if DOCKER {
		clients = []string{"data-client-raw-data", "data-client-raw-data", "data-client-both", "data-client-no-computation"}
	} else {
		clients = []string{"127.0.0.1"}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		pamRequest, err := middleware.BuildPamRequest(r)
		startDate := pamRequest.GetParam("startDate")
		endDate := pamRequest.GetParam("endDate")

		averageActiveEnergyPerMinuteFromAllClients := 0.0
		respondingClientCount := 0
		for _, client := range clients {
			log.Printf("Requesting data from %s", client)
			httpRequest, err := http.NewRequest("GET", fmt.Sprintf("http://%s:3001/", client), nil)
			if err != nil {
				log.Println("Error:", err)
				http.Error(w, err.Error(), 500)
				return
			}

			req := middleware.PamRequest{
				Policy:      &policy,
				HttpRequest: httpRequest,
			}

			req.SetParam("startDate", startDate)
			req.SetParam("endDate", endDate)

			pamResp, err := req.Send()
			// TODO: check if the database failed to connect and error properly, might now be fixed
			if err != nil {
				log.Printf("Request to %s produced an error: %s", client, err.Error())
				continue
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
				continue
			}

			resp := pamResp.HttpResponse

			// Read response
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				log.Println("Error reading body of response:", err)
				resp.Body.Close()
				continue
			}

			switch pamResp.ComputationLevel {
			case middleware.CanCompute:
				averageActiveEnergyPerMinute, err := strconv.ParseFloat(string(body), 64)
				if err != nil {
					log.Printf("%s returned a value which could not be parsed as a float, error: %s", client, err)
					continue
				}
				averageActiveEnergyPerMinuteFromAllClients += averageActiveEnergyPerMinute
				respondingClientCount += 1
			case middleware.RawData:
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
				respondingClientCount += 1
				log.Printf("%s gave value %f", client, averageActiveEnergyPerMinute)
			case middleware.NoComputation:
				log.Printf("client: %s could not compute a result", client)
			}

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
	// Create actual function to run
	getAveragePowerConsumptionHandler, err := createGetAveragePowerConsumptionHandler()
	if err != nil {
		log.Fatal(err)
	}
	handler := http.HandlerFunc(getAveragePowerConsumptionHandler)

	// Define computation policy
	computationPolicy := middleware.NewStaticComputationPolicy()
	computationPolicy.Register("/get-average-power-consumption", middleware.CanCompute, handler)

	// Register the composite handler at '/get-average-power-consumption' on port 3002
	http.Handle("/get-average-power-consumption", middleware.PrivacyAwareHandler(computationPolicy))
	log.Println("Listening...")
	err = http.ListenAndServe(":3002", nil)
	if err != nil {
		log.Fatal(err)
	}
}
