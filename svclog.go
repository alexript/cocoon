// MIT License
//
// Copyright (c) 2018 Alexander Malyshev <alexript@outlook.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
//

package cocoon

/*
based on https://github.com/google/logger
*/

import (
	"fmt"
	"strings"

	"golang.org/x/sys/windows"

	"golang.org/x/sys/windows/svc/eventlog"
)

type severity int

// Severity levels.
const (
	sInfo severity = iota
	sWarning
	sError
	sFatal
)

type prependedRecord struct {
	Level severity
	Text  string
}

var (
	logLevel      severity = sWarning
	svclogWriter  *eventlog.Log
	regTitle      string = `Cocoon`
	messagePrefix string = ""
	buffer        []prependedRecord
)

func ParseLogLevel(levelString string) severity {
	switch levelString {
	case "info":
		return sInfo
	case "warning":
		return sWarning
	case "error":
		return sError
	default:
		return sError
	}
}

func SetLogLevel(level severity) {
	logLevel = level
}

func GetLogLevel() severity {
	return logLevel
}

func newWriter(src string) (*eventlog.Log, error) {
	// Continue if we receive "registry key already exists" or if we get
	// ERROR_ACCESS_DENIED so that we can log without administrative permissions
	// for pre-existing eventlog sources.
	if err := eventlog.InstallAsEventCreate(src, eventlog.Info|eventlog.Warning|eventlog.Error); err != nil {
		if !strings.Contains(err.Error(), "registry key already exists") && err != windows.ERROR_ACCESS_DENIED {
			return nil, err
		}
	}
	el, err := eventlog.Open(src)
	if err != nil {
		return nil, err
	}
	return el, nil
}

func Initlog(title string, exeFileName string) {
	regTitle = title
	messagePrefix = exeFileName + ": "
	w, err := newWriter(regTitle)
	if err != nil {
		abort("InitLog", err)
	}
	svclogWriter = w

	if buffer != nil && len(buffer) > 0 {
		for _, record := range buffer {
			writeToLog(record.Level, record.Text)
		}
		buffer = nil
	}

}

func CloseLog() {
	if svclogWriter != nil {
		svclogWriter.Close()
	}
}

func writeToLog(level severity, text string) {

	if level < logLevel {
		return
	}

	if svclogWriter == nil {
		buffer = append(buffer, prependedRecord{Level: level, Text: text})
		return
	}

	switch level {
	case sInfo:
		svclogWriter.Info(1, messagePrefix+text)
		return
	case sWarning:
		svclogWriter.Warning(3, messagePrefix+text)
		return
	case sError:
		svclogWriter.Error(2, messagePrefix+text)
		return
	}
	LogFatal(fmt.Sprintf("unrecognized severity: %v", level))
}

func LogInfo(v ...interface{}) {
	writeToLog(sInfo, fmt.Sprint(v...))
}
func LogWarning(v ...interface{}) {
	writeToLog(sWarning, fmt.Sprint(v...))
}
func LogError(v ...interface{}) {
	writeToLog(sError, fmt.Sprint(v...))
}

func LogFatal(v ...interface{}) {
	panic(fmt.Sprintf("svclog failed: %v", v...))
}
