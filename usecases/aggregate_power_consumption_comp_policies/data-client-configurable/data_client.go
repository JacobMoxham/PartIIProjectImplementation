package main

import (
	"flag"
	"fmt"
	"github.com/JacobMoxham/PartIIProjectImplementation/middleware"
	"github.com/joho/sqltocsv"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"
)

const DOCKER = true

func createPowerConsumptionRawDataHandler() (func(http.ResponseWriter, *http.Request), *middleware.MySQLPrivateDatabase, error) {
	transformsForEntities := make(map[string]middleware.TableTransform)
	removedColumnsForEntities := map[string][]string{"CentralServer": {}}

	group := middleware.NewPrivacyGroup("CentralServer")
	group.Add("server")
	groups := []*middleware.PrivacyGroup{group}
	transforms := middleware.DataTransforms{group: &middleware.TableOperations{transformsForEntities, removedColumnsForEntities}}

	db := middleware.MySQLPrivateDatabase{
		DataPolicy:  middleware.NewStaticDataPolicy(groups, transforms),
		CacheTables: true,
	}

	var err error
	if DOCKER {
		dbName := "database-both"
		// Support database name as command line argument
		if len(flag.Args()) > 0 {
			dbName = flag.Args()[0]
		}
		err = db.Connect("demouser", "demopassword", "power_consumption", dbName, 3306)
	} else {
		err = db.Connect("demouser", "demopassword", "power_consumption", "localhost", 3306)
	}

	if err != nil {
		return nil, nil, err
	}

	db.SetMaxIdleConns(100)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Second * 100)

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

			// Use double close pattern as rows.Close() is idempotent
			if rows.Err() != nil {
				http.Error(w, err.Error(), 500)
				log.Println(err.Error())
				return
			}

			if err := rows.Close(); err != nil {
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
func createAveragePowerConsumptionHandler() (func(http.ResponseWriter, *http.Request), *middleware.MySQLPrivateDatabase, error) {
	transformsForEntities := make(map[string]middleware.TableTransform)
	removedColumnsForEntities := map[string][]string{"CentralServer": {}}

	group := middleware.NewPrivacyGroup("CentralServer")
	group.Add("server")

	groups := []*middleware.PrivacyGroup{group}
	transforms := middleware.DataTransforms{group: &middleware.TableOperations{transformsForEntities, removedColumnsForEntities}}

	db := middleware.MySQLPrivateDatabase{
		DataPolicy:  middleware.NewStaticDataPolicy(groups, transforms),
		CacheTables: true,
	}

	var err error
	if DOCKER {
		dbName := "database-both"
		// Support database name as command line argument
		if len(flag.Args()) > 0 {
			dbName = flag.Args()[0]
		}
		log.Printf("Connecting to database: %s", dbName)
		err = db.Connect("demouser", "demopassword", "power_consumption", dbName, 3306)
	} else {
		log.Printf("Connecting to local database")

		err = db.Connect("demouser", "demopassword", "power_consumption", "localhost", 3306)
	}

	if err != nil {
		return nil, nil, err
	}

	db.SetMaxIdleConns(100)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Second * 20)

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
	// Get flags
	canCompute := flag.Bool("can-compute", false, "determines whether we register a CanCompute handler")
	rawData := flag.Bool("raw-data", false, "determines whether we register a Raw Data handler")
	flag.Parse()

	// Define computation policy
	computationPolicy := middleware.NewStaticComputationPolicy()
	if canCompute != nil && *canCompute {
		averagePowerConsumptionHandler, db, err := createAveragePowerConsumptionHandler()
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		averageDataHandler := http.HandlerFunc(averagePowerConsumptionHandler)
		computationPolicy.Register("/", middleware.CanCompute, averageDataHandler)
	}
	if rawData != nil && *rawData {
		powerConsumptionRawDataHandler, db, err := createPowerConsumptionRawDataHandler()
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
		rawDataHandler := http.HandlerFunc(powerConsumptionRawDataHandler)

		computationPolicy.Register("/", middleware.RawData, rawDataHandler)
	}

	// Register the composite handler at '/' on port 3001
	http.Handle("/", middleware.PolicyAwareHandler(computationPolicy))
	log.Println("Listening...")
	err := http.ListenAndServe(":3001", nil)
	if err != nil {
		log.Fatal(err)
	}
}
