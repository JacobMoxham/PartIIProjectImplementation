package middleware

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"github.com/xwb1989/sqlparser"
	"log"
	"math/rand"
	"regexp"
	"strings"
	"sync"
	"time"
)

const batchSize = 1000

// PrivateRelationalDatabase wraps an SQL database and edits queries so that they operate over tables adjusted to match privacy policies
type PrivateRelationalDatabase interface {
	Connect(string, string, string) error
	Close() error
	Query(string *RequestPolicy) (*sql.Rows, error)
	// TODO: implement writes
}

type mutexMap struct {
	sync.Mutex
	rawMutexMap map[string]*sync.Mutex
}

func (m *mutexMap) GetMutex(key string) *sync.Mutex {
	m.Lock()
	defer m.Unlock()

	// Lazily initialise map
	if m.rawMutexMap == nil {
		m.rawMutexMap = make(map[string]*sync.Mutex)
	}

	// Try to get a mutex for the table and if not, create one
	mutex, ok := m.rawMutexMap[key]
	if !ok {
		mutex = &sync.Mutex{}
		m.rawMutexMap[key] = mutex
	}
	return mutex
}

type MySqlPrivateDatabase struct {
	StaticDataPolicy *StaticDataPolicy
	CacheTables      bool
	database         *sql.DB
	databaseName     string
	tableMutexes     mutexMap
}

func (mspd *MySqlPrivateDatabase) Connect(user string, password string, databaseName string, uri string, port int) error {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=UTC", user, password, uri, port, databaseName))
	if err != nil {
		return err
	}
	db.SetMaxIdleConns(0)
	db.SetConnMaxLifetime(time.Second * 20)
	mspd.database = db
	mspd.databaseName = databaseName
	return nil
}

func (mspd *MySqlPrivateDatabase) Close() error {
	err := mspd.database.Close()
	return err
}

