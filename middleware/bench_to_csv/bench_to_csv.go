package main

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/JacobMoxham/PartIIProjectImplementation/middleware"
	"os"
	"strconv"
	"testing"
	"time"
)

func main() {
	benchmarkResults := GetBenchmarkResults()

	f, err := os.Create("/home/jacob/PycharmProjects/PartIIProjectDataAnalysis/benchmarks/data/read-benchmarks.csv")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	w := csv.NewWriter(f)
	for _, br := range benchmarkResults {
		record := []string{strconv.FormatInt(br.NsPerOp(), 10),
			strconv.FormatInt(br.AllocsPerOp(), 10),
			strconv.FormatInt(br.AllocedBytesPerOp(), 10),
		}
		err := w.Write(record)
		if err != nil {
			panic(err)
		}
	}
	w.Flush()
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
		err = r.Close()
		if err != nil {
			b.Error(err.Error())
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
		err = r.Close()
		if err != nil {
			b.Error(err.Error())
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
		err = r.Close()
		if err != nil {
			b.Error(err.Error())
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

func GetBenchmarkResults() []testing.BenchmarkResult {
	return []testing.BenchmarkResult{
		testing.Benchmark(BenchmarkMySQLDatabaseQueryRead_100),
		testing.Benchmark(BenchmarkMySQLDatabaseQueryRead_200),
		testing.Benchmark(BenchmarkMySQLDatabaseQueryRead_300),
		testing.Benchmark(BenchmarkMySQLDatabaseQueryRead_400),
		testing.Benchmark(BenchmarkMySQLDatabaseQueryRead_500),
		testing.Benchmark(BenchmarkMySQLDatabaseQueryRead_600),
		testing.Benchmark(BenchmarkMySQLDatabaseQueryRead_700),
		testing.Benchmark(BenchmarkMySQLDatabaseQueryRead_800),
		testing.Benchmark(BenchmarkMySQLDatabaseQueryRead_900),
		testing.Benchmark(BenchmarkMySQLDatabaseQueryRead_1000),
		testing.Benchmark(BenchmarkMySQLDatabaseQueryRead_2000),
		testing.Benchmark(BenchmarkMySQLDatabaseQueryRead_3000),
		testing.Benchmark(BenchmarkMySQLDatabaseQueryRead_4000),
		testing.Benchmark(BenchmarkMySQLDatabaseQueryRead_5000),
		testing.Benchmark(BenchmarkMySQLDatabaseQueryRead_8000),
		testing.Benchmark(BenchmarkMySQLDatabaseQueryRead_10000),
		testing.Benchmark(BenchmarkMySQLDatabaseQueryRead_15000),
		testing.Benchmark(BenchmarkMySQLDatabaseQueryRead_20000),
		testing.Benchmark(BenchmarkMySQLDatabaseQueryRead_25000),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_Caching_100),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_Caching_200),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_Caching_300),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_Caching_400),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_Caching_500),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_Caching_600),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_Caching_700),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_Caching_800),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_Caching_900),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_Caching_1000),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_Caching_2000),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_Caching_3000),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_Caching_4000),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_Caching_5000),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_Caching_8000),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_Caching_10000),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_Caching_15000),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_Caching_20000),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_Caching_25000),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_100),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_200),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_300),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_400),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_500),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_600),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_700),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_800),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_900),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_1000),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_2000),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_3000),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_4000),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_5000),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_8000),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_10000),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_15000),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_20000),
		testing.Benchmark(BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching_25000),
	}
}
