package middleware

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)
import (
	"github.com/stretchr/testify/require"
)

// TESTS

func TestMySqlPrivateDatabase_ConnectAndClose(t *testing.T) {
	db := MySQLPrivateDatabase{}
	err := db.Connect("demouser", "demopassword", "store1", "localhost", 33060)
	t.Name()
	require.NoError(t, err)

	err = db.Close()
	require.NoError(t, err)
}

func TestMySqlPrivateDatabase_Query_NoCaching(t *testing.T) {
	db := validPrivateDBConnection(t, "store1")

	err := db.Connect("demouser", "demopassword", "store1", "127.0.0.1", 3306)
	require.NoError(t, err)
	_, err = db.Query("SELECT * from people", &RequestPolicy{"jacob", Local, true})
	require.NoError(t, err)
}

func TestMySqlPrivateDatabase_Query_Caching(t *testing.T) {
	db := validPrivateDBConnection(t, "store1")

	err := db.Connect("demouser", "demopassword", "store1", "127.0.0.1", 3306)
	require.NoError(t, err)
	_, err = db.Query("SELECT * from people", &RequestPolicy{"jacob", Local, true})
	require.NoError(t, err)
}

func TestMySqlPrivateDatabase_QueryRow(t *testing.T) {
	db := validPrivateDBConnection(t, "store1")
	_, err := db.QueryRow("SELECT * from people", &RequestPolicy{"jacob", Local, true})
	require.NoError(t, err)
}

func TestMySqlPrivateDatabase_Exec_Read(t *testing.T) {
	db := validPrivateDBConnection(t, "store1")

	_, err := db.Exec("SELECT * from people", &RequestPolicy{"jacob", Local, true})
	require.NoError(t, err)
}

func TestMySqlPrivateDatabase_Exec_Write(t *testing.T) {
	db := validPrivateDBConnection(t, "store1")

	requestPolicy := &RequestPolicy{"jacob", Local, true}

	// Write a record
	result, err := db.Exec(`INSERT INTO people (name, dob) VALUES ('steve', '1996-02-07')`,
		requestPolicy)
	require.NoError(t, err)

	writeID, err := result.LastInsertId()
	require.NoError(t, err)

	// Read that record back
	row, err := db.QueryRow(`SELECT name, dob FROM people WHERE id=?`, requestPolicy, writeID)
	require.NoError(t, err)

	var (
		name string
		dob  time.Time
	)

	err = row.Scan(&name, &dob)
	require.NoError(t, err)

	dobString, err := time.Parse("2006-01-02", "1996-02-07")
	require.NoError(t, err)

	require.Equal(t, name, `steve`)
	require.Equal(t, dob, dobString)
}

func TestMySqlPrivateDatabase_Exec_Write_To_Excluded_Col(t *testing.T) {
	// Exclude dob column from access
	funcMap := validFuncMap()
	colMap := map[string][]string{"people": {"dob"}}

	group := &PrivacyGroup{"TestGroup", map[string]bool{"jacob": true}}

	staticDataPolicy := NewStaticDataPolicy([]*PrivacyGroup{group},
		DataTransforms{group: &TableOperations{funcMap, colMap}})

	db := MySQLPrivateDatabase{
		DataPolicy: staticDataPolicy,
	}

	err := db.Connect("demouser", "demopassword", "store1", "127.0.0.1", 3306)
	require.NoError(t, err)

	// Attempt to update the dob column (we assume the existence of (id, jacob, 1997-11-01 in the database)
	_, err = db.Exec(`UPDATE people SET dob = '1996-02-07' WHERE name = jacob`,
		&RequestPolicy{"jacob", Local, true})
	require.EqualError(t, err, "query failed")
}

func TestMySqlPrivateDatabase_Exec_Write_Not_To_Excluded_Col(t *testing.T) {
	// Get a database connection
	funcMap := validFuncMap()
	colMap := map[string][]string{"people": {"dob"}}

	group := &PrivacyGroup{"TestGroup", map[string]bool{"jacob": true}}

	staticDataPolicy := NewStaticDataPolicy([]*PrivacyGroup{group},
		DataTransforms{group: &TableOperations{funcMap, colMap}})

	db := MySQLPrivateDatabase{
		DataPolicy: staticDataPolicy,
	}

	err := db.Connect("demouser", "demopassword", "store1", "127.0.0.1", 3306)
	require.NoError(t, err)

	requestPolicy := &RequestPolicy{"jacob", Local, true}

	// Update the name column
	_, err = db.Exec(`UPDATE people SET name = 'William' WHERE name = jacob`, requestPolicy)

	// It was considered whether to not you should be able to update columns from a table where some are excluded from
	// you as long as you don't rely on any of these columns. However it was decided that as the parser does not always
	// make it clear which table a column is from this would not be possible.
	require.EqualError(t, err, "query failed")
}

