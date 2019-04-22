package main

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/JacobMoxham/PartIIProjectImplementation/middleware"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

type benchResult struct {
	nsPerOp     float64
	bytesperOp  float64
	allocsPerOp float64
}

func main() {
	//runReadBenchmarks()
	runWriteBenchmarks()
}

func runReadBenchmarks() {
	for i, filename := range []string{
		"/home/jacob/PycharmProjects/PartIIProjectDataAnalysis/benchmarks/data/read-benchmarks-no-mware-collection.csv",
		"/home/jacob/PycharmProjects/PartIIProjectDataAnalysis/benchmarks/data/read-benchmarks-caching-collection.csv",
		"/home/jacob/PycharmProjects/PartIIProjectDataAnalysis/benchmarks/data/read-benchmarks-no-caching-collection.csv",
	} {
		benchmarkResults := GetBenchmarkResults(readBenchmarkLists[i])

		f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}

		w := csv.NewWriter(f)
		for _, br := range benchmarkResults {
			record := []string{
				strconv.FormatFloat(br.nsPerOp, 'f', 5, 64),
				strconv.FormatFloat(br.allocsPerOp, 'f', 5, 64),
				strconv.FormatFloat(br.bytesperOp, 'f', 5, 64),
			}
			err := w.Write(record)
			if err != nil {
				panic(err)
			}
		}
		w.Flush()
		err = f.Close()
		if err != nil {
			panic(err)
		}
	}
}

func runWriteBenchmarks() {
	for i, filename := range []string{
		"/home/jacob/PycharmProjects/PartIIProjectDataAnalysis/benchmarks/data/write-benchmarks-no-mware-collection.csv",
		"/home/jacob/PycharmProjects/PartIIProjectDataAnalysis/benchmarks/data/write-benchmarks-mware-collection.csv",
	} {
		benchmarkResults := GetBenchmarkResults(writeBenchmarkLists[i])

		f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}

		w := csv.NewWriter(f)
		for _, br := range benchmarkResults {
			record := []string{
				strconv.FormatFloat(br.nsPerOp, 'f', 5, 64),
				strconv.FormatFloat(br.allocsPerOp, 'f', 5, 64),
				strconv.FormatFloat(br.bytesperOp, 'f', 5, 64),
			}
			err := w.Write(record)
			if err != nil {
				panic(err)
			}
		}
		w.Flush()
		err = f.Close()
		if err != nil {
			panic(err)
		}
	}
}

var readBenchmarkLists = [][]func(*testing.B){noMwareReadBenchmarks, cachingReadBenchmarks, noCachingReadBenchmarks}

