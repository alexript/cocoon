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

import (
	"log"
	"os"
	"path/filepath"

	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)

var (
	errLog  *log.Logger
	warnLog *log.Logger
)

func GetOutputs(folder string, filenamePrefix string) (stdout *os.File, stderr *os.File, err error) {
	path, err := filepath.Abs(folder)
	if err != nil {
		return nil, nil, err
	}

	possibleLogSubdirName := filepath.Join(path, "logs")
	if di, err := os.Stat(possibleLogSubdirName); err == nil {
		if di.IsDir() {
			path = possibleLogSubdirName
		}
	}

	stdoutName := filepath.Join(path, filenamePrefix+".stdout")
	stderrName := filepath.Join(path, filenamePrefix+".stderr")

	warnRotator := &lumberjack.Logger{
		Filename:   stdoutName,
		MaxSize:    1,  // megabytes after which new file is created
		MaxBackups: 3,  // number of backups
		MaxAge:     28, //days
		Compress:   true,
	}
	warnLog = log.New(warnRotator, "", log.Ldate|log.Ltime)
	errRotator := &lumberjack.Logger{
		Filename:   stderrName,
		MaxSize:    1,  // megabytes after which new file is created
		MaxBackups: 3,  // number of backups
		MaxAge:     28, //days
		Compress:   true,
	}
	errLog = log.New(errRotator, "", log.Ldate|log.Ltime)

	warnLog.Println("stdout attached")
	errLog.Println("stderr attached")

	err = warnRotator.Close()
	if err != nil {
		return nil, nil, err
	}

	err = errRotator.Close()
	if err != nil {
		return nil, nil, err
	}

	stdout, err = os.OpenFile(stdoutName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		return nil, nil, err
	}

	stderr, err = os.OpenFile(stderrName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0660)
	if err != nil {
		return nil, nil, err
	}
	return stdout, stderr, nil
}
