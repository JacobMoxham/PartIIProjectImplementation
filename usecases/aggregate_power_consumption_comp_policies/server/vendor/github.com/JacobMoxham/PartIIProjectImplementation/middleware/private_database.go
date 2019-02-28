package middleware

import (
	"context"
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

// PrivateRelationalDatabase wraps an SQL database and edits queries so that they operate
// over tables adjusted to match privacy policies
type PrivateRelationalDatabase interface {
	Connect(user, password, databaseName, uri string, port int) error
	Close() error
	Query(query string, requestPolicy *RequestPolicy, args ...interface{}) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, requestPolicy *RequestPolicy, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, requestPolicy *RequestPolicy, args ...interface{}) (*sql.Row, error)
	QueryRowContext(ctx context.Context, query string, requestPolicy *RequestPolicy, args ...interface{}) (*sql.Row, error)
	Exec(ctx context.Context, query string, requestPolicy *RequestPolicy, args ...interface{}) (sql.Result, error)
	ExecContext(ctx context.Context, query string, requestPolicy *RequestPolicy, args ...interface{}) (sql.Result, error)
	Stats() sql.DBStats
	SetConnMaxLifetime(d time.Duration)
	SetMaxOpenConns(n int)
	SetMaxIdleConns(n int)
	Ping() error
	PingContext(ctx context.Context) error
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

// MySQLPrivateDatabase is a wrapper around a MySQL database which implements the PrivateRelationalDatabase interface,
// it supports DataPolicies which specify transforms for columns and excluded columns on a per PrivacyGroup basis
type MySQLPrivateDatabase struct {
	DataPolicy   DataPolicy
	CacheTables  bool
	database     *sql.DB
	databaseName string
	tableMutexes mutexMap
}

// Connect opens the connection to the MySQL database
func (mspd *MySQLPrivateDatabase) Connect(user, password, databaseName, uri string, port int) error {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=UTC", user, password, uri, port, databaseName))
	if err != nil {
		return err
	}
	// TODO: what are the consequences of changing these?
	db.SetMaxIdleConns(100)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(time.Second * 20)
	mspd.database = db
	mspd.databaseName = databaseName
	return nil
}

// Close closes the connection to the MySQL database
func (mspd *MySQLPrivateDatabase) Close() error {
	err := mspd.database.Close()
	return err
}

// Query takes a query string and a RequestPolicy and resolves the DataPolicy from the MySQLPrivateDatabase with the
// request policy to give a result to the query on transformed versions of the actual database tables
func (mspd *MySQLPrivateDatabase) Query(query string, requestPolicy *RequestPolicy, args ...interface{}) (*sql.Rows, error) {
	return mspd.QueryContext(context.Background(), query, requestPolicy, args...)
}

// QueryContext takes a query string and a RequestPolicy and resolves the DataPolicy from the MySQLPrivateDatabase with the
// request policy to give a result to the query on transformed versions of the actual database tables
func (mspd *MySQLPrivateDatabase) QueryContext(ctx context.Context, query string, requestPolicy *RequestPolicy, args ...interface{}) (*sql.Rows, error) {
	// Transform tables
	transformedQuery, err := mspd.transformTables(query, requestPolicy)
	if err != nil {
		return nil, err
	}

	// Execute query
	rows, err := mspd.database.QueryContext(ctx, transformedQuery, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// QueryRow takes a query string and a RequestPolicy and resolves the DataPolicy from the MySQLPrivateDatabase with the
//// request policy to give a result to the query on transformed versions of the actual database tables
func (mspd *MySQLPrivateDatabase) QueryRow(query string, requestPolicy *RequestPolicy, args ...interface{}) (*sql.Row, error) {
	return mspd.QueryRowContext(context.Background(), query, requestPolicy, args...)
}

// QueryRowContext takes a query string and a RequestPolicy and resolves the DataPolicy from the MySQLPrivateDatabase with the
// request policy to give a result to the query on transformed versions of the actual database tables
func (mspd *MySQLPrivateDatabase) QueryRowContext(ctx context.Context, query string, requestPolicy *RequestPolicy, args ...interface{}) (*sql.Row, error) {
	// Transform tables
	transformedQuery, err := mspd.transformTables(query, requestPolicy)
	if err != nil {
		return nil, err
	}

	// Execute query
	row := mspd.database.QueryRowContext(ctx, transformedQuery, args...)

	return row, nil
}

// Exec takes a query string and a RequestPolicy and resolves the DataPolicy from the MySQLPrivateDatabase with the
// request policy to give a result to the query on transformed versions of the actual database tables
func (mspd *MySQLPrivateDatabase) Exec(query string, requestPolicy *RequestPolicy, args ...interface{}) (sql.Result, error) {
	return mspd.ExecContext(context.Background(), query, requestPolicy, args...)
}

// ExecContext takes a query string and a RequestPolicy and resolves the DataPolicy from the MySQLPrivateDatabase with the
// request policy to give a result to the query on transformed versions of the actual database tables
func (mspd *MySQLPrivateDatabase) ExecContext(ctx context.Context, query string, requestPolicy *RequestPolicy, args ...interface{}) (sql.Result, error) {
	// Transform tables
	transformedQuery, err := mspd.transformTables(query, requestPolicy)
	if err != nil {
		return nil, err
	}

	// Execute query
	result, err := mspd.database.ExecContext(ctx, transformedQuery, args...)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Stats returns database statistics
func (mspd *MySQLPrivateDatabase) Stats() sql.DBStats {
	return mspd.database.Stats()
}

// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
//
// Expired connections may be closed lazily before reuse.
//
// If d <= 0, connections are reused forever.
func (mspd *MySQLPrivateDatabase) SetConnMaxLifetime(d time.Duration) {
	mspd.database.SetConnMaxLifetime(d)
}

// SetMaxOpenConns sets the maximum number of open connections to the database.
//
// If MaxIdleConns is greater than 0 and the new MaxOpenConns is less than
// MaxIdleConns, then MaxIdleConns will be reduced to match the new
// MaxOpenConns limit.
//
// If n <= 0, then there is no limit on the number of open connections.
// The default is 0 (unlimited).
func (mspd *MySQLPrivateDatabase) SetMaxOpenConns(n int) {
	mspd.database.SetMaxOpenConns(n)
}

// SetMaxIdleConns sets the maximum number of connections in the idle
// connection pool.
//
// If MaxOpenConns is greater than 0 but less than the new MaxIdleConns,
// then the new MaxIdleConns will be reduced to match the MaxOpenConns limit.
//
// If n <= 0, no idle connections are retained.
//
// The default max idle connections is currently 2. This may change in
// a future release.
func (mspd *MySQLPrivateDatabase) SetMaxIdleConns(n int) {
	mspd.database.SetMaxIdleConns(n)
}

// Ping verifies a connection to the database is still alive,
// establishing a connection if necessary.
func (mspd *MySQLPrivateDatabase) Ping() error {
	return mspd.database.Ping()
}

// PingContext verifies a connection to the database is still alive,
// establishing a connection if necessary.
func (mspd *MySQLPrivateDatabase) PingContext(ctx context.Context) error {
	return mspd.database.PingContext(ctx)
}

// TODO: consider wrapping Conns as well as DBs, also need to work out if Ping breaks our wrapper
// TODO: consider supporting Prepare and PrepareContext and BeginTx, Begin etc.

func (mspd *MySQLPrivateDatabase) transformTables(query string, requestPolicy *RequestPolicy) (string, error) {
	// Parse query
	stmt, err := sqlparser.Parse(query)
	if err != nil {
		return "", err
	}

	// Get all statements in the query
	var statements []sqlparser.Statement
	err = sqlparser.Walk(
		func(node sqlparser.SQLNode) (kcontinue bool, err error) {
			switch node := node.(type) {
			case sqlparser.Statement:
				statements = append(statements, node)
			}
			return true, nil
		}, stmt)
	if err != nil {
		return "", err
	}

	// Transform tables if the query only reads,
	// don't transform them but check for excluded column access if it only writes,
	// error if the query both reads and writes (the user needs to separate these queries)
	queryReads := false
	queryWrites := false
	for _, s := range statements {
		// TODO: add a test that this covers all of the statements we need it to
		switch s.(type) {
		case *sqlparser.Select:
			queryReads = true
		case *sqlparser.Update:
			queryWrites = true
		case *sqlparser.Insert:
			queryWrites = true
		case *sqlparser.Delete:
			queryWrites = true
		}
	}

	if queryReads && queryWrites {
		return "", errors.New("cannot support SQL which both reads and writes to the database")
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
		return "", err
	}

	groupPrefix := fmt.Sprintf("transformed_%s_", requestPolicy.RequesterID)
	for _, tableName := range tableNames {
		if queryReads {
			// Create a version of the table with the privacy policy applied
			transformedTableName := groupPrefix + tableName
			tableOperations, err := mspd.DataPolicy.Resolve(requestPolicy.RequesterID)
			if err != nil {
				return "", err
			}
			err = mspd.transformTable(tableName, transformedTableName, tableOperations.TableTransforms[tableName], tableOperations.ExcludedCols[tableName])
			if err != nil {
				return "", err
			}

			// Replace table Name with transformed table Name in query
			regexString := fmt.Sprintf("\\b%s\\b", tableName)
			re := regexp.MustCompile(regexString)
			query = re.ReplaceAllString(query, transformedTableName)
		} else if queryWrites {
			tableOperations, err := mspd.DataPolicy.Resolve(requestPolicy.RequesterID)
			if err != nil {
				return "", err
			}

			err = mspd.checkForExcludedColumns(tableName, tableOperations.ExcludedCols[tableName])
			if err != nil {
				return "", err
			}
		} else {
			return "", errors.New("unsupported query")
		}
	}
	return query, nil
}

func (mspd *MySQLPrivateDatabase) checkForExcludedColumns(tableName string, excludedColumns []string) error {
	// Get the columns in the table
	columnNamesString := fmt.Sprintf("SELECT column_name, data_type FROM information_schema.columns WHERE table_name='%s';", tableName)
	columnNames, err := mspd.database.Query(columnNamesString)
	if err != nil {
		return err
	}
	defer columnNames.Close()

	// Check if any columns which should be excluded are accessed
	var (
		colName string
		colType string
	)

	for columnNames.Next() {
		err := columnNames.Scan(&colName, &colType)
		if err != nil {
			return err
		}
		if contains(excludedColumns, colName) {
			// return an error if we access an excluded column but do not reveal the reason so we don't reveal
			// information through error messages
			return errors.New("query failed")
		}
	}

	if columnNames.Err() != nil {
		return columnNames.Err()
	}

	return nil
}

func (mspd *MySQLPrivateDatabase) transformTable(tableName string, transformedTableName string,
	transforms TableTransform, excludedColumns []string) error {

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

func (mspd *MySQLPrivateDatabase) doTransform(tableName string, transformedTableName string,
	transforms TableTransform, excludedColumns []string) error {
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
		excludeRow, err := applyTransformsToRows(&vals, colsToCopy, transforms)
		if err != nil {
			return err
		}

		if !excludeRow {
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
	}

	if rows.Err() != nil {
		return rows.Err()
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
	transforms TableTransform) (bool, error) {
	for i, val := range *vals {
		currentCol := colsToCopy[i]
		transform, ok := transforms[currentCol]
		if ok {
			transformedVal, excludeRow, err := transform(val)
			if err != nil {
				return true, err
			}
			if excludeRow {
				return true, nil
			}
			(*vals)[i] = transformedVal
		}
	}
	return false, nil
}

func (mspd *MySQLPrivateDatabase) writeToTable(tableName string, rowsToWrite string, rowArguments []interface{}) error {
	// Write rows to transformed table
	insertString := fmt.Sprintf("INSERT INTO %s VALUES %s", tableName, rowsToWrite)
	_, err := mspd.database.Exec(insertString, rowArguments...)
	if err != nil {
		return err
	}
	return nil
}

func (mspd *MySQLPrivateDatabase) dropTableIfExists(table string) error {
	dropTableString := fmt.Sprintf("DROP TABLE IF EXISTS %s;", table)
	_, err := mspd.database.Exec(dropTableString)
	return err
}

func (mspd *MySQLPrivateDatabase) isTransformedTableValid(tableName string, transformedTableName string) (bool, error) {
	// Check when the data policy was last updated
	timeOfLastPolicyUpdate := mspd.DataPolicy.LastUpdated()

	// Check when the table was last updated
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
	return timeOfTransformCreation.After(timeOfLastUpdate) || timeOfTransformCreation.After(timeOfLastPolicyUpdate), nil
}

func (mspd *MySQLPrivateDatabase) getColsToCopy(tableName string, excludedColumns []string) ([]string, error) {
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

	if columnNames.Err() != nil {
		return nil, columnNames.Err()
	}

	return colsToCopy, nil
}

func (mspd *MySQLPrivateDatabase) checkCache(tableName string, transformedTableName string) (bool, error) {
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
