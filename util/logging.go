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

	// Name of the environment variable that limits the payload size
	envLogMaxPayloadSize = "GOVCD_MAX_PAYLOAD_SIZE"

	// Name of the environment variable that limits the result size
	envLogMaxMaxResultSize = "GOVCD_MAX_RESULT_SIZE"
)

var (
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
	logOpen bool = false

	// The log file handle
	apiLog *os.File

	// Text lines used for logging of http requests and responses
	lineLength int    = 80
	dashLine   string = strings.Repeat("-", lineLength)
	hashLine   string = strings.Repeat("#", lineLength)
)

// initializes logging with known parameters
func SetLog() {
	if !EnableLogging {
		log.SetOutput(ioutil.Discard)
		return
	}
	var err error

	// If no file name was set, logging goes to the screen
	if ApiLogFileName == "" {
		if LogOnScreen == "stderr" || LogOnScreen == "err" {
			log.SetOutput(os.Stderr)
		} else {
			log.SetOutput(os.Stdout)
		}
	} else {
		apiLog, err = os.OpenFile(ApiLogFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			fmt.Printf("error opening log file %s : %v", ApiLogFileName, err)
			os.Exit(1)
		}
		log.SetOutput(apiLog)
		logOpen = true
	}
}

// Closes a log
func CloseLog() {
	if logOpen {
		apiLog.Close()
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

func logSanitizedHeader(input_header http.Header) {
	for key, value := range input_header {
		if (key == "Config-Secret" || key == "authorization" || key == "Authorization" || key == "X-Vcloud-Authorization") &&
			!LogPasswords {
			value = []string{"********"}
		}
		log.Printf("\t%s: %s\n", key, value)
	}
}

// Logs the essentials of a HTTP request
func ProcessRequestOutput(caller, operation, url, payload string, req *http.Request) {
	if !LogHttpRequest {
		return
	}
	log.Printf("%s\n", dashLine)
	log.Printf("Request caller: %s\n", caller)
	log.Printf("%s %s\n", operation, url)
	log.Printf("%s\n", dashLine)
	data_size := len(payload)
	if isBinary(payload, req) {
		payload = "[binary data]"
	}
	if data_size > 0 {
		log.Printf("Request data: [%d] %s\n", data_size, hidePasswords(payload, false))
	}
	log.Printf("Req header:\n")
	logSanitizedHeader(req.Header)
}

// Logs the essentials of a HTTP response
func ProcessResponseOutput(caller string, resp *http.Response, result string) {
	if !LogHttpResponse {
		return
	}
	log.Printf("%s\n", hashLine)
	log.Printf("Response caller %s\n", caller)
	log.Printf("Response status %s\n", resp.Status)
	log.Printf("%s\n", hashLine)
	log.Printf("Response header:\n")
	logSanitizedHeader(resp.Header)
	data_size := len(result)
	log.Printf("Response text: [%d] %s\n", data_size, result)
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