func (mspd *MySqlPrivateDatabase) Query(query string, context *RequestPolicy) (*sql.Rows, error) {
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
		err = mspd.transformTable(tableName, transformedTableName, tableOperations.TableTransforms[tableName], tableOperations.ExcludedCols[tableName])
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

func (mspd *MySqlPrivateDatabase) transformTable(tableName string, transformedTableName string,
	transforms map[string]func(interface{}) (interface{}, error), excludedColumns []string) error {

	if mspd.CacheTables {
		// Check if we have a valid cached table
		valid, err := mspd.checkCache(tableName, transformedTableName)
		if err != nil {
			return err
		}
		if valid {
			return nil
		}
	} else {
		// Add a random ID to the table name to avoid clashes with concurrent requests
		transformedTableName += fmt.Sprintf("%d", rand.Intn(99999))
	}

	err := mspd.doTransform(tableName, transformedTableName, transforms, excludedColumns)
	if err != nil {
		return err
	}

	if !mspd.CacheTables {
		// Drop the table we created
		err = mspd.dropTableIfExists(transformedTableName)
		if err != nil {
			return err
		}
	}

	return nil
}

func (mspd *MySqlPrivateDatabase) doTransform(tableName string, transformedTableName string,
	transforms map[string]func(interface{}) (interface{}, error), excludedColumns []string) error {
	// Get the column types
	colsToCopy, err := mspd.getColsToCopy(tableName, excludedColumns)
	if err != nil {
		return err
	}

	// Create a string of the column names to be copied over and another string with the names and types
	columnString := strings.Join(colsToCopy, ", ")

	if columnString == "" {
		return errors.New("all columns are excluded, cannot create transformed table")
	}
	// Drop the table if it already exists - it should not exist at this point
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

	// Create arguments of the correct types to scan values into
	vals, scanArgs, err := getVariablesForRows(rows)

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
		err := applyTransformsToRows(&vals, colsToCopy, transforms)
		if err != nil {
			return err
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
			// Remove the last comma and space
			rowsToWrite = strings.TrimSuffix(rowsToWrite, ", ")

			// Write to database and then continue
			err := mspd.writeToTable(transformedTableName, rowsToWrite, rowArguments)
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
		err := mspd.writeToTable(transformedTableName, rowsToWrite, rowArguments)
		if err != nil {
			return err
		}
	}
	return nil
}

func applyTransformsToRows(vals *[]interface{}, colsToCopy []string,
	transforms map[string]func(interface{}) (interface{}, error)) error {
	for i, val := range *vals {
		currentCol := colsToCopy[i]
		transform, ok := transforms[currentCol]
		if ok {
			transformedVal, err := transform(val)
			if err != nil {
				return err
			}
			(*vals)[i] = transformedVal
		}
	}
	return nil
}

func (mspd *MySqlPrivateDatabase) writeToTable(tableName string, rowsToWrite string, rowArguments []interface{}) error {
	// Write rows to transformed table
	insertString := fmt.Sprintf("INSERT INTO %s VALUES %s", tableName, rowsToWrite)
	_, err := mspd.database.Exec(insertString, rowArguments...)
	if err != nil {
		return err
	}
	return nil
}

func (mspd *MySqlPrivateDatabase) dropTableIfExists(table string) error {
	dropTableString := fmt.Sprintf("DROP TABLE IF EXISTS %s;", table)
	_, err := mspd.database.Exec(dropTableString)
	return err
}

func (mspd *MySqlPrivateDatabase) isTransformedTableValid(tableName string, transformedTableName string) (bool, error) {
	// Check when the table was last updated
	// TODO: check when the security policy was last changed
	timeOfLastUpdateRow := mspd.database.QueryRow(
		`SELECT create_time, update_time FROM information_schema.tables WHERE table_schema = ? AND table_name = ?`,
		mspd.databaseName, tableName)

	var (
		timeOfLastUpdate     time.Time
		nullCreateTime       mysql.NullTime
		nullTimeOfLastUpdate mysql.NullTime
	)
	switch err := timeOfLastUpdateRow.Scan(&nullTimeOfLastUpdate, &nullCreateTime); err {
	case nil:
		if nullTimeOfLastUpdate.Valid {
			timeOfLastUpdate = nullTimeOfLastUpdate.Time
		} else if nullCreateTime.Valid {
			// use the creation time
			timeOfLastUpdate = nullCreateTime.Time
		} else {
			// Can't find a create or updated time
			return false, fmt.Errorf("no creation or last updated time could be found for %s", tableName)
		}
	case sql.ErrNoRows:
		// Table doesn't exist
		return false, fmt.Errorf("table %s doesn't exist", tableName)
	default:
		return false, fmt.Errorf("error reading table %s", tableName)
	}

	// Check when the transform was created
	timeOfTransformCreationRow := mspd.database.QueryRow(
		`SELECT create_time FROM information_schema.tables WHERE table_schema = ? AND table_name = ?`,
		mspd.databaseName, transformedTableName)

	var (
		nullTimeOfTransformCreation mysql.NullTime
		timeOfTransformCreation     time.Time
	)

	switch err := timeOfTransformCreationRow.Scan(&nullTimeOfTransformCreation); err {
	case nil:
		if nullTimeOfTransformCreation.Valid {
			timeOfTransformCreation = nullTimeOfTransformCreation.Time
		} else {
			// Continue and recreate table but log it
			log.Printf("malformed table entry %s, recreating", transformedTableName)
		}
	case sql.ErrNoRows:
		// No transform has been built so we can continue
		log.Printf("No transform found, creating table %s", transformedTableName)
	default:
		// Create the table anyway but log the errors
		log.Printf(err.Error())
	}

	// Work out whether the transform is valid
	return timeOfTransformCreation.After(timeOfLastUpdate), nil
}

func (mspd *MySqlPrivateDatabase) getColsToCopy(tableName string, excludedColumns []string) ([]string, error) {
	// Get the columns in the table
	columnNamesString := fmt.Sprintf("SELECT column_name, data_type FROM information_schema.columns WHERE table_name='%s';", tableName)
	columnNames, err := mspd.database.Query(columnNamesString)
	if err != nil {
		return nil, err
	}
	defer columnNames.Close()

	// Remove any columns which should be excluded
	var (
		colName string
		colType string
	)
	var colsToCopy []string

	for columnNames.Next() {
		err := columnNames.Scan(&colName, &colType)
		if err != nil {
			return nil, err
		}
		if !contains(excludedColumns, colName) {
			colsToCopy = append(colsToCopy, colName)
		}

	}
	return colsToCopy, nil
}

func (mspd *MySqlPrivateDatabase) checkCache(tableName string, transformedTableName string) (bool, error) {
	// Take out table lock
	mutex := mspd.tableMutexes.GetMutex(tableName)
	mutex.Lock()
	defer mutex.Unlock()

	// Check if transform is valid
	transformValid, err := mspd.isTransformedTableValid(tableName, transformedTableName)
	if err != nil {
		return false, err
	}

	// Do not recreate the transform if it is valid
	return transformValid, nil
}

func getVariablesForRows(rows *sql.Rows) ([]interface{}, []interface{}, error) {
	var vals []interface{}

	// Create scan variables of the correct underlying type
	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, nil, err
	}
	for _, ct := range colTypes {
		appendCorrectArgumentType(&vals, ct)
	}

	scanArgs := make([]interface{}, len(vals))
	// Make the scan args point at the values
	for i := range vals {
		scanArgs[i] = &vals[i]
	}

	return vals, scanArgs, nil
}

func appendCorrectArgumentType(vals *[]interface{}, columnType *sql.ColumnType) {
	databaseTypeName := strings.ToLower(columnType.DatabaseTypeName())

	// TODO: use these
	//nullable, hasNullable := columnType.Nullable()
	//precision, scale, hasPrecisionSclae := columnType.DecimalSize()

	// TODO: this very much needs testing and debugging as well as use of the flags and nullable fields
	switch databaseTypeName {
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
		log.Println(fmt.Sprintf("Could not find corresponding GO type for %s, using interface{}", databaseTypeName))
		*vals = append(*vals, *new(interface{}))
	}
}
