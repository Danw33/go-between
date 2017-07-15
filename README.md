# Go-Between

Just a noob fumbling around in go.
Learning the ropes of the go language and creating a small web API for communicating with MS-SQL (a "go-between" as such).

## Usage

I hadn't intended this for public use, rather just storing it here for personal reference.
If you find it useful, or can learn from it in any way (just as I am), drop me a tweet - @Danw33 to let me know!

### Run from source

To run go-between directly from source, use the following command:

```
go run ./src/go-between.go [flags]
```

(where `[flags]` is a sequence of arguments to configure the application)

### Build and run

To compile go-between from source to an executable binary, then run it, use the following commands:

```
go build ./src/go-between.go
chmod +X ./go-between
./go-between [flags]
```

(where `[flags]` is a sequence of arguments to configure the application)

By default, the go compiler will include a symbol table. To build without this the following can be used:

```
go build -ldflags "-s -w" ./src/go-between.go
```

While this produces a smaller output binary, crashes would yield no real usable stack trace information.

### Flags
The go-between application can take a number of runtime arguments (flags) to configure various aspects of the program.

```
Usage of ./go-between:
  -dbDriver string
    	Database Driver (mssql/sqlserver) (default "mssql")
  -dbHostname string
    	Database Server Hostname or IP Address
  -dbInstance string
    	Database Instance (optional)
  -dbPassword string
    	 Database Password
  -dbPort int
    	Database Server Port (default 1433)
  -dbSchema string
    	Database Schema Name
  -dbUser string
    	Database Username
  -debug
    	Enable debug logging
  -listenAddress string
    	HTTP API Listen Address (default "127.0.0.1")
  -listenPort int
    	HTTP API Listen Port (default 8080)
```

Debugging can be enabled by passing the flag `-debug` when launching go-between; this enables additional console log output which may be useful when troubleshooting.

## Useful Links

 - [The Go Programming Language](https://golang.org/)
 - [gorilla/mux](https://github.com/gorilla/mux) - The [Gorilla Toolkit Package mux](http://www.gorillatoolkit.org/pkg/mux) HTTP request router and dispatcher.
 - [denisenkom/go-mssqldb](https://github.com/denisenkom/go-mssqldb) - Go MSSQL driver for Go's database/sql package

## License

You've read this far so something must be of use to you! Go ahead - take what you need (**MIT License**).

See LICENSE.md for full terms of use.
