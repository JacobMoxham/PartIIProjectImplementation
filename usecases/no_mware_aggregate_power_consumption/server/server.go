package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"strconv"
)

var clients = []string{"no-mware-data-client-raw-data"}

func createGetAveragePowerConsumptionHandler() (func(http.ResponseWriter, *http.Request), error) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get date interval from received request
		requestParams := r.URL.Query()
		startDate := requestParams.Get("startDate")
		endDate := requestParams.Get("endDate")

		averageActiveEnergyPerMinuteFromAllClients := 0.0
		for _, client := range clients {
			httpRequest, _ := http.NewRequest("GET", fmt.Sprintf("http://%s:3001/", client), nil)
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
			reader := csv.NewReader(bytes.NewBuffer(body))
			lines, err := reader.ReadAll()
			if err != nil {
				log.Println("Error:", err)
				http.Error(w, err.Error(), 200)
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
						http.Error(w, err.Error(), 200)
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
		_, err := w.Write([]byte(fmt.Sprintf("%.2f", averageActiveEnergyPerMinuteFromAllClients)))
		if err != nil {
			log.Println("Error:", err)
			http.Error(w, err.Error(), 200)
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

	// Register the composite handler at '/get-average-power-consumption' on port 3002
	http.Handle("/get-average-power-consumption", handler)
	log.Println("Listening...")
	err = http.ListenAndServe(":3002", nil)
	if err != nil {
		log.Fatal(err)
	}
}
