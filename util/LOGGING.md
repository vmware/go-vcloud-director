# LOGGING


## Defaults for logging

Use of the standard Go `log` package is cwdeprecated and should be avoided. 
The recommended way of logging is through the logger `util.GovcdLogger`, which supports [all the functions normally available to `log`](https://golang.org/pkg/log/#Logger).


By default, **logging is disabled**. Any `GovcdLogger.Printf` statement will simply be discarded.

To enable logging, you should use

```go
util.EnableLogging = true
util.InitLogging()
```

When enabled, the default output for logging is a file named `go-vcloud-director.log`.
The file name can be changed using

```go
util.ApiLogFileName = "my_file_name.log"
```


If you want logging on the screen, use

```go
util.GovcdLogger.SetOutput(os.Stdout)
```

or

```
util.GovcdLogger.SetOutput(os.Stderr)
```

## Automatic logging of HTTP requests and responses.

The HTTP requests and responses are automatically logged.
Since all the HTTP operations go through `NewRequest` and `decodeBody`, the logging captures the input and output of the request with calls to `util.ProcessRequestOutput` and `util.ProcessResponseOutput`.

These two functions will show the request or response, and the function from which they were called, giving devs an useful tracking tool.

The output of these functions can be quite large. If you want to mute the HTTP processing, you can use:

```go
util.LogHttpRequest = false
util.LogHttpResponse = false
```

During the request and response processing, any password or authentication token found through pattern matching will be automatically hidden. To show passwords in your logs, use

```go
util.LogPasswords = true
```

## Emergency logging of HTTP operations

When logging is **disabled**, there is an emergency stack that keeps in memory the latest N HTTP operations (4 by default). When a failure occurs, this stack is written to the log, which gets updated on-the-fly.

If you want also this emergency system to be disabled, you can add this code to your client:

```go
util. EnableHttpStack = false
util.InitLogging()
```

## Custom logger

If the configuration options are not enough for your needs, you can supply your own logger.

```go
util.SetCustomLogger(mylogger)
```

## Environment variables

The logging behavior can be changed without coding. There are a few environment variables that are checked when the library is used:

```EnableLogging``` corresponds to

Variable                    | Corresponding environment var 
--------------------------- | :-------------------------------
`EnableLogging`             | `GOVCD_LOG`
`ApiLogFileName`            | `GOVCD_LOG_FILE`
`LogPasswords`              | `GOVCD_LOG_PASSWORDS`
`LogOnScreen`               | `GOVCD_LOG_ON_SCREEN`
`LogHttpRequest`            | `GOVCD_LOG_SKIP_HTTP_REQ`
`LogHttpResponse`           | `GOVCD_LOG_SKIP_HTTP_RESP`