var noMwareReadBenchmarks = []func(*testing.B){
	BenchmarkMySQLDatabaseQueryRead_100,
	BenchmarkMySQLDatabaseQueryRead_200,
	BenchmarkMySQLDatabaseQueryRead_300,
	BenchmarkMySQLDatabaseQueryRead_400,
	BenchmarkMySQLDatabaseQueryRead_500,
	BenchmarkMySQLDatabaseQueryRead_600,
	BenchmarkMySQLDatabaseQueryRead_700,
	BenchmarkMySQLDatabaseQueryRead_800,
	BenchmarkMySQLDatabaseQueryRead_900,
	BenchmarkMySQLDatabaseQueryRead_1000,
	BenchmarkMySQLDatabaseQueryRead_2000,
	BenchmarkMySQLDatabaseQueryRead_3000,
	BenchmarkMySQLDatabaseQueryRead_4000,
	BenchmarkMySQLDatabaseQueryRead_5000,
	BenchmarkMySQLDatabaseQueryRead_8000,
	BenchmarkMySQLDatabaseQueryRead_10000,
	BenchmarkMySQLDatabaseQueryRead_15000,
	BenchmarkMySQLDatabaseQueryRead_20000,
	BenchmarkMySQLDatabaseQueryRead_25000,
}
var cachingReadBenchmarks = []func(*testing.B){
	BenchmarkMySQLPrivateDatabase_Query_Read_Caching_100,
	BenchmarkMySQLPrivateDatabase_Query_Read_Caching_200,
	BenchmarkMySQLPrivateDatabase_Query_Read_Caching_300,
	BenchmarkMySQLPrivateDatabase_Query_Read_Caching_400,
	BenchmarkMySQLPrivateDatabase_Query_Read_Caching_500,
	BenchmarkMySQLPrivateDatabase_Query_Read_Caching_600,
	BenchmarkMySQLPrivateDatabase_Query_Read_Caching_700,
	BenchmarkMySQLPrivateDatabase_Query_Read_Caching_800,
	BenchmarkMySQLPrivateDatabase_Query_Read_Caching_900,
	BenchmarkMySQLPrivateDatabase_Query_Read_Caching_1000,
	BenchmarkMySQLPrivateDatabase_Query_Read_Caching_2000,
	BenchmarkMySQLPrivateDatabase_Query_Read_Caching_3000,
	BenchmarkMySQLPrivateDatabase_Query_Read_Caching_4000,
	BenchmarkMySQLPrivateDatabase_Query_Read_Caching_5000,
	BenchmarkMySQLPrivateDatabase_Query_Read_Caching_8000,
	BenchmarkMySQLPrivateDatabase_Query_Read_Caching_10000,
	BenchmarkMySQLPrivateDatabase_Query_Read_Caching_15000,
	BenchmarkMySQLPrivateDatabase_Query_Read_Caching_20000,
	BenchmarkMySQLPrivateDatabase_Query_Read_Caching_25000,
}
var noCachingReadBenchmarks = []func(*testing.B){
	BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_100,
	BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_200,
	BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_300,
	BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_400,
	BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_500,
	BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_600,
	BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_700,
	BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_800,
	BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_900,
	BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_1000,
	BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_2000,
	BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_3000,
	BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_4000,
	BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_5000,
	BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_8000,
	BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_10000,
	BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_15000,
	BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_20000,
	BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_25000,
}

var writeBenchmarkLists = [][]func(*testing.B){noMwareWriteBenchmarks, writeBenchmarks}

var noMwareWriteBenchmarks = []func(*testing.B){
	BenchmarkMySQLDatabase_Exec_Write_100,
	BenchmarkMySQLDatabase_Exec_Write_200,
	BenchmarkMySQLDatabase_Exec_Write_300,
	BenchmarkMySQLDatabase_Exec_Write_400,
	BenchmarkMySQLDatabase_Exec_Write_500,
	BenchmarkMySQLDatabase_Exec_Write_600,
	BenchmarkMySQLDatabase_Exec_Write_700,
	BenchmarkMySQLDatabase_Exec_Write_800,
	BenchmarkMySQLDatabase_Exec_Write_900,
	BenchmarkMySQLDatabase_Exec_Write_1000,
	BenchmarkMySQLDatabase_Exec_Write_2000,
	BenchmarkMySQLDatabase_Exec_Write_3000,
	BenchmarkMySQLDatabase_Exec_Write_4000,
	BenchmarkMySQLDatabase_Exec_Write_5000,
	BenchmarkMySQLDatabase_Exec_Write_8000,
	BenchmarkMySQLDatabase_Exec_Write_10000,
	BenchmarkMySQLDatabase_Exec_Write_15000,
	BenchmarkMySQLDatabase_Exec_Write_20000,
	BenchmarkMySQLDatabase_Exec_Write_25000,
}

