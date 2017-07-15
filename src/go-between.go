// go-between - https://github.com/Danw33/go-between/
// Copyright (C) 2017 Daniel Wilson - @Danw33
// MIT License - see LICENSE.md for conditions of use
package gobetween

import (
  "time"
  "log"
  "fmt"
  "runtime"
  "flag"
  "os"
  "os/signal"
  "syscall"
  "net/url"
  "net/http"
  "database/sql"
  "encoding/json"
  "github.com/gorilla/mux"
_ "github.com/denisenkom/go-mssqldb"
)

const version = "0.0.1"
const appname = "Danw33's Go-Between API Server"

var (
  // Application
  debug bool
  startTime time.Time

  // HTTP API Server Configuration
  listenAddress string
  listenPort int

  // DB Configuration
  dbDriver string
  dbPassword string
  dbPort int
  dbHostname string
  dbInstance string
  dbSchema string
  dbUser string

  // DB Connection Pointer
  db *sql.DB;
)

// Blocking (sleep) function for main thread
var quit = make(chan struct{})

type Response struct {
    Status string
    Message string
    Time int64
    ResponseData
}

type ResponseData struct {
    Data map[string]string
}

// Go-Between
func main() {

  // Startup
  logStartupInfo()

  // Parse Flags
  parseFlags()

  if debug {
    log.Println("  -- -- -- ")
    log.Println("WARNING: Debug mode enabled, Sensitive data may be logged to tty/file!")
    log.Println("  -- -- -- ")
  }

  connectionString := configureSQLConnection()
  srv := configureApiServer()

  if debug {
    log.Printf(" - Server Pointer [%p]", &srv)
  }

  db = openSQLConnection(dbDriver, connectionString)

  if debug {
    log.Printf(" - DB Pointer [%p]", &db)
  }

  log.Println("Creating shutdown signal handler for interrupt SIGINT/SIGTERM")
  ct := make(chan os.Signal, 2)
  signal.Notify(ct, os.Interrupt, syscall.SIGTERM)
  go func() {
    <-ct
    log.Println("Trapped OS interrupt signal SIGINT/SIGTERM!")
    shutdown(db)
  }()

  // Ensure that the thread ending would close the DB connection cleanly
  defer shutdown(db)

  // Start the HTTP Server
  go startApiServer(srv)

  // Run a basic check of the backend DB
  go checkBackendSanity(db)

  log.Println("Startup completed.")

  // Block main thread until quit is called
  <-quit

}

// Clean shudown sequenece - Closes DB connections and exits
func shutdown(db *sql.DB) {

  log.Println("Shutting Down...")

  log.Println("Closing DB connection(s)")
  if debug {
    log.Printf(" - Using DB Pointer [%p]", &db)
  }
  db.Close()

  log.Println("Clean shutdown sequenece completed.")
  os.Exit(0)
}

// Parses configuration flags into pointers
func parseFlags() {

    log.Println("Parsing input flags...")

    // Application
    flag.BoolVar(&debug, "debug", false, "Enable debug logging")

    // API Server
    flag.StringVar(&listenAddress, "listenAddress", "127.0.0.1", "HTTP API Listen Address")
    flag.IntVar(&listenPort, "listenPort", 8080, "HTTP API Listen Port")

    // Backend
    flag.StringVar(&dbDriver, "dbDriver", "mssql", "Database Driver (mssql/sqlserver)")
    flag.StringVar(&dbHostname, "dbHostname", "", "Database Server Hostname or IP Address")
    flag.IntVar(&dbPort, "dbPort", 1433, "Database Server Port")
    flag.StringVar(&dbInstance, "dbInstance", "", "Database Instance (optional)")
    flag.StringVar(&dbSchema, "dbSchema", "", "Database Schema Name")
    flag.StringVar(&dbUser, "dbUser", "", "Database Username")
    flag.StringVar(&dbPassword, "dbPassword", "", " Database Password")

    flag.Parse()

    if debug {
      log.Printf("Flags:          debug [%p]: %t", &debug, debug)
      log.Printf("Flags:  listenAddress [%p]: %s", &listenAddress, listenAddress)
      log.Printf("Flags:     listenPort [%p]: %d", &listenPort, listenPort)
      log.Printf("Flags:       dbDriver [%p]: %s", &dbDriver, dbDriver)
      log.Printf("Flags:     dbHostname [%p]: %s", &dbHostname, dbHostname)
      log.Printf("Flags:         dbPort [%p]: %d", &dbPort, dbPort)
      log.Printf("Flags:     dbInstance [%p]: %s", &dbInstance, dbInstance)
      log.Printf("Flags:       dbSchema [%p]: %s", &dbSchema, dbSchema)
      log.Printf("Flags:         dbUser [%p]: %s", &dbUser, dbUser)
      log.Printf("Flags:     dbPassword [%p]: %s", &dbPassword, dbPassword)
    }

    log.Println("Input flags processed.")

}

