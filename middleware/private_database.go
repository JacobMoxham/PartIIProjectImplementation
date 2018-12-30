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
	Connect() error
	Close() error
	Query(string *PrivacyAwareMiddlewareContext) (*sql.Rows, error)
	// TODO: consider how updates are handled? Perhaps edit update query and check for excluded rows? Perhaps just consider read only
}

type MySqlPrivateDatabase struct {
	StaticDataPolicy *StaticDataPolicy
	database         *sql.DB
}

func (mspd *MySqlPrivateDatabase) Connect() error {
	// TODO consider getting the time.Time location from somewhere
	db, err := sql.Open("mysql", "demouser:demopassword@/store1?parseTime=true&loc=UTC")
	if err != nil {
		return err
	}
	mspd.database = db
	return nil
}

func (mspd *MySqlPrivateDatabase) Close() error {
	err := mspd.database.Close()
	return err
}

func (mspd *MySqlPrivateDatabase) Query(query string, context *PrivacyAwareMiddlewareContext) (*sql.Rows, error) {
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

	// Create arguments of the correct types to scan values into
	var scanArgs []interface{}

	for columnNames.Next() {
		err := columnNames.Scan(&colName, &colType)
		if err != nil {
			log.Fatal(err)
		}
		if !contains(excludedColumns[tableName], colName) {
			colsToCopy = append(colsToCopy, colName)
			// TODO: remove if the below works
			//appendCorrectArgumentType(&scanArgs, colType)
		}

	}
	columnNames.Close()

	// Create a string of the column names to be copied over and another string with the names and types
	// TODO: deal with primary keys and auto increments etc. - I think this is sorted by using LIKE
	columnString := strings.Join(colsToCopy, ", ")

	// Drop the table if it already exists
	err = mspd.dropTableIfExists(transformedTableName)
	if err != nil {
		return err
	}

	if columnString == "" {
		return errors.New("all columns are exluded, cannot create transformed table")
	} else {
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

		colTypes, err := rows.ColumnTypes()
		if err != nil {
			return err
		}
		for _, ct := range colTypes {
			databaseTypeName := ct.DatabaseTypeName()
			// TODO add this in
			//nullable, ok := ct.Nullable()
			//args, ok := ct.DecimalSize()
			// TODO test when I can connect to mysql to see the results
			appendCorrectArgumentType(&scanArgs, databaseTypeName)
		}

		vals := make([]interface{}, len(colsToCopy))

		var rowArguments []interface{}
		rowCount := 0
		rowsToWrite := ""
		for rows.Next() {
			// Read rows into variables
			err = rows.Scan(scanArgs...)
			// Apply transforms to rows
			for i, val := range scanArgs {
				currentCol := colsToCopy[i]
				transform, ok := transforms[currentCol]
				if ok {
					transformedVal, err := transform(val)
					if err != nil {
						return err
					}
					vals[i] = transformedVal
				} else {
					vals[i] = val
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
			}
		}
		// Remove the last comma and space
		rowsToWrite = strings.TrimSuffix(rowsToWrite, ", ")

		// Write rows to transformed table
		insertString := fmt.Sprintf("INSERT INTO %s VALUES %s", transformedTableName, rowsToWrite)
		_, err = mspd.database.Exec(insertString, rowArguments...)
		if err != nil {
			return err
		}
	}
	return nil
}

func appendCorrectArgumentType(scanArgs *[]interface{}, colType string) {
	// TODO: this very much needs testing and debugging as well as use of the flags and nullable fields
	switch colType {
	case "tinyint":
		// TODO consider flags for unsigned and nullable
		*scanArgs = append(*scanArgs, new(int8))
	case "smallint":
		fallthrough
	case "year":
		// TODO consider flags for unsigned and nullable
		*scanArgs = append(*scanArgs, new(int16))
	case "mediumint":
		fallthrough
	case "int":
		fallthrough
	case "integer":
		// TODO consider flags for unsigned and nullable
		*scanArgs = append(*scanArgs, new(int32))
	case "bigint":
		// TODO consider flags for unsigned and nullable
		// TODO SERIAL is an alias for BIGINT UNSIGNED NOT NULL AUTO_INCREMENT UNIQUE
		*scanArgs = append(*scanArgs, new(int64))
	case "float":
		*scanArgs = append(*scanArgs, new(float32))
	case "double":
		*scanArgs = append(*scanArgs, new(float64))
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
		*scanArgs = append(*scanArgs, new(string))
	case "date":
		fallthrough
	case "datetime":
		fallthrough
	case "timestamp":
		fallthrough
	case "time":
		*scanArgs = append(*scanArgs, new(time.Time))
	}
}
