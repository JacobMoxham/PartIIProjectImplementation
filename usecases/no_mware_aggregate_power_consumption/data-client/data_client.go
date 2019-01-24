package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/sqltocsv"
	"log"
	"net/http"
	_ "net/http/pprof"
	"time"
)

func createPowerConsumptionDataHandler() (func(http.ResponseWriter, *http.Request), *sql.DB, error) {
	db, err := sql.Open("mysql",
		fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=UTC",
			"demouser", "demopassword", "no-mware-database", 3306, "power_consumption"))
	if err != nil {
		return nil, nil, err
	}
	db.SetMaxIdleConns(0)
	db.SetConnMaxLifetime(time.Second * 20)

	if err != nil {
		return nil, nil, err
	}

	return func(w http.ResponseWriter, r *http.Request) {
			// Get date interval from received request
			requestParams := r.URL.Query()
			startDate := requestParams.Get("startDate")
			endDate := requestParams.Get("endDate")

			// Parse as time for validation purposes
			startTime, err := time.Parse("2006-01-02", startDate)
			if err != nil {
				http.Error(w, err.Error(), 200)
			}
			endTime, err := time.Parse("2006-01-02", endDate)
			if err != nil {
				http.Error(w, err.Error(), 200)
			}

			if err != nil {
				http.Error(w, err.Error(), 200)
				return
			}
			queryString := fmt.Sprintf("SELECT datetime, "+
				"global_active_power*1000/60 - sub_metering_1 - sub_metering_2 - sub_metering_3 "+
				"AS active_energy_per_minute "+
				"FROM household_power_consumption "+
				"WHERE datetime BETWEEN \"%s\" AND \"%s\" ", startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))
			rows, err := db.Query(queryString)
			if err != nil {
				http.Error(w, err.Error(), 200)
				return
			}

			defer rows.Close()

			resultString, err := sqltocsv.WriteString(rows)
			if err != nil {
				http.Error(w, err.Error(), 200)
				return
			}
			_, err = w.Write([]byte(resultString))
			if err != nil {
				http.Error(w, err.Error(), 200)
				return
			}
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
