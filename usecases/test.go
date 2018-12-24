package main

import (
	"errors"
	"github.com/JacobMoxham/partiiproject/middleware"
	"time"
)

func main() {
	db := middleware.MySqlPrivateDatabase{}
	db.Connect()
	funcMap := make(map[string]func(interface{}) (interface{}, error))
	funcMap["dob"] = func(arg interface{}) (interface{}, error) {
		date, ok := arg.(*time.Time)
		if !ok {
			return nil, errors.New("argument could not be asserted as time.Time")
		}
		onlyYear := time.Date(date.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		return onlyYear, nil
	}
	cols := []string{}
	db.TransformRows("people", "trans_table", cols, funcMap)
}
