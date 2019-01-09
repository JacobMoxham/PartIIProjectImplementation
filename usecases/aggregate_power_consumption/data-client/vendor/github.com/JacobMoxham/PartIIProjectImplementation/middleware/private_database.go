package middleware

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/xwb1989/sqlparser"
	"log"
	"regexp"
	"strings"
	"time"
)

const batchSize = 1000

// PrivateRelationalDatabase wraps an SQL database and edits queries so that they operate over tables adjusted to match privacy policies
type PrivateRelationalDatabase interface {
	Connect(string, string, string) error
	Close() error
	Query(string *RequestPolicy) (*sql.Rows, error)
	// TODO: consider how updates are handled? Perhaps edit update query and check for excluded rows? Perhaps just consider read only
}

type MySqlPrivateDatabase struct {
	StaticDataPolicy *StaticDataPolicy
	database         *sql.DB
}

func (mspd *MySqlPrivateDatabase) Connect(user string, password string, databaseName string, uri string, port int) error {
	// TODO consider getting the time.Time location from somewhere
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=UTC", user, password, uri, port, databaseName))
	if err != nil {
		return err
	}
	db.SetMaxIdleConns(0)
	mspd.database = db
	return nil
}

func (mspd *MySqlPrivateDatabase) Close() error {
	err := mspd.database.Close()
	return err
}

