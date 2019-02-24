package middleware

import (
	"database/sql"
	"errors"
	"fmt"
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

func TestMySqlPrivateDatabase_Query(t *testing.T) {
	// TODO: set this up so that it initially creates the necessary tables. Unfortunately this is always going to need
	// a database running unless that could be mocked - may look into this but not a major requirement
	// also write the initial data into the DB and check the output data
	funcMap := validFuncMap()
	colMap := map[string][]string{"TestGroup": []string{}}

	group := &PrivacyGroup{"TestGroup", map[string]bool{"jacob": true}}

	staticDataPolicy := NewStaticDataPolicy([]*PrivacyGroup{group},
		DataTransforms{group: &TableOperations{funcMap, colMap}})

	db := MySQLPrivateDatabase{
		DataPolicy: staticDataPolicy,
	}

	err := db.Connect("demouser", "demopassword", "store1", "127.0.0.1", 3306)
	require.NoError(t, err)
	_, err = db.Query("SELECT * from people", &RequestPolicy{"jacob", Local, true})
	require.NoError(t, err)
}

func TestMySqlPrivateDatabase_QueryRow(t *testing.T) {
	// TODO: set this up so that it initially creates the necessary tables. Unfortunately this is always going to need
	// a database running unless that could be mocked - may look into this but not a major requirement
	// also write the initial data into the DB and check the output data
	funcMap := validFuncMap()
	colMap := map[string][]string{"TestGroup": []string{}}

	group := &PrivacyGroup{"TestGroup", map[string]bool{"jacob": true}}

	staticDataPolicy := NewStaticDataPolicy([]*PrivacyGroup{group},
		DataTransforms{group: &TableOperations{funcMap, colMap}})

	db := MySQLPrivateDatabase{
		DataPolicy: staticDataPolicy,
	}

	err := db.Connect("demouser", "demopassword", "store1", "127.0.0.1", 3306)
	require.NoError(t, err)
	_, err = db.QueryRow("SELECT * from people", &RequestPolicy{"jacob", Local, true})
	require.NoError(t, err)
}

func TestMySqlPrivateDatabase_Exec(t *testing.T) {
	// TODO: set this up so that it initially creates the necessary tables. Unfortunately this is always going to need
	// a database running unless that could be mocked - may look into this but not a major requirement
	// also write the initial data into the DB and check the output data

	funcMap := validFuncMap()
	colMap := map[string][]string{"TestGroup": []string{}}

	group := &PrivacyGroup{"TestGroup", map[string]bool{"jacob": true}}

	staticDataPolicy := NewStaticDataPolicy([]*PrivacyGroup{group},
		DataTransforms{group: &TableOperations{funcMap, colMap}})

	db := MySQLPrivateDatabase{
		DataPolicy: staticDataPolicy,
	}

	err := db.Connect("demouser", "demopassword", "store1", "127.0.0.1", 3306)
	require.NoError(t, err)
	_, err = db.Query("SELECT * from people", &RequestPolicy{"jacob", Local, true})
	require.NoError(t, err)
}

// TODO: add a test for batching
// TODO: test read/write logic
// TODO: add a test for applying transforms properly

func validFuncMap() map[string]map[string]func(interface{}) (interface{}, error) {
	funcMap := make(map[string]map[string]func(interface{}) (interface{}, error))
	funcMap["TestGroup"] = make(map[string]func(interface{}) (interface{}, error))

	funcMap["TestGroup"]["dob"] = func(arg interface{}) (interface{}, error) {
		date, ok := arg.(*time.Time)

		if !ok {
			return nil, errors.New("argument could not be asserted as time.Time")
		}
		onlyYear := time.Date(date.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		return onlyYear, nil
	}
	funcMap["TestGroup"]["name"] = func(arg interface{}) (interface{}, error) {
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

	return funcMap
}

// BENCHMARKS

func benchmarkMySQLDatabaseQueryRead(b *testing.B, queryString string) {
	b.StopTimer()
	db, err := sql.Open("mysql",
		fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=UTC",
			"demouser", "demopassword", "127.0.0.1", 3306, "store1"))
	if err != nil {
		b.Error(err.Error())
	}
	db.SetMaxIdleConns(0)
	db.SetConnMaxLifetime(time.Second * 20)
	b.StartTimer()

	_, err = db.Query(queryString)
	if err != nil {
		b.Error(err.Error())
	}
}

func BenchmarkMySQLDatabase_Query_Read(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkMySQLDatabaseQueryRead(b, "SELECT * from people")
	}
}

func benchmarkMySQLPrivateDatabaseQueryReadNoCaching(b *testing.B, queryString string) {
	b.StopTimer()
	funcMap := validFuncMap()
	colMap := map[string][]string{"TestGroup": {}}

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

	_, err = db.Query(queryString, &RequestPolicy{"jacob", Local, true})
	if err != nil {
		b.Error(err.Error())
	}
}

func BenchmarkMySQLPrivateDatabase_Query_Read_No_Caching(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkMySQLPrivateDatabaseQueryReadNoCaching(b, "SELECT * from people")
	}
}

func benchmarkMySQLPrivateDatabaseQueryReadCaching(b *testing.B, queryString string) {
	b.StopTimer()
	funcMap := validFuncMap()
	colMap := map[string][]string{"TestGroup": {}}

	group := &PrivacyGroup{"TestGroup", map[string]bool{"jacob": true}}

	staticDataPolicy := NewStaticDataPolicy([]*PrivacyGroup{group},
		DataTransforms{group: &TableOperations{funcMap, colMap}})

	db := MySQLPrivateDatabase{
		DataPolicy:  staticDataPolicy,
		CacheTables: true,
	}
	err := db.Connect("demouser", "demopassword", "store1", "127.0.0.1", 3306)
	if err != nil {
		b.Error(err.Error())
	}

	// Make the query once so we know we have a cached version of the table
	_, err = db.Query(queryString, &RequestPolicy{"jacob", Local, true})
	if err != nil {
		b.Error(err.Error())
	}
	b.StartTimer()

	_, err = db.Query(queryString, &RequestPolicy{"jacob", Local, true})
	if err != nil {
		b.Error(err.Error())
	}
}

func BenchmarkMySQLPrivateDatabase_Query_Read_Caching(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkMySQLPrivateDatabaseQueryReadCaching(b, "SELECT * from people")
	}
}

func benchmarkMySQLDatabaseExecWrite(b *testing.B, execString string) {
	b.StopTimer()
	db, err := sql.Open("mysql",
		fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=UTC",
			"demouser", "demopassword", "127.0.0.1", 3306, "store1"))
	if err != nil {
		b.Error(err.Error())
	}
	db.SetMaxIdleConns(0)
	db.SetConnMaxLifetime(time.Second * 20)
	b.StartTimer()

	_, err = db.Exec(execString)
	if err != nil {
		b.Error(err.Error())
	}
}

func BenchmarkMySQLDatabase_Exec_Write(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkMySQLDatabaseExecWrite(b, `INSERT INTO people (name, dob) VALUES ('steve', '1996-02-07')`)
	}
}

func benchmarkMySQLPrivateDatabaseExecWrite(b *testing.B, execString string) {
	b.StopTimer()
	funcMap := validFuncMap()
	colMap := map[string][]string{"TestGroup": {}}

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
		&RequestPolicy{"jacob", Local, true})
	if err != nil {
		b.Error(err.Error())
	}
}

func BenchmarkMySQLPrivateDatabase_Exec_Write(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkMySQLPrivateDatabaseExecWrite(b, `INSERT INTO people (name, dob) VALUES ('steve', '1996-02-07')`)
	}
}