var writeBenchmarks = []func(*testing.B){
	BenchmarkMySQLPrivateDatabase_Exec_Write_100,
	BenchmarkMySQLPrivateDatabase_Exec_Write_200,
	BenchmarkMySQLPrivateDatabase_Exec_Write_300,
	BenchmarkMySQLPrivateDatabase_Exec_Write_400,
	BenchmarkMySQLPrivateDatabase_Exec_Write_500,
	BenchmarkMySQLPrivateDatabase_Exec_Write_600,
	BenchmarkMySQLPrivateDatabase_Exec_Write_700,
	BenchmarkMySQLPrivateDatabase_Exec_Write_800,
	BenchmarkMySQLPrivateDatabase_Exec_Write_900,
	BenchmarkMySQLPrivateDatabase_Exec_Write_1000,
	BenchmarkMySQLPrivateDatabase_Exec_Write_2000,
	BenchmarkMySQLPrivateDatabase_Exec_Write_3000,
	BenchmarkMySQLPrivateDatabase_Exec_Write_4000,
	BenchmarkMySQLPrivateDatabase_Exec_Write_5000,
	BenchmarkMySQLPrivateDatabase_Exec_Write_8000,
	BenchmarkMySQLPrivateDatabase_Exec_Write_10000,
	BenchmarkMySQLPrivateDatabase_Exec_Write_15000,
	BenchmarkMySQLPrivateDatabase_Exec_Write_20000,
	BenchmarkMySQLPrivateDatabase_Exec_Write_25000,
}

func GetBenchmarkResults(benchmarks []func(*testing.B)) []*benchResult {
	var benchResults []*benchResult
	for i := 0; i < len(benchmarks); i++ {
		benchResults = append(benchResults, new(benchResult))
	}

	for j, f := range benchmarks {
		backoff := 10 * time.Second
		for {
			benchmarkResult := testing.Benchmark(f)
			log.Printf("%+v", benchmarkResult)
			if benchmarkResult.NsPerOp() == 0 {
				log.Println("Retry")
				time.Sleep(backoff)
				backoff *= 2
				continue
			} else {
				benchResults[j].nsPerOp = float64(benchmarkResult.NsPerOp())
				benchResults[j].bytesperOp = float64(benchmarkResult.AllocedBytesPerOp())
				benchResults[j].allocsPerOp = float64(benchmarkResult.AllocsPerOp())
				break
			}
		}
	}

	return benchResults
}

func validFuncMap() map[string]middleware.TableTransform {
	funcMap := make(map[string]middleware.TableTransform)
	funcMap["TestGroup"] = make(middleware.TableTransform)

	funcMap["TestGroup"]["dob"] = func(arg interface{}) (interface{}, bool, error) {
		date, ok := arg.(*time.Time)

		if !ok {
			return nil, true, errors.New("argument could not be asserted as time.Time")
		}
		onlyYear := time.Date(date.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		return onlyYear, false, nil
	}
	funcMap["TestGroup"]["name"] = func(arg interface{}) (interface{}, bool, error) {
		name, ok := arg.(*string)

		if !ok {
			return nil, true, errors.New("argument could not be asserted as string")
		}

		hiddenName := ""
		for i, c := range *name {
			if i > 2 {
				hiddenName += "*"
			} else {
				hiddenName += fmt.Sprintf("%c", c)
			}
		}
		return hiddenName, false, nil
	}

	return funcMap
}

var result *sql.Rows

func benchmarkMySQLDatabaseQuery(b *testing.B, db *sql.DB, queryString string) *sql.Rows {
	r, err := db.Query(queryString)
	if err != nil {
		b.Error(err.Error())
	}
	return r
}

func benchmarkMySQLDatabaseQueryRead(b *testing.B, queryString string) {
	b.StopTimer()
	db, err := sql.Open("mysql",
		fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=UTC",
			"demouser", "demopassword", "127.0.0.1", 3306, "power_consumption"))
	if err != nil {
		b.Error(err.Error())
	}
	db.SetMaxIdleConns(0)
	db.SetConnMaxLifetime(time.Second * 20)
	b.StartTimer()

	var r *sql.Rows
	for i := 0; i < b.N; i++ {
		r = benchmarkMySQLDatabaseQuery(b, db, queryString)
		if r == nil {
			b.Log("Nil result")
			return
		}
		err = r.Close()
		if err != nil {
			b.Log(err.Error())
		}
	}
	result = r

}

