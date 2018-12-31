/*
 * Copyright 2018 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package util

import (
	"os"
	"path/filepath"
	"regexp"
	"testing"
)

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func testLog(logn int, t *testing.T, filename string, want_enabled bool, success_msg, failure_msg string) {
	Logger.Printf("test %d\n", logn)
	if want_enabled {
		if fileExists(filename) {
			t.Logf("ok - [%d] %s", logn, success_msg)
		} else {
			t.Logf("not ok - [%d] %s", logn, failure_msg)
			t.Fail()
		}
	} else {
		if !fileExists(filename) {
			t.Logf("ok - [%d] %s", logn, success_msg)
		} else {
			t.Logf("not ok - [%d] %s", logn, failure_msg)
			t.Fail()
		}
	}
}

func TestEnableLogging(t *testing.T) {
	ApiLogFileName = "temporary-for-test.log"
	custom_log_file := "temporary-custom-for-test.log"
	if fileExists(ApiLogFileName) {
		os.Remove(ApiLogFileName)
	}
	if fileExists(custom_log_file) {
		os.Remove(custom_log_file)
	}

	EnableLogging = true
	SetLog()
	testLog(1, t, ApiLogFileName, true, "log enabled", "log was not enabled")
	os.Remove(ApiLogFileName)

	EnableLogging = false
	SetLog()
	testLog(2, t, ApiLogFileName, false, "log was disabled", "log was not disabled")

	EnableLogging = false
	os.Setenv(envUseLog, "1")
	InitLogging()
	testLog(3, t, ApiLogFileName, true, "log enabled via env variable", "log was not enabled via env variable")
	os.Remove(ApiLogFileName)

	EnableLogging = false
	os.Setenv(envUseLog, "")
	InitLogging()
	testLog(4, t, ApiLogFileName, false, "log was disabled via env variable", "log was not disabled via env variable")
	customLogger := newLogger(custom_log_file)
	SetCustomLogger(customLogger)
	testLog(5, t, custom_log_file, true, "log was enabled via custom logger", "log was not enabled via custom logger")
	os.Remove(custom_log_file)
}

func TestCaller(t *testing.T) {
	type callData struct {
		fun      func() string
		label    string
		expected string
	}
	var data = []callData{
		{
			label:    "current function name",
			fun:      CurrentFuncName,
			expected: `^util.TestCaller$`,
		},
		{
			label:    "function caller",
			fun:      CallFuncName,
			expected: `^testing.tRunner$`,
		},
		{
			label:    "function stack",
			fun:      FuncNameCallStack,
			expected: "testing.tRunner",
		},
	}

	for _, d := range data {
		value := filepath.Base(d.fun())
		reFunc := regexp.MustCompile(`\b` + d.expected + `\b`)
		if reFunc.MatchString(value) {
			t.Logf("ok - %s as expected: '%s' matches '%s' \n", d.label, value, d.expected)
		} else {
			t.Logf("not ok - %s doesn't match. Expected: '%s' - Found: '%s'\n", d.label, d.expected, value)
			t.Fail()
		}
	}
}
