/*
 * Copyright (c) 2021, 2022 Oracle and/or its affiliates.
 * Licensed under the Universal Permissive License v 1.0 as shown at
 * https://oss.oracle.com/licenses/upl.
 */

package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ohler55/ojg/jp"
	"github.com/oracle/coherence-cli/pkg/constants"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"unicode"
)

var (
	// DebugEnabled defines if debugging is enabled
	DebugEnabled bool

	// Logger is the logger to use for writing logs
	Logger *zap.Logger
)

// GetError returns a formatted error and prints to log
func GetError(message string, err error) error {
	var (
		errorDetails = fmt.Sprintf("%v", err)
		caller       = "unknown"
	)
	_, sourceFile, lineNo, ok := runtime.Caller(1)
	if ok {
		caller = fmt.Sprintf("%s#%d", filepath.Base(sourceFile), lineNo)
	}
	fields := []zapcore.Field{
		zap.String("location", caller),
		zap.String("message", message),
		zap.String("cause", errorDetails),
	}
	Logger.Error("Error", fields...)
	return fmt.Errorf("%s: %s", message, errorDetails)
}

// IsDistributedCache returns true if the service type is distributed
func IsDistributedCache(serviceType string) bool {
	return serviceType == constants.DistributedService || serviceType == constants.FederatedService
}

// SliceContains returns true of the slice contains the value
func SliceContains(theSlice []string, value string) bool {
	return GetSliceIndex(theSlice, value) != -1
}

// GetUniqueValues returns the slice of unique values
func GetUniqueValues(input []string) []string {
	result := make([]string, 0)
	for _, value := range input {
		if !SliceContains(result, value) {
			result = append(result, value)
		}
	}
	return result
}

// GetSliceIndex returns the index of the matching slice value
func GetSliceIndex(theSlice []string, value string) int {
	if len(theSlice) != 0 {
		for i, v := range theSlice {
			if v == value {
				return i
			}
		}
	}
	return -1
}

// ProcessJSONPath parses json path expression on Json and returns the json
func ProcessJSONPath(jsonData interface{}, jsonPathQuery string) ([]byte, error) {
	x, err := jp.ParseString(jsonPathQuery)
	if err != nil {
		return constants.EmptyByte, err
	}

	data, err := json.Marshal(x.Get(jsonData))
	return data, err
}

// GetJSONPathResults returns jsonapth results
func GetJSONPathResults(jsonData []byte, jsonPath string) (string, error) {
	var result interface{}
	err := json.Unmarshal(jsonData, &result)
	if err != nil {
		return "", GetError("GetJSONPathResults", err)
	}
	actualJSONPath := strings.ReplaceAll(jsonPath, constants.JSONPATH, "")

	results, err := ProcessJSONPath(result, actualJSONPath)
	if err != nil {
		return "", GetError("ProcessJSONPath", err)
	}
	return string(results), nil
}

// EnsureDirectory ensures a directory exists and if not then will create it
func EnsureDirectory(directory string) error {
	if _, err := os.Stat(directory); err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(directory, 0700)
			if err != nil {
				return GetError("unable to create directory "+directory, err)
			}
		}
	}
	return nil
}

// DirectoryExists returns a bool indicating if a directory exists
func DirectoryExists(directory string) bool {
	file, err := os.Stat(directory)
	if err != nil {
		fmt.Println(err)
		return !os.IsNotExist(err)
	}
	return file.IsDir()
}

// IsValidInt returns true or false indicating if a string int is a valid integer
func IsValidInt(value string) bool {
	_, err := strconv.Atoi(value)
	return err == nil
}

// SanitizeSnapshotName sanitizes a snapshot name by replacing any unwanted characters with '-'
func SanitizeSnapshotName(snapshotName string) string {
	var (
		sb = strings.Builder{}
	)

	for _, c := range []byte(snapshotName) {
		r := rune(c)
		if unicode.IsNumber(r) || unicode.IsLetter(r) || c == '-' || c == '_' {
			sb.WriteString(string(c))
		} else {
			sb.WriteString("-")
		}
	}
	return sb.String()
}

// GetErrors return an error containing either the single error or an
// error indicating there are multiple errors in the log
func GetErrors(errorList []error) error {
	if len(errorList) == 1 {
		return errorList[0]
	}
	for _, value := range errorList {
		_ = GetError("error", value)
	}
	return errors.New("multiple errors retrieving data, please see log file")
}

// CombineByteArraysForJSON combines byte arrays for json output
func CombineByteArraysForJSON(elements [][]byte, elementName []string) ([]byte, error) {
	var (
		result       = make([]byte, 0)
		length       = len(elements)
		comma        = []byte(",")
		openBracket  = []byte("{")
		closeBracket = []byte("}")
	)

	if length != len(elementName) {
		return constants.EmptyByte,
			fmt.Errorf("element names (%v) must be same length (%d) as elements", elementName, length)
	}

	result = append(result, openBracket...)
	for i, element := range elements {
		result = append(result, []byte(fmt.Sprintf("\"%s\":", elementName[i]))...)

		if len(element) > 0 {
			result = append(result, element...)
		} else {
			result = append(result, openBracket...)
			result = append(result, closeBracket...)
		}

		result = append(result, comma...)
	}

	// remove trailing "," if one exists
	l := len(result)
	if string(result[l-1]) == "," {
		result = result[:l-1]
		return append(result, closeBracket...), nil
	}
	return append(result, closeBracket...), nil
}