func BenchmarkMySQLDatabaseQueryRead_100(b *testing.B) {
	benchmarkMySQLDatabaseQueryRead(b, "SELECT * FROM power_cons_100")
}

func BenchmarkMySQLDatabaseQueryRead_200(b *testing.B) {
	benchmarkMySQLDatabaseQueryRead(b, "SELECT * FROM power_cons_200")
}

func BenchmarkMySQLDatabaseQueryRead_300(b *testing.B) {
	benchmarkMySQLDatabaseQueryRead(b, "SELECT * FROM power_cons_300")
}

func BenchmarkMySQLDatabaseQueryRead_400(b *testing.B) {
	benchmarkMySQLDatabaseQueryRead(b, "SELECT * FROM power_cons_400")
}

func BenchmarkMySQLDatabaseQueryRead_500(b *testing.B) {
	benchmarkMySQLDatabaseQueryRead(b, "SELECT * FROM power_cons_500")
}

func BenchmarkMySQLDatabaseQueryRead_600(b *testing.B) {
	benchmarkMySQLDatabaseQueryRead(b, "SELECT * FROM power_cons_600")
}

func BenchmarkMySQLDatabaseQueryRead_700(b *testing.B) {
	benchmarkMySQLDatabaseQueryRead(b, "SELECT * FROM power_cons_700")
}

func BenchmarkMySQLDatabaseQueryRead_800(b *testing.B) {
	benchmarkMySQLDatabaseQueryRead(b, "SELECT * FROM power_cons_800")
}

func BenchmarkMySQLDatabaseQueryRead_900(b *testing.B) {
	benchmarkMySQLDatabaseQueryRead(b, "SELECT * FROM power_cons_900")
}

func BenchmarkMySQLDatabaseQueryRead_1000(b *testing.B) {
	benchmarkMySQLDatabaseQueryRead(b, "SELECT * FROM power_cons_1000")
}

func BenchmarkMySQLDatabaseQueryRead_2000(b *testing.B) {
	benchmarkMySQLDatabaseQueryRead(b, "SELECT * FROM power_cons_2000")
}

func BenchmarkMySQLDatabaseQueryRead_3000(b *testing.B) {
	benchmarkMySQLDatabaseQueryRead(b, "SELECT * FROM power_cons_3000")
}

func BenchmarkMySQLDatabaseQueryRead_4000(b *testing.B) {
	benchmarkMySQLDatabaseQueryRead(b, "SELECT * FROM power_cons_4000")
}

func BenchmarkMySQLDatabaseQueryRead_5000(b *testing.B) {
	benchmarkMySQLDatabaseQueryRead(b, "SELECT * FROM power_cons_5000")
}

func BenchmarkMySQLDatabaseQueryRead_8000(b *testing.B) {
	benchmarkMySQLDatabaseQueryRead(b, "SELECT * FROM power_cons_8000")
}

func BenchmarkMySQLDatabaseQueryRead_10000(b *testing.B) {
	benchmarkMySQLDatabaseQueryRead(b, "SELECT * FROM power_cons_10000")
}

func BenchmarkMySQLDatabaseQueryRead_15000(b *testing.B) {
	benchmarkMySQLDatabaseQueryRead(b, "SELECT * FROM power_cons_15000")
}

func BenchmarkMySQLDatabaseQueryRead_20000(b *testing.B) {
	benchmarkMySQLDatabaseQueryRead(b, "SELECT * FROM power_cons_20000")
}

func BenchmarkMySQLDatabaseQueryRead_25000(b *testing.B) {
	benchmarkMySQLDatabaseQueryRead(b, "SELECT * FROM power_cons_25000")
}