// Write information about the runtime and application to the log
// Also responsable for setting the startTime var
func logStartupInfo() {

  startTime = time.Now()

  log.Printf("Starting %s, version %s.", appname, version)

  switch os := runtime.GOOS; os {
  case "darwin":
    log.Printf("Runtime OS is MacOS (%s).", os)
  case "linux":
    log.Printf("Runtime OS is Linux (%s).", os)
  case "windows":
    log.Printf("Runtime OS is Windows (%s).", os)
  default:
    log.Printf("Runtime OS is Unknown (%s).", os)
  }

  arch := runtime.GOARCH
  log.Printf("Runtime Architecture is %s.", arch)

  log.Printf("Started at %s (%d)", startTime.Format(time.RFC3339), startTime.UnixNano())

}

// Calculate and return application uptime (as nanoseconds)
func getAppUptime() int64 {

  start :=  startTime.UnixNano()
  now := time.Now().UnixNano()

  uptime := now - start;

  return uptime
}

// Buld and return the configuration for the HTTP server
func configureApiServer() *http.Server {

  log.Println("API: Preparing HTTP API Server configuration...")

  // Router Setup
  router := mux.NewRouter()

  // Routes
  router.HandleFunc("/", outputTestWebResponse)
  router.HandleFunc("/status", outputStatusWebResponse)
  router.HandleFunc("/tables", outputTablesWebResponse)

  // HTTP Server Configuration
  srv := &http.Server{
    Handler:      router,
    Addr:         fmt.Sprintf("%s:%d", listenAddress, listenPort),
    WriteTimeout: 15 * time.Second,
    ReadTimeout:  15 * time.Second,
    MaxHeaderBytes: 1 << 20,
  }

  log.Println("API: HTTP Server configuration built.")

  return srv

}

// Build and return the configuration (connection string) for the SQL DB connection
func configureSQLConnection() string {

  log.Println("DB: Preparing SQL Server connection string...")

  query := url.Values{}
  query.Add("database", fmt.Sprintf("%s", dbSchema))

  dsn := &url.URL{
      Scheme:   "sqlserver",
      User:     url.UserPassword(dbUser, dbPassword),
      Host:     fmt.Sprintf("%s:%d", dbHostname, dbPort),
      Path:     dbInstance, // if connecting to an instance instead of a port
      RawQuery: query.Encode(),
  }

  connectionString := dsn.String()

  log.Println("DB: SQL Server connection string built.")

  return connectionString

}

// Open the SQL database connection using the given driver and connection string
// Returns a pointer to the database connection upon success
func openSQLConnection(driver string, connectionString string) *sql.DB {

  log.Printf("DB: Opening new SQL Server connection using driver '%s'...", driver)

  db, err := sql.Open(driver, connectionString)

  if err != nil {
    log.Println("DB: Failed to connect to SQL Server: ", err.Error())
    return db
  }

  err = db.Ping()
  if err != nil {
    log.Println("DB: Failed to connect to and ping SQL Server: ", err.Error())
    return db
  }

  log.Println("DB: SQL Server connection ready.")

  if debug {
    log.Printf("DB: Database connection: [%p] %s", &db, db)
  }

  return db

}

// Start the HTTP server using the given configured instance pointer
func startApiServer(srv *http.Server) {

  log.Printf("API: Starting built-in HTTP server on %s", srv.Addr)

  if debug {
    log.Printf(" - Using Server Pointer [%p]", &srv)
  }

  log.Fatal(srv.ListenAndServe())

}

