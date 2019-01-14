package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"
)

func createPowerConsumptionDataHandler() (func(http.ResponseWriter, *http.Request), *sql.DB, error) {

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=UTC", "demouser", "demopassword", "database", 3306, "power_consumption"))
	if err != nil {
		return nil, nil, err
	}
	db.SetMaxIdleConns(0)
	db.SetConnMaxLifetime(time.Second * 20)

	return func(w http.ResponseWriter, r *http.Request) {
			log.Println("Request Received")
			rows, err := db.Query(`SELECT datetime, global_active_power*1000/60 - sub_metering_1 - sub_metering_2 - sub_metering_3 AS active_energy_per_minute FROM household_power_consumption LIMIT 10`)
			if err != nil {
				http.Error(w, err.Error(), 200)
				return
			}
			defer rows.Close()

			var (
				datetime              time.Time
				activeEnergyPerMinute float32
			)
			// TODO: write in a better format i.e. JSON
			_, err = w.Write([]byte("Datetime		GlobalActivePower\n"))
			if err != nil {
				http.Error(w, err.Error(), 200)
				return
			}
			for rows.Next() {
				err = rows.Scan(&datetime, &activeEnergyPerMinute)
				if err != nil {
					http.Error(w, err.Error(), 200)
					return
				}

				_, err = w.Write([]byte(fmt.Sprintf("%s	%f\n", datetime.Format("2006-01-02 15:04:05"), activeEnergyPerMinute)))
				if err != nil {
					http.Error(w, err.Error(), 200)
					return
				}
			}
			log.Println("Handler Complete")
		},
		db, nil
}

func main() {
	// Create actual function to run
	powerConsumptionHandler, db, err := createPowerConsumptionDataHandler()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	handler := http.HandlerFunc(powerConsumptionHandler)

	// Register the composite handler at '/' on port 3001
	http.Handle("/", handler)
	log.Println("Listening...")
	err = http.ListenAndServe(":3001", nil)
	if err != nil {
		log.Fatal(err)
	}
}
