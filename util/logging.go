/*
 * Copyright 2018 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

// Package util provides ancillary functionality to go-vcloud-director library
// logging.go regulates logging for the whole library.
// See LOGGING.md for detailed usage
package util

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strings"
)

// Used by GovcdHttpStack to store operations
type OperationRec struct {
	Caller    string
	Operation string
	Url       string
	Data      string
	OpType    string
	Req       *http.Request
	Resp      *http.Response
}

const (
	// Name of the environment variable that enables logging
	envUseLog = "GOVCD_LOG"

	// Name of the environment variable with the log file name
	envLogFileName = "GOVCD_LOG_FILE"

	// Name of the environment variable with the screen output
	envLogOnScreen = "GOVCD_LOG_ON_SCREEN"

	// Name of the environment variable that enables logging of passwords
	envLogPasswords = "GOVCD_LOG_PASSWORDS"

	// Name of the environment variable that enables logging of HTTP requests
	envLogSkipHttpReq = "GOVCD_LOG_SKIP_HTTP_REQ"

	// Name of the environment variable that enables logging of HTTP responses
	envLogSkipHttpResp = "GOVCD_LOG_SKIP_HTTP_RESP"
)

var (
	// All go-vcloud director logging goes through this logger
	GovcdLogger *log.Logger

	// HTTP operation stack
	// When logging is **disabled**, the stack keeps in memory
	// several (GovcdHttpStackSize) HTTP operations, to be used
	// in case of failure.
	// When there is a failure, main HTTP processing functions in api.go
	// will call util.LogHttpOperations() to empty the stack into the
	// default log, which is enabled on-the-fly.
	GovcdHttpStack FixedSizeStack

	// HTTP operation stack size
	// 4 operations == two requests and two responses
	GovcdHttpStackSize int = 4

	// Enable storing HTTP operations
	EnableHttpStack bool = true

	// It's true if we're using an user provided logger
	customLogging bool = false

	// Name of the log file
	// activated by GOVCD_LOG_FILE
	ApiLogFileName string = "go-vcloud-director.log"

	// Globally enabling logs
	// activated by GOVCD_LOG
	EnableLogging bool = false

	// Enable logging of passwords
	// activated by GOVCD_LOG_PASSWORDS
	LogPasswords bool = false

	// Enable logging of Http requests
	// disabled by GOVCD_LOG_SKIP_HTTP_REQ
	LogHttpRequest bool = true

	// Enable logging of Http responses
	// disabled by GOVCD_LOG_SKIP_HTTP_RESP
	LogHttpResponse bool = true

	// Sends log to screen. If value is either "stderr" or "err"
	// logging will go to os.Stderr. For any other value it will
	// go to os.Stdout
	LogOnScreen string = ""

	// Flag indicating that a log file is open
	// logOpen bool = false

	// The log file handle
	apiLog *os.File

	// Text lines used for logging of http requests and responses
	lineLength int    = 80
	dashLine   string = strings.Repeat("-", lineLength)
	hashLine   string = strings.Repeat("#", lineLength)
)

func newLogger(logpath string) *log.Logger {
	// println("LogFile: " + logpath)
	file, err := os.Create(logpath)
	if err != nil {
		fmt.Printf("error opening log file %s : %v", logpath, err)
		os.Exit(1)
	}
	return log.New(file, "", log.Ldate|log.Ltime)
}

func SetCustomLogger(customLogger *log.Logger) {
	GovcdLogger = customLogger
	EnableLogging = true
	customLogging = true
}

// initializes logging with known parameters
func SetLog() {
	if customLogging {
		return
	}
	if !EnableLogging {
		GovcdLogger = log.New(ioutil.Discard, "", log.Ldate|log.Ltime)
		return
	}

	// If no file name was set, logging goes to the screen
	if ApiLogFileName == "" {
		if LogOnScreen == "stderr" || LogOnScreen == "err" {
			log.SetOutput(os.Stderr)
			GovcdLogger = log.New(os.Stderr, "", log.Ldate|log.Ltime)
		} else {
			GovcdLogger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
		}
	} else {
		GovcdLogger = newLogger(ApiLogFileName)
	}
}

// Hides passwords that may be used in a request
func hidePasswords(in string, onScreen bool) string {
	if !onScreen && LogPasswords {
		return in
	}
	re := regexp.MustCompile(`("[^\"]*[Pp]assword"\s*:\s*)"[^\"]+"`)
	return re.ReplaceAllString(in, `${1}"********"`)
}

// Determines whether a string is likely to contain binary data
func isBinary(data string, req *http.Request) bool {
	re_content_range := regexp.MustCompile(`(?i)content-range`)
	re_multipart := regexp.MustCompile(`(?i)multipart/form`)
	re_media_xml := regexp.MustCompile(`(?i)media+xml;`)
	for key, value := range req.Header {
		if re_content_range.MatchString(key) {
			return true
		}
		if re_multipart.MatchString(key) {
			return true
		}
		for _, v := range value {
			if re_media_xml.MatchString(v) {
				return true
			}
		}
	}
	return false
}

// Scans the header for known keys that contain authentication tokens
// and hide the contents
func logSanitizedHeader(input_header http.Header) {
	for key, value := range input_header {
		if (key == "Config-Secret" || key == "authorization" || key == "Authorization" || key == "X-Vcloud-Authorization") &&
			!LogPasswords {
			value = []string{"********"}
		}
		GovcdLogger.Printf("\t%s: %s\n", key, value)
	}
}

// Logs the essentials of a HTTP request
func ProcessRequestOutput(caller, operation, url, payload string, req *http.Request) {
	if EnableHttpStack {
		GovcdHttpStack.Push(OperationRec{OpType: "request", Caller: caller, Req: req, Data: payload, Url: url, Operation: operation})
	}
	if !LogHttpRequest {
		return
	}
	GovcdLogger.Printf("%s\n", dashLine)
	GovcdLogger.Printf("Request caller: %s\n", caller)
	GovcdLogger.Printf("%s %s\n", operation, url)
	GovcdLogger.Printf("%s\n", dashLine)
	data_size := len(payload)
	if isBinary(payload, req) {
		payload = "[binary data]"
	}
	if data_size > 0 {
		GovcdLogger.Printf("Request data: [%d] %s\n", data_size, hidePasswords(payload, false))
	}
	GovcdLogger.Printf("Req header:\n")
	logSanitizedHeader(req.Header)
}

// Logs the essentials of a HTTP response
func ProcessResponseOutput(caller string, resp *http.Response, result string) {
	if EnableHttpStack {
		GovcdHttpStack.Push(OperationRec{OpType: "response", Caller: caller, Resp: resp, Data: result})
	}
	if !LogHttpResponse {
		return
	}
	GovcdLogger.Printf("%s\n", hashLine)
	GovcdLogger.Printf("Response caller %s\n", caller)
	GovcdLogger.Printf("Response status %s\n", resp.Status)
	GovcdLogger.Printf("%s\n", hashLine)
	GovcdLogger.Printf("Response header:\n")
	logSanitizedHeader(resp.Header)
	data_size := len(result)
	GovcdLogger.Printf("Response text: [%d] %s\n", data_size, result)
}

func LogHttpOperations() {
	if !EnableHttpStack {
		return
	}
	if GovcdHttpStack.Len() == 0 {
		return
	}
	EnableLogging = true
	EnableHttpStack = false
	InitLogging()
	GovcdLogger.Printf("%s\n", hashLine)
	GovcdLogger.Printf("THERE ARE %d STORED HTTP OPERATIONS (Lower numbers are most recent) \n", GovcdHttpStack.Len())
	GovcdLogger.Printf("%s\n", hashLine)
	stackSize := GovcdHttpStack.Len()
	for N := 0; N < GovcdHttpStackSize; N++ {
		var opRec OperationRec
		latest := GovcdHttpStack.Pop()
		if latest == nil {
			return
		}
		opRec = latest.(OperationRec)
		GovcdLogger.Printf("%s\n", dashLine)
		GovcdLogger.Printf("STORED OPERATION (%s) # %d\n", opRec.OpType, stackSize-N)
		GovcdLogger.Printf("%s\n", dashLine)
		if opRec.OpType == "request" {
			ProcessRequestOutput(opRec.Caller, opRec.Operation, opRec.Url, opRec.Data, opRec.Req)
		} else {
			ProcessResponseOutput(opRec.Caller, opRec.Resp, opRec.Data)
		}
	}
}

// Initializes default logging values
func InitLogging() {
	if os.Getenv(envLogSkipHttpReq) != "" {
		LogHttpRequest = false
	}

	if os.Getenv(envLogSkipHttpResp) != "" {
		LogHttpResponse = false
	}
	if os.Getenv(envLogPasswords) != "" {
		EnableLogging = true
		LogPasswords = true
	}

	if os.Getenv(envLogFileName) != "" {
		EnableLogging = true
		ApiLogFileName = os.Getenv(envLogFileName)
	}

	LogOnScreen = os.Getenv(envLogOnScreen)
	if LogOnScreen != "" {
		ApiLogFileName = ""
		EnableLogging = true
	}

	if EnableLogging || os.Getenv(envUseLog) != "" {
		EnableLogging = true
	}
	SetLog()
	if !EnableLogging {
		GovcdHttpStack = NewFixedSizeStack(GovcdHttpStackSize)
		EnableHttpStack = true
	}
}

func init() {
	InitLogging()
}

// Returns the name of the function that called the
// current function.
// Used by functions that call processResponseOutput and
// processRequestOutput
func CallFuncName() string {
	fpcs := make([]uintptr, 1)
	n := runtime.Callers(3, fpcs)
	if n > 0 {
		fun := runtime.FuncForPC(fpcs[0] - 1)
		if fun != nil {
			return fun.Name()
		}
	}
	return ""
}
