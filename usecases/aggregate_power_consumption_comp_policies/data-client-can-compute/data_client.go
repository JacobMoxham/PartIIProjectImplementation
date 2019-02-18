package main

import (
	"errors"
	"fmt"
	"github.com/JacobMoxham/PartIIProjectImplementation/middleware"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"
)

const DOCKER = true

func createAveragePowerConsumptionHandler() (func(http.ResponseWriter, *http.Request), *middleware.MySqlPrivateDatabase, error) {
	transformsForEntities := make(map[string]map[string]func(interface{}) (interface{}, error))
	transformsForEntities["household_power_consumption"] = map[string]func(arg interface{}) (interface{}, error){"datetime": func(arg interface{}) (interface{}, error) {
		date, ok := arg.(*time.Time)

		if !ok {
			return nil, errors.New("argument could not be asserted as time.Time")
		}
		onlyYear := time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, time.UTC)
		return onlyYear, nil
	}}
	removedColumnsForEntities := map[string][]string{"CentralServer": {}}

	group := &middleware.PrivacyGroup{Name: "CentralServer", Members: map[string]bool{"server": true}}

	db := middleware.MySqlPrivateDatabase{
		StaticDataPolicy: &middleware.StaticDataPolicy{
			PrivacyGroups: []*middleware.PrivacyGroup{group},
			Transforms:    middleware.DataTransforms{group: &middleware.TableOperations{transformsForEntities, removedColumnsForEntities}},
		},
		CacheTables: true,
	}
	var err error
	if DOCKER {
		dbName := "database-can-compute"
		// Support database name as command line argument
		if len(os.Args) > 1 {
			dbName = os.Args[1]
		}
		err = db.Connect("demouser", "demopassword", "power_consumption", dbName, 3306)
	} else {
		err = db.Connect("demouser", "demopassword", "power_consumption", "localhost", 3306)
	}

	if err != nil {
		return nil, nil, err
	}

	return func(w http.ResponseWriter, r *http.Request) {
			pamRequest, err := middleware.BuildPamRequest(r)
			startDate := pamRequest.GetParam("startDate")
			endDate := pamRequest.GetParam("endDate")
			requestPolicy := pamRequest.Policy

			// Parse as time for validation purposes
			startTime, err := time.Parse("2006-01-02", startDate)
			if err != nil {
				http.Error(w, err.Error(), 400)
				log.Println(err.Error())
				return
			}
			endTime, err := time.Parse("2006-01-02", endDate)
			if err != nil {
				http.Error(w, err.Error(), 400)
				log.Println(err.Error())
				return
			}

			queryString := fmt.Sprintf("SELECT datetime, "+
				"global_active_power*1000/60 - sub_metering_1 - sub_metering_2 - sub_metering_3 "+
				"AS active_energy_per_minute "+
				"FROM household_power_consumption "+
				"WHERE datetime BETWEEN \"%s\" AND \"%s\" ", startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))
			rows, err := db.Query(queryString, requestPolicy)
			if err != nil {
				http.Error(w, err.Error(), 500)
				log.Println(err.Error())
				return
			}

			defer rows.Close()

			var (
				datetime              time.Time
				activeEnergyPerMinute float64
			)
			numRows := 0
			totalActiveEnergyPerMinute := 0.0

			for rows.Next() {
				numRows += 1
				rows.Scan(&datetime, &activeEnergyPerMinute)
				totalActiveEnergyPerMinute += activeEnergyPerMinute
			}
			averageActiveEnergyPerMinute := 0.0
			if numRows > 0 {
				averageActiveEnergyPerMinute = totalActiveEnergyPerMinute / float64(numRows)
			}
			_, err = w.Write([]byte(fmt.Sprintf("%.2f", averageActiveEnergyPerMinute)))
			if err != nil {
				log.Println("Error:", err)
				http.Error(w, err.Error(), 500)
				return
			}
		},
		&db, nil
}

func main() {
	// Create actual function to run
	averagePowerConsumptionHandler, db, err := createAveragePowerConsumptionHandler()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	averageDataHandler := http.HandlerFunc(averagePowerConsumptionHandler)

	// Define computation policy
	computationPolicy := middleware.NewStaticComputationPolicy()
	computationPolicy.Register("/", middleware.CanCompute, averageDataHandler)

	// Register the composite handler at '/' on port 3001
	http.Handle("/", middleware.PrivacyAwareHandler(computationPolicy))
	log.Println("Listening...")
	err = http.ListenAndServe(":3001", nil)
	if err != nil {
		log.Fatal(err)
	}
}