// TODO: test delete
// TODO: add a test for applying transforms properly
// TODO: add test for excludedRows (write done, need to do read)

func validFuncMap() map[string]TableTransform {
	funcMap := make(map[string]TableTransform)
	funcMap["TestGroup"] = make(TableTransform)

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

func validPrivateDBConnection(t *testing.T, databaseName string) MySQLPrivateDatabase {
	// TODO: set this up so that it initially creates the necessary tables. Unfortunately this is always going to need
	// a database running unless that could be mocked - may look into this but not a major requirement
	// also write the initial data into the DB and check the output data

	funcMap := validFuncMap()
	colMap := map[string][]string{}

	group := &PrivacyGroup{"TestGroup", map[string]bool{"jacob": true}}

	staticDataPolicy := NewStaticDataPolicy([]*PrivacyGroup{group},
		DataTransforms{group: &TableOperations{funcMap, colMap}})

	db := MySQLPrivateDatabase{
		DataPolicy: staticDataPolicy,
	}

	err := db.Connect("demouser", "demopassword", databaseName, "127.0.0.1", 3306)
	require.NoError(t, err)

	return db
}

// BENCHMARKS

var globalResult *sql.Rows

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
	globalResult = r

	b.StopTimer()
	err = db.Close()
	if err != nil {
		b.Error(err.Error())
	}
	b.StartTimer()
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
	colMap := map[string][]string{}

	group := &PrivacyGroup{"TestGroup", map[string]bool{"jacob": true}}

	staticDataPolicy := NewStaticDataPolicy([]*PrivacyGroup{group},
		DataTransforms{group: &TableOperations{funcMap, colMap}})

	db := MySQLPrivateDatabase{
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
	globalResult = r

	b.StopTimer()
	err = db.Close()
	if err != nil {
		b.Error(err.Error())
	}
	b.StartTimer()
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

func benchmarkMySQLPrivateDatabaseQuery(b *testing.B, db MySQLPrivateDatabase, queryString string) *sql.Rows {
	r, err := db.Query(queryString, &RequestPolicy{"jacob", Local, true})
	if err != nil {
		b.Error(err.Error())
	}
	return r
}

func benchmarkMySQLPrivateDatabaseQueryReadCaching(b *testing.B, queryString string) {
	b.StopTimer()
	funcMap := validFuncMap()
	colMap := map[string][]string{}

	group := &PrivacyGroup{"TestGroup", map[string]bool{"jacob": true}}

	staticDataPolicy := NewStaticDataPolicy([]*PrivacyGroup{group},
		DataTransforms{group: &TableOperations{funcMap, colMap}})

	db := MySQLPrivateDatabase{
		DataPolicy:  staticDataPolicy,
		CacheTables: true,
	}
	err := db.Connect("demouser", "demopassword", "power_consumption", "127.0.0.1", 3306)
	if err != nil {
		b.Error(err.Error())
	}
	// Make the query once so we know we have a cached version of the table
	_, err = db.Query(queryString, &RequestPolicy{"jacob", Local, true})
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
	globalResult = r

	b.StopTimer()
	err = db.Close()
	if err != nil {
		b.Error(err.Error())
	}
	b.StartTimer()
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

	group := &PrivacyGroup{"TestGroup", map[string]bool{"jacob": true}}

	staticDataPolicy := NewStaticDataPolicy([]*PrivacyGroup{group},
		DataTransforms{group: &TableOperations{funcMap, colMap}})

	db := MySQLPrivateDatabase{
		DataPolicy:  staticDataPolicy,
		CacheTables: false,
	}
	err := db.Connect("demouser", "demopassword", "store1", "127.0.0.1", 3306)
	if err != nil {
		b.Error(err.Error())
	}
	b.StartTimer()

	_, err = db.Exec(execString,
		&RequestPolicy{"jacob", Local, true},
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
