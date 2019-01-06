package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/JacobMoxham/PartIIProjectImplementation/middleware"
	"github.com/justinas/alice"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

func createThermoDataHandler() (func(http.ResponseWriter, *http.Request), *middleware.MySqlPrivateDatabase) {
	transformsForEntities := make(map[string]func(interface{}) (interface{}, error))
	transformsForEntities["dob"] = func(arg interface{}) (interface{}, error) {
		date, ok := arg.(*time.Time)

		if !ok {
			return nil, errors.New("argument could not be asserted as time.Time")
		}
		onlyYear := time.Date(date.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		return onlyYear, nil
	}
	transformsForEntities["name"] = func(arg interface{}) (interface{}, error) {
		name, ok := arg.(*string)

		if !ok {
			return nil, errors.New("argument could not be asserted as string")
		}

		hiddenName := ""
		for i, c := range *name {
			if i > 2 {
				hiddenName += "*"
			} else {
				hiddenName += fmt.Sprintf("%c", c)
			}
		}
		return hiddenName, nil
	}
	removedColumnsForEntities := map[string][]string{"CentralServer": []string{}}

	group := &middleware.PrivacyGroup{Name: "CentralServer", Members: map[string]bool{"server": true}}

	db := middleware.MySqlPrivateDatabase{
		StaticDataPolicy: &middleware.StaticDataPolicy{
			PrivacyGroups: []*middleware.PrivacyGroup{group},
			Transforms:    middleware.DataTransforms{group: &middleware.TableOperations{transformsForEntities, removedColumnsForEntities}},
		},
	}
	db.Connect("demouser", "demopassword", "power_consumption")

	return func(w http.ResponseWriter, r *http.Request) {
			log.Println("PamRequest Received")
			requestPolicy, err := middleware.BuildRequestPolicy(r)
			if err != nil {
				http.Error(w, err.Error(), 500)
			}
			//queryString := `SELECT datetime, global_active_power*1000/60 - sub_metering_1 - sub_metering_2 - sub_metering_3 AS active_energy_per_minute from household_power_consumption`
			rows, err := db.Query(`SELECT datetime, global_active_power*1000/60 - sub_metering_1 - sub_metering_2 - sub_metering_3 AS active_energy_per_minute FROM household_power_consumption`, requestPolicy)
			if err != nil {
				http.Error(w, err.Error(), 200)
			}

			var (
				datetime              time.Time
				activeEnergyPerMinute float32
			)
			_, err = w.Write([]byte("Datetime		GlobalActivePower\n"))
			if err != nil {
				http.Error(w, err.Error(), 200)
			}
			for rows.Next() {
				err = rows.Scan(&datetime, &activeEnergyPerMinute)
				if err != nil {
					http.Error(w, err.Error(), 200)
				}

				_, err = w.Write([]byte(fmt.Sprintf("%s	%f\n", datetime.Format("2006-01-02 15:04:05"), activeEnergyPerMinute)))
				if err != nil {
					http.Error(w, err.Error(), 200)
				}
			}
			rows.Close()
		},
		&db
}

func centralServer() {
	policy := middleware.RequestPolicy{
		RequesterID:                 "server",
		PreferredProcessingLocation: middleware.Remote,
		HasAllRequiredData:          false,
	}
	httpRequest, _ := http.NewRequest("GET", "http://127.0.0.1:3001/", nil)
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

	log.Println(fmt.Sprintf("%s", body))
}

func thermometerDevice() {
	// Create actual function to run
	thermoDataHandler, db := createThermoDataHandler()
	defer db.Close()
	finalHandler := http.HandlerFunc(thermoDataHandler)

	// Define computation policy
	computationPolicy := middleware.NewStaticComputationPolicy()
	computationPolicy.Register("/", middleware.CanCompute)

	// Chain together "other" middlewares
	handlers := alice.New(middleware.PrivacyAwareHandler(computationPolicy)).Then(finalHandler)

	// Register the composite handler at '/' on port 3001
	http.Handle("/", handlers)
	log.Println("Listening...")
	http.ListenAndServe(":3001", nil)
}

func power_main() {
	var wg sync.WaitGroup
	wg.Add(2)
	// TODO run many using docker each with different ips which the central server can be assumed to know
	go thermometerDevice()
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Press a key to send start central server: ")
	reader.ReadString('\n')
	go centralServer()
	wg.Wait()
}