func (mspd *MySqlPrivateDatabase) Query(query string, context *RequestPolicy) (*sql.Rows, error) {
	// Parse query
	// Parse query
	stmt, err := sqlparser.Parse(query)
	if err != nil {
		return nil, err
	}

	// Get all tables in query
	var tableNames []string
	err = sqlparser.Walk(
		func(node sqlparser.SQLNode) (kcontinue bool, err error) {
			switch node := node.(type) {
			case sqlparser.TableName:
				tableName := node.Name.String()
				if tableName != "" {
					tableNames = append(tableNames, node.Name.String())
				}
			}
			return true, nil
		}, stmt)
	if err != nil {
		return nil, err
	}

	groupPrefix := fmt.Sprintf("transformed_%s_", context.RequesterID)
	for _, tableName := range tableNames {
		// Create a version of the table with the privacy policy applied
		transformedTableName := groupPrefix + tableName
		tableOperations, err := mspd.StaticDataPolicy.Resolve(context.RequesterID)
		if err != nil {
			return nil, err
		}
		err = mspd.transformRows(tableName, transformedTableName, tableOperations.TableTransforms, tableOperations.ExcludedCols)
		if err != nil {
			// TODO: remove log
			//  log.Fatal(err)
			return nil, err
		}

		// Replace table Name with transformed table Name in query
		regexString := fmt.Sprintf("\\b%s\\b", tableName)
		re := regexp.MustCompile(regexString)
		query = re.ReplaceAllString(query, transformedTableName)
	}

	// Execute query
	rows, err := mspd.database.Query(query)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func contains(stringList []string, element string) bool {
	for _, el := range stringList {
		if el == element {
			return true
		}
	}
	return false
}

func (mspd *MySqlPrivateDatabase) dropTableIfExists(table string) error {
	dropTableString := fmt.Sprintf("DROP TABLE IF EXISTS %s;", table)
	_, err := mspd.database.Exec(dropTableString)
	return err
}

func (mspd *MySqlPrivateDatabase) transformRows(tableName string, transformedTableName string,
	transforms map[string]func(interface{}) (interface{}, error), excludedColumns map[string][]string) error {
	// Get the column types
	// TODO: see if we can do this better - we almost certainly can
	columnNamesString := fmt.Sprintf("SELECT column_name, data_type FROM information_schema.columns WHERE table_name='%s';", tableName)
	columnNames, err := mspd.database.Query(columnNamesString)
	if err != nil {
		return err
	}
	var (
		colName string
		colType string
	)
	var colsToCopy []string

	for columnNames.Next() {
		err := columnNames.Scan(&colName, &colType)
		if err != nil {
			log.Fatal(err)
		}
		if !contains(excludedColumns[tableName], colName) {
			colsToCopy = append(colsToCopy, colName)
		}

	}
	columnNames.Close()

	// Create a string of the column names to be copied over and another string with the names and types
	// TODO: deal with primary keys and auto increments etc. - I think this is sorted by using LIKE
	columnString := strings.Join(colsToCopy, ", ")

	if columnString == "" {
		return errors.New("all columns are excluded, cannot create transformed table")
	} else {
		// Drop the table if it already exists
		err = mspd.dropTableIfExists(transformedTableName)
		if err != nil {
			return err
		}

		// Copy table
		createTableString := fmt.Sprintf("CREATE TABLE %s LIKE %s;", transformedTableName, tableName)
		_, err = mspd.database.Exec(createTableString)
		if err != nil {
			return err
		}

		// Get necessary columns from database
		selectedColumnsString := fmt.Sprintf("SELECT %s FROM %s", columnString, tableName)
		rows, err := mspd.database.Query(selectedColumnsString)
		if err != nil {
			return err
		}
		defer rows.Close()

		//// TODO: see if this works
		//colTypes, err := rows.ColumnTypes()
		//if err != nil {
		//	return err
		//}
		//for i, ct := range colTypes {
		//	scanArg := ct.ScanType()
		//	fmt.Println(scanArg.String())
		//	// Issue is it prefers its own types to build in go ones
		//	scanArgs[i] = reflect.New(scanArg).Interface()
		//}

		// Create arguments of the correct types to scan values into
		var vals []interface{}

		// Create scan variables of the correct underlying type
		colTypes, err := rows.ColumnTypes()
		if err != nil {
			return err
		}
		for _, ct := range colTypes {
			databaseTypeName := strings.ToLower(ct.DatabaseTypeName())
			// TODO add this in
			//nullable, ok := ct.Nullable()
			//args, ok := ct.DecimalSize()
			appendCorrectArgumentType(&vals, databaseTypeName)
		}

		scanArgs := make([]interface{}, len(vals))
		// Make the scan args point at the values
		for i := range vals {
			scanArgs[i] = &vals[i]
		}

		//vals := make([]interface{}, len(colsToCopy))

		var rowArguments []interface{}
		rowCount := 0
		rowsToWrite := ""
		for rows.Next() {
			// Read rows into variables
			err = rows.Scan(scanArgs...)
			if err != nil {
				return err
			}
			// Apply transforms to rows
			for i, val := range vals {
				currentCol := colsToCopy[i]
				transform, ok := transforms[currentCol]
				if ok {
					transformedVal, err := transform(val)
					if err != nil {
						return err
					}
					vals[i] = transformedVal
				}
			}

			// Add rows to string
			rowsToWrite += "("
			for i := 0; i < len(vals); i++ {
				rowsToWrite += "?, "
			}
			rowsToWrite = strings.TrimSuffix(rowsToWrite, ", ")
			rowsToWrite += "), "
			rowArguments = append(rowArguments, vals...)

			rowCount += 1
			if rowCount == batchSize {
				// Write to database and then continue
				// Remove the last comma and space
				rowsToWrite = strings.TrimSuffix(rowsToWrite, ", ")

				// Write rows to transformed table
				insertString := fmt.Sprintf("INSERT INTO %s VALUES %s", transformedTableName, rowsToWrite)
				_, err = mspd.database.Exec(insertString, rowArguments...)
				if err != nil {
					return err
				}

				// Reset accumulators
				rowCount = 0
				rowsToWrite = ""
				rowArguments = rowArguments[:0]
			}
		}
		if rowCount > 0 {
			// Remove the last comma and space
			rowsToWrite = strings.TrimSuffix(rowsToWrite, ", ")

			// Write rows to transformed table
			insertString := fmt.Sprintf("INSERT INTO %s VALUES %s", transformedTableName, rowsToWrite)
			_, err = mspd.database.Exec(insertString, rowArguments...)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//// appendSliceWithoutAliasing copies the values from the second slice into the first. This solves the problem of go
//// reusing the underlying array of the first slice if it can do in the built in append function
//func appendSliceWithoutAliasing(slice1 []interface{}, slice2 []interface{}) []interface{} {
//	// Extend the first slice so we can copy in the values from the second slice
//	oldLen := len(slice1)
//	slice1 = append(slice1, make([]interface{}, len(slice2))...)
//	for i, val := range slice2 {
//		slice1[oldLen+i] = val
//	}
//	return slice1
//}

func appendCorrectArgumentType(vals *[]interface{}, colType string) {
	// TODO: this very much needs testing and debugging as well as use of the flags and nullable fields
	switch colType {
	case "tinyint":
		// TODO consider flags for unsigned and nullable
		*vals = append(*vals, *new(int8))
	case "smallint":
		fallthrough
	case "year":
		// TODO consider flags for unsigned and nullable
		*vals = append(*vals, *new(int16))
	case "mediumint":
		fallthrough
	case "int":
		fallthrough
	case "integer":
		// TODO consider flags for unsigned and nullable
		*vals = append(*vals, *new(int32))
	case "bigint":
		// TODO consider flags for unsigned and nullable
		// TODO SERIAL is an alias for BIGINT UNSIGNED NOT NULL AUTO_INCREMENT UNIQUE
		*vals = append(*vals, *new(int64))
	case "float":
		*vals = append(*vals, *new(float32))
	case "double":
		*vals = append(*vals, *new(float64))
	case "varchar":
		fallthrough
	case "text":
		fallthrough
	case "longtext":
		fallthrough
	case "char":
		fallthrough
	case "enum":
		fallthrough
	case "set":
		fallthrough
	case "blob":
		fallthrough
	case "tinyblob":
		fallthrough
	case "mediumblob":
		fallthrough
	case "longblob":
		fallthrough
	case "varbinary":
		*vals = append(*vals, *new(string))
	case "date":
		fallthrough
	case "datetime":
		fallthrough
	case "timestamp":
		fallthrough
	case "time":
		*vals = append(*vals, *new(time.Time))
	default:
		log.Println(fmt.Sprintf("Could not find corresponding GO type for %s, using interface{}", colType))
		*vals = append(*vals, *new(interface{}))
	}
}
