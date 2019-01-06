package usecases

import (
	"errors"
	"fmt"
	"github.com/JacobMoxham/PartIIProjectImplementation/middleware"
	"time"
)

func test_main() {
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

	group := &middleware.PrivacyGroup{"TestGroup", map[string]bool{"jacob": true}}

	db := middleware.MySqlPrivateDatabase{
		StaticDataPolicy: &middleware.StaticDataPolicy{
			PrivacyGroups: []*middleware.PrivacyGroup{group},
			Transforms:    middleware.DataTransforms{group: &middleware.TableOperations{funcMap, colMap}},
		},
	}
	db.Connect("demouser", "demopassword", "store1")

	db.Query("SELECT * from people", &middleware.RequestPolicy{"jacob", middleware.Local, true})
}
