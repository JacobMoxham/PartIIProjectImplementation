package middleware

import "testing"
import (
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	"time"
)

func TestMySqlPrivateDatabase_ConnectAndClose(t *testing.T) {
	db := MySqlPrivateDatabase{}
	err := db.Connect("demouser", "demopassword", "store1")
	t.Name()
	require.NoError(t, err)

	err = db.Close()
	require.NoError(t, err)
}

func TestMySqlPrivateDatabase_Query(t *testing.T) {
	// TODO: set this up so that it initially creates the necessary tables. Unfortunately this is always going to need
	// a database running unless that could be mocked - may look into this but not a major requirement

	// TODO: also write the initial data into the DB and check the output data

	funcMap := make(map[string]func(interface{}) (interface{}, error))
	funcMap["dob"] = func(arg interface{}) (interface{}, error) {
		date, ok := arg.(*time.Time)

		if !ok {
			return nil, errors.New("argument could not be asserted as time.Time")
		}
		onlyYear := time.Date(date.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		return onlyYear, nil
	}
	funcMap["name"] = func(arg interface{}) (interface{}, error) {
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
	colMap := map[string][]string{"TestGroup": []string{}}

	group := &PrivacyGroup{"TestGroup", map[string]bool{"jacob": true}}

	db := MySqlPrivateDatabase{
		StaticDataPolicy: &StaticDataPolicy{
			PrivacyGroups: []*PrivacyGroup{group},
			Transforms:    DataTransforms{group: &TableOperations{funcMap, colMap}},
		},
	}
	db.Connect("demouser", "demopassword", "store1")

	db.Query("SELECT * from people", &PamContext{"jacob"})
}

// TODO: add a test for batching
