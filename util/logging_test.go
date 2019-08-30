/*
 * Copyright 2019 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
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
	return !os.IsNotExist(err)
}

func removeFile(t *testing.T, filename string) error {
	err := os.Remove(filename)
	if err != nil {
		t.Logf("Can't remove file %s", filename)
		t.Fail()
		return err
	}
	return nil
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
	customLogFile := "temporary-custom-for-test.log"
	if fileExists(ApiLogFileName) {
		err := removeFile(t, ApiLogFileName)
		if err != nil {
			return
		}
	}
	if fileExists(customLogFile) {
		err := removeFile(t, customLogFile)
		if err != nil {
			return
		}
	}

	EnableLogging = true
	SetLog()
	testLog(1, t, ApiLogFileName, true, "log enabled", "log was not enabled")
	err := removeFile(t, ApiLogFileName)
	if err != nil {
		return
	}

	EnableLogging = false
	SetLog()
	testLog(2, t, ApiLogFileName, false, "log was disabled", "log was not disabled")

	EnableLogging = false
	_ = os.Setenv(envUseLog, "1")
	InitLogging()
	testLog(3, t, ApiLogFileName, true, "log enabled via env variable", "log was not enabled via env variable")
	err = removeFile(t, ApiLogFileName)
	if err != nil {
		return
	}

	EnableLogging = false
	_ = os.Setenv(envUseLog, "")
	InitLogging()
	testLog(4, t, ApiLogFileName, false, "log was disabled via env variable", "log was not disabled via env variable")
	customLogger := newLogger(customLogFile)
	SetCustomLogger(customLogger)
	testLog(5, t, customLogFile, true, "log was enabled via custom logger", "log was not enabled via custom logger")
	err = removeFile(t, customLogFile)
	if err != nil {
		return
	}
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

func init() {
	// Before running log tests, let's make sure all the log related
	// environment variables are unset
	_ = os.Setenv(envUseLog, "")
	_ = os.Setenv(envLogFileName, "")
	_ = os.Setenv(envLogOnScreen, "")
	_ = os.Setenv(envLogPasswords, "")
	_ = os.Setenv(envLogSkipHttpReq, "")
	_ = os.Setenv(envLogSkipHttpResp, "")
	_ = os.Setenv(envLogSkipTagList, "")
	_ = os.Setenv(envLogFileName, "")
}
