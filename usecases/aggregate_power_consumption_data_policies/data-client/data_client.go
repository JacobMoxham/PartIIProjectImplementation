package main

import (
	"errors"
	"fmt"
	"github.com/JacobMoxham/PartIIProjectImplementation/middleware"
	"github.com/joho/sqltocsv"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"
)

const DOCKER = false

func createPowerConsumptionDataHandler() (func(http.ResponseWriter, *http.Request), *middleware.MySqlPrivateDatabase, error) {
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
		CacheTables: true,
	}
	var err error
	if DOCKER {
		err = db.Connect("demouser", "demopassword", "power_consumption", "database", 3306)
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

			resultString, err := sqltocsv.WriteString(rows)
			if err != nil {
				http.Error(w, err.Error(), 500)
				log.Println(err.Error())
				return
			}
			_, err = w.Write([]byte(resultString))
			if err != nil {
				http.Error(w, err.Error(), 500)
				log.Println(err.Error())
				return
			}
		},
		&db, nil
}

func main() {
	// Create actual function to run
	powerConsumptionHandler, db, err := createPowerConsumptionDataHandler()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	handler := http.HandlerFunc(powerConsumptionHandler)

	// Define computation policy
	computationPolicy := middleware.NewStaticComputationPolicy()
	computationPolicy.Register("/", middleware.CanCompute, handler)

	// Register the composite handler at '/' on port 3001
	http.Handle("/", middleware.PrivacyAwareHandler(computationPolicy))
	log.Println("Listening...")
	err = http.ListenAndServe(":3001", nil)
	if err != nil {
		log.Fatal(err)
	}
}
