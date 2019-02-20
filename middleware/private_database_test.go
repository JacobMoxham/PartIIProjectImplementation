package middleware

import (
	"errors"
	"fmt"
	"testing"
	"time"
)
import (
	"github.com/stretchr/testify/require"
)

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

	db := MySQLPrivateDatabase{
		StaticDataPolicy: &StaticDataPolicy{
			PrivacyGroups: []*PrivacyGroup{group},
			Transforms:    DataTransforms{group: &TableOperations{funcMap, colMap}},
		},
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

	db := MySQLPrivateDatabase{
		StaticDataPolicy: &StaticDataPolicy{
			PrivacyGroups: []*PrivacyGroup{group},
			Transforms:    DataTransforms{group: &TableOperations{funcMap, colMap}},
		},
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

	db := MySQLPrivateDatabase{
		StaticDataPolicy: &StaticDataPolicy{
			PrivacyGroups: []*PrivacyGroup{group},
			Transforms:    DataTransforms{group: &TableOperations{funcMap, colMap}},
		},
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