// Perform basic sanity checks against the SQL database
func checkBackendSanity(db *sql.DB) bool {

  log.Println("DB: Running backend sanity check...")

  if debug {
    log.Printf(" - DB Pointer [%p]", &db)
  }

  if sqlCountTables(db) > 0 {
    log.Println("DB: Sanity check passed.")
    return true
  }

  log.Fatal("DB: Sanity check failed! No tables found in DB.")

  return false

}

// Count the number of tables in the configured database schema
// Returns the table count which is also written to the log
func sqlCountTables(db *sql.DB) int {

  stmt, err := db.Prepare("SELECT Distinct COUNT(TABLE_NAME) FROM information_schema.TABLES")

  if err != nil {
    log.Fatal("DB: Failed to construct prepared statement:", err.Error())
  }
  defer stmt.Close()

  // Query = Row Set, QueryRow = Single Row
  row := stmt.QueryRow()

  var tcount int
  err = row.Scan(&tcount)
  if err != nil {
    log.Fatal("Scan failed:", err.Error())
  }
  log.Printf("DB: Discovered %d tables from information_schema\n", tcount)

  return tcount

}

// Transmits (writes) the responseData to the responseWriter as JSON
func transmit(responseWriter http.ResponseWriter, responseData Response) {

  b, err := json.Marshal(responseData)

  if err != nil {
    log.Fatal("API: Error encoding JSON Response:", err.Error())
  }

  responseWriter.Write(b)
}

// Logs a HTTP request (with debug information, where enabled)
func logRequest(req *http.Request) {

  log.Printf("API: Handling %s %s request from %s to %s", req.Proto, req.Method, req.RemoteAddr, req.URL)

  if debug {
    log.Printf(" - Request Pointer [%p]", &req)
    log.Printf(" - Request Data: %s", req)
  }

}

// HTTP API: Handler for /
// Outputs a JSON response with a basic "It Works!" message
func outputTestWebResponse(response http.ResponseWriter, req *http.Request) {

  logRequest(req)

  responseData := Response{"success", "It Works!", time.Now().UnixNano(), ResponseData{nil}}

  transmit(response, responseData)

}

// HTTP API: Handler for /status
// Outputs a JSON-formatted status message
func outputStatusWebResponse(response http.ResponseWriter, req *http.Request) {

  logRequest(req)

  var healthData = make(map[string]string)

  healthData["debug"] = fmt.Sprintf("%t", debug);
  healthData["started"] = fmt.Sprintf("%d", startTime.UnixNano());
  healthData["uptime"] = fmt.Sprintf("%d", getAppUptime());
  healthData["version"] = fmt.Sprintf("%s", version);

  responseData := Response{"success", "System functional", time.Now().UnixNano(), ResponseData{healthData}}

  transmit(response, responseData)

}

// HTTP API: Handler for /tables
// Outputs a JSON response with a list of database tables
func outputTablesWebResponse(response http.ResponseWriter, req *http.Request) {

  logRequest(req)

  stmt, err := db.Prepare("SELECT Distinct TABLE_NAME FROM information_schema.TABLES")

  if err != nil {
    log.Fatal("DB: Failed to construct prepared statement: ", err.Error())
  }
  defer stmt.Close()

  // Query = Row Set, QueryRow = Single Row
  rows, err := stmt.Query()
  if err != nil {
    log.Fatal("DB: Failed to execute prepared query against database: ", err.Error())
  }

  var resultField string
  var tableData = make(map[string]string)
  var i int = 0

  for rows.Next() {

    err = rows.Scan(&resultField)
    if err != nil {
      log.Fatal("DB: Row scan failed: ", err.Error())
    }

    if debug {
      log.Printf("DB: Found Table: %s", resultField)
    }

    tableData[fmt.Sprintf("%d", i)] = fmt.Sprintf("%s", resultField)
    i++

  }

  responseData := Response{"success", "", time.Now().UnixNano(), ResponseData{tableData}}

  transmit(response, responseData)

}