func benchmarkMySQLPrivateDatabaseQueryReadNoCaching(b *testing.B, queryString string) {
	b.StopTimer()
	funcMap := validFuncMap()
	colMap := map[string][]string{"TestGroup": {}}

	group := &middleware.PrivacyGroup{"TestGroup", map[string]bool{"jacob": true}}

	staticDataPolicy := middleware.NewStaticDataPolicy([]*middleware.PrivacyGroup{group},
		middleware.DataTransforms{group: &middleware.TableOperations{funcMap, colMap}})

	db := middleware.MySQLPrivateDatabase{
		DataPolicy:  staticDataPolicy,
		CacheTables: false,
	}
	err := db.Connect("demouser", "demopassword", "power_consumption", "127.0.0.1", 3306)
	if err != nil {
		b.Error(err.Error())
	}
	b.StartTimer()

	var r *sql.Rows
	for i := 0; i < b.N; i++ {
		r = benchmarkMySQLPrivateDatabaseQuery(b, db, queryString)
		if r == nil {
			b.Log("Nil result")
			return
		}
		err = r.Close()
		if err != nil {
			b.Log(err.Error())
		}
	}
	result = r
}

func BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_100(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadNoCaching(b, "SELECT * FROM power_cons_100")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_200(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadNoCaching(b, "SELECT * FROM power_cons_200")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_300(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadNoCaching(b, "SELECT * FROM power_cons_300")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_400(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadNoCaching(b, "SELECT * FROM power_cons_400")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_500(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadNoCaching(b, "SELECT * FROM power_cons_500")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_600(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadNoCaching(b, "SELECT * FROM power_cons_600")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_700(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadNoCaching(b, "SELECT * FROM power_cons_700")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_800(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadNoCaching(b, "SELECT * FROM power_cons_800")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_900(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadNoCaching(b, "SELECT * FROM power_cons_900")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_1000(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadNoCaching(b, "SELECT * FROM power_cons_1000")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_2000(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadNoCaching(b, "SELECT * FROM power_cons_2000")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_3000(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadNoCaching(b, "SELECT * FROM power_cons_3000")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_4000(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadNoCaching(b, "SELECT * FROM power_cons_4000")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_5000(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadNoCaching(b, "SELECT * FROM power_cons_5000")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_8000(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadNoCaching(b, "SELECT * FROM power_cons_8000")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_10000(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadNoCaching(b, "SELECT * FROM power_cons_10000")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_15000(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadNoCaching(b, "SELECT * FROM power_cons_15000")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_20000(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadNoCaching(b, "SELECT * FROM power_cons_20000")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_25000(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadNoCaching(b, "SELECT * FROM power_cons_25000")
}

func benchmarkMySQLPrivateDatabaseQuery(b *testing.B, db middleware.MySQLPrivateDatabase, queryString string) *sql.Rows {
	r, err := db.Query(queryString, &middleware.RequestPolicy{"jacob", middleware.Local, true})
	if err != nil {
		b.Error(err.Error())
	}
	return r
}

func benchmarkMySQLPrivateDatabaseQueryReadCaching(b *testing.B, queryString string) {
	b.StopTimer()
	funcMap := validFuncMap()
	colMap := map[string][]string{"TestGroup": {}}

	group := &middleware.PrivacyGroup{"TestGroup", map[string]bool{"jacob": true}}

	staticDataPolicy := middleware.NewStaticDataPolicy([]*middleware.PrivacyGroup{group},
		middleware.DataTransforms{group: &middleware.TableOperations{funcMap, colMap}})

	db := middleware.MySQLPrivateDatabase{
		DataPolicy:  staticDataPolicy,
		CacheTables: true,
	}
	err := db.Connect("demouser", "demopassword", "power_consumption", "127.0.0.1", 3306)
	if err != nil {
		b.Error(err.Error())
	}
	// Make the query once so we know we have a cached version of the table
	_, err = db.Query(queryString, &middleware.RequestPolicy{"jacob", middleware.Local, true})
	if err != nil {
		b.Error(err.Error())
	}
	b.StartTimer()

	var r *sql.Rows
	for i := 0; i < b.N; i++ {
		r = benchmarkMySQLPrivateDatabaseQuery(b, db, queryString)
		if r == nil {
			b.Log("Nil result")
			return
		}
		err = r.Close()
		if err != nil {
			b.Log(err.Error())
		}
	}
	result = r
}

func BenchmarkMySQLPrivateDatabase_Query_Read_Caching_100(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadCaching(b, "SELECT * FROM power_cons_100")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_Caching_200(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadCaching(b, "SELECT * FROM power_cons_200")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_Caching_300(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadCaching(b, "SELECT * FROM power_cons_300")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_Caching_400(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadCaching(b, "SELECT * FROM power_cons_400")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_Caching_500(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadCaching(b, "SELECT * FROM power_cons_500")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_Caching_600(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadCaching(b, "SELECT * FROM power_cons_600")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_Caching_700(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadCaching(b, "SELECT * FROM power_cons_700")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_Caching_800(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadCaching(b, "SELECT * FROM power_cons_800")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_Caching_900(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadCaching(b, "SELECT * FROM power_cons_900")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_Caching_1000(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadCaching(b, "SELECT * FROM power_cons_1000")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_Caching_2000(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadCaching(b, "SELECT * FROM power_cons_2000")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_Caching_3000(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadCaching(b, "SELECT * FROM power_cons_3000")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_Caching_4000(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadCaching(b, "SELECT * FROM power_cons_4000")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_Caching_5000(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadCaching(b, "SELECT * FROM power_cons_5000")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_Caching_8000(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadCaching(b, "SELECT * FROM power_cons_8000")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_Caching_10000(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadCaching(b, "SELECT * FROM power_cons_10000")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_Caching_15000(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadCaching(b, "SELECT * FROM power_cons_15000")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_Caching_20000(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadCaching(b, "SELECT * FROM power_cons_20000")
}

func BenchmarkMySQLPrivateDatabase_Query_Read_Caching_25000(b *testing.B) {
	benchmarkMySQLPrivateDatabaseQueryReadCaching(b, "SELECT * FROM power_cons_25000")
}

func benchmarkMySQLDatabaseExecWrite(b *testing.B, execString string, args ...interface{}) {
	b.StopTimer()
	db, err := sql.Open("mysql",
		fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=UTC",
			"demouser", "demopassword", "127.0.0.1", 3306, "store1"))
	if err != nil {
		b.Error(err.Error())
	}
	if db == nil {
		b.Error(errors.New("DB is nil"))
	}

	db.SetMaxIdleConns(0)
	db.SetConnMaxLifetime(time.Second * 20)
	b.StartTimer()

	_, err = db.Exec(execString, args...)
	if err != nil {
		b.Error(err.Error())
	}

	b.StopTimer()
	err = db.Close()
	if err != nil {
		b.Error(err.Error())
	}
	b.StartTimer()
}

var row = []interface{}{
	"steve",
	time.Date(1996, 2, 7, 0, 0, 0, 0, time.UTC),
}

func getNRows(numRows int) (string, []interface{}) {
	rowsToWrite := ""
	var rowArguments []interface{}
	for i := 0; i < numRows; i++ {
		// Add rows to string
		rowsToWrite += "("
		for i := 0; i < len(row); i++ {
			rowsToWrite += "?, "
		}
		rowsToWrite = strings.TrimSuffix(rowsToWrite, ", ")
		rowsToWrite += "), "
		rowArguments = append(rowArguments, row...)
	}

	// Remove the last comma and space
	rowsToWrite = strings.TrimSuffix(rowsToWrite, ", ")

	return rowsToWrite, rowArguments
}

func benchmarkMySQLDatabaseExecWriteN(b *testing.B, numRows int) {
	b.StopTimer()
	rowsToWrite, rowArguments := getNRows(numRows)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		benchmarkMySQLDatabaseExecWrite(b, `INSERT INTO people (name, dob) VALUES `+rowsToWrite, rowArguments...)
	}
}

func BenchmarkMySQLDatabase_Exec_Write_100(b *testing.B) {
	benchmarkMySQLDatabaseExecWriteN(b, 100)
}

func BenchmarkMySQLDatabase_Exec_Write_200(b *testing.B) {
	benchmarkMySQLDatabaseExecWriteN(b, 200)
}

func BenchmarkMySQLDatabase_Exec_Write_300(b *testing.B) {
	benchmarkMySQLDatabaseExecWriteN(b, 300)
}

func BenchmarkMySQLDatabase_Exec_Write_400(b *testing.B) {
	benchmarkMySQLDatabaseExecWriteN(b, 400)
}

func BenchmarkMySQLDatabase_Exec_Write_500(b *testing.B) {
	benchmarkMySQLDatabaseExecWriteN(b, 500)
}

func BenchmarkMySQLDatabase_Exec_Write_600(b *testing.B) {
	benchmarkMySQLDatabaseExecWriteN(b, 600)
}

func BenchmarkMySQLDatabase_Exec_Write_700(b *testing.B) {
	benchmarkMySQLDatabaseExecWriteN(b, 700)
}

func BenchmarkMySQLDatabase_Exec_Write_800(b *testing.B) {
	benchmarkMySQLDatabaseExecWriteN(b, 800)
}

func BenchmarkMySQLDatabase_Exec_Write_900(b *testing.B) {
	benchmarkMySQLDatabaseExecWriteN(b, 900)
}

func BenchmarkMySQLDatabase_Exec_Write_1000(b *testing.B) {
	benchmarkMySQLDatabaseExecWriteN(b, 1000)
}

func BenchmarkMySQLDatabase_Exec_Write_2000(b *testing.B) {
	benchmarkMySQLDatabaseExecWriteN(b, 2000)
}

func BenchmarkMySQLDatabase_Exec_Write_3000(b *testing.B) {
	benchmarkMySQLDatabaseExecWriteN(b, 3000)
}

func BenchmarkMySQLDatabase_Exec_Write_4000(b *testing.B) {
	benchmarkMySQLDatabaseExecWriteN(b, 4000)
}

func BenchmarkMySQLDatabase_Exec_Write_5000(b *testing.B) {
	benchmarkMySQLDatabaseExecWriteN(b, 5000)
}

func BenchmarkMySQLDatabase_Exec_Write_8000(b *testing.B) {
	benchmarkMySQLDatabaseExecWriteN(b, 8000)
}

func BenchmarkMySQLDatabase_Exec_Write_10000(b *testing.B) {
	benchmarkMySQLDatabaseExecWriteN(b, 10000)
}

func BenchmarkMySQLDatabase_Exec_Write_15000(b *testing.B) {
	benchmarkMySQLDatabaseExecWriteN(b, 15000)
}

func BenchmarkMySQLDatabase_Exec_Write_20000(b *testing.B) {
	benchmarkMySQLDatabaseExecWriteN(b, 20000)
}

func BenchmarkMySQLDatabase_Exec_Write_25000(b *testing.B) {
	benchmarkMySQLDatabaseExecWriteN(b, 25000)
}

func benchmarkMySQLPrivateDatabaseExecWrite(b *testing.B, execString string, args ...interface{}) {
	b.StopTimer()
	funcMap := validFuncMap()
	colMap := map[string][]string{}

	group := &middleware.PrivacyGroup{"TestGroup", map[string]bool{"jacob": true}}

	staticDataPolicy := middleware.NewStaticDataPolicy([]*middleware.PrivacyGroup{group},
		middleware.DataTransforms{group: &middleware.TableOperations{funcMap, colMap}})

	db := middleware.MySQLPrivateDatabase{
		DataPolicy:  staticDataPolicy,
		CacheTables: false,
	}
	err := db.Connect("demouser", "demopassword", "store1", "127.0.0.1", 3306)
	if err != nil {
		b.Error(err.Error())
	}
	b.StartTimer()

	_, err = db.Exec(execString,
		&middleware.RequestPolicy{"jacob", middleware.Local, true},
		args...)
	if err != nil {
		b.Error(err.Error())
	}

	b.StopTimer()
	err = db.Close()
	if err != nil {
		b.Error(err.Error())
	}
	b.StartTimer()
}

func benchmarkMySQLPrivateDatabaseEvecWriteN(b *testing.B, numRows int) {
	b.StopTimer()
	rowsToWrite, rowArguments := getNRows(numRows)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		benchmarkMySQLPrivateDatabaseExecWrite(b, `INSERT INTO people (name, dob) VALUES `+rowsToWrite, rowArguments...)
	}
}

func BenchmarkMySQLPrivateDatabase_Exec_Write_100(b *testing.B) {
	benchmarkMySQLPrivateDatabaseEvecWriteN(b, 100)
}

func BenchmarkMySQLPrivateDatabase_Exec_Write_200(b *testing.B) {
	benchmarkMySQLPrivateDatabaseEvecWriteN(b, 200)
}

func BenchmarkMySQLPrivateDatabase_Exec_Write_300(b *testing.B) {
	benchmarkMySQLPrivateDatabaseEvecWriteN(b, 300)
}

func BenchmarkMySQLPrivateDatabase_Exec_Write_400(b *testing.B) {
	benchmarkMySQLPrivateDatabaseEvecWriteN(b, 400)
}

func BenchmarkMySQLPrivateDatabase_Exec_Write_500(b *testing.B) {
	benchmarkMySQLPrivateDatabaseEvecWriteN(b, 500)
}

func BenchmarkMySQLPrivateDatabase_Exec_Write_600(b *testing.B) {
	benchmarkMySQLPrivateDatabaseEvecWriteN(b, 600)
}

func BenchmarkMySQLPrivateDatabase_Exec_Write_700(b *testing.B) {
	benchmarkMySQLPrivateDatabaseEvecWriteN(b, 700)
}

func BenchmarkMySQLPrivateDatabase_Exec_Write_800(b *testing.B) {
	benchmarkMySQLPrivateDatabaseEvecWriteN(b, 800)
}

func BenchmarkMySQLPrivateDatabase_Exec_Write_900(b *testing.B) {
	benchmarkMySQLPrivateDatabaseEvecWriteN(b, 900)
}

func BenchmarkMySQLPrivateDatabase_Exec_Write_1000(b *testing.B) {
	benchmarkMySQLPrivateDatabaseEvecWriteN(b, 1000)
}

func BenchmarkMySQLPrivateDatabase_Exec_Write_2000(b *testing.B) {
	benchmarkMySQLPrivateDatabaseEvecWriteN(b, 2000)
}

func BenchmarkMySQLPrivateDatabase_Exec_Write_3000(b *testing.B) {
	benchmarkMySQLPrivateDatabaseEvecWriteN(b, 3000)
}

func BenchmarkMySQLPrivateDatabase_Exec_Write_4000(b *testing.B) {
	benchmarkMySQLPrivateDatabaseEvecWriteN(b, 4000)
}

func BenchmarkMySQLPrivateDatabase_Exec_Write_5000(b *testing.B) {
	benchmarkMySQLPrivateDatabaseEvecWriteN(b, 5000)
}

func BenchmarkMySQLPrivateDatabase_Exec_Write_8000(b *testing.B) {
	benchmarkMySQLPrivateDatabaseEvecWriteN(b, 8000)
}

func BenchmarkMySQLPrivateDatabase_Exec_Write_10000(b *testing.B) {
	benchmarkMySQLPrivateDatabaseEvecWriteN(b, 10000)
}

func BenchmarkMySQLPrivateDatabase_Exec_Write_15000(b *testing.B) {
	benchmarkMySQLPrivateDatabaseEvecWriteN(b, 15000)
}

func BenchmarkMySQLPrivateDatabase_Exec_Write_20000(b *testing.B) {
	benchmarkMySQLPrivateDatabaseEvecWriteN(b, 20000)
}

func BenchmarkMySQLPrivateDatabase_Exec_Write_25000(b *testing.B) {
	benchmarkMySQLPrivateDatabaseEvecWriteN(b, 25000)
}
