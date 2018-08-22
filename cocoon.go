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
	"bytes"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	initfile "gopkg.in/ini.v1"
	validator "gopkg.in/validator.v2"
)

// InitCocoon initialize cocoon
func InitCocoon() {
	_ = validator.SetValidationFunc("fileExists", fileExists)
}

// GetMyselfName returns abs filepath to executable file
func GetMyselfName() (string, error) {
	prog := os.Args[0]
	p, err := filepath.Abs(prog)
	return p, err
}

// GetMyselfDir returns folder where executable file is.
func GetMyselfDir() string {
	myselfName, _ := GetMyselfName()
	myselfDir := filepath.Dir(myselfName)
	return myselfDir
}

// GetCocoonAssetName returns filename, produced by cocoon exe file name without extension + given extension
func GetCocoonAssetName(extension string) string {
	myname, err := GetMyselfName()
	if err != nil {
		myname = "cocoon.exe"
	}
	ext := path.Ext(myname)
	return myname[0:len(myname)-len(ext)] + extension
}

func is64bitToString(is bool) string {
	if is {
		return "64bit"
	}
	return "32bit"

}

// GetAbsolutePath transforms given path to the absolute path in myself dir if relative. Returns as is if path is absolute.
func GetAbsolutePath(somepath string) string {
	isAbsPath := filepath.IsAbs(somepath)
	if isAbsPath {
		return somepath
	}

	return filepath.Join(GetMyselfDir(), somepath)

}

func getCocoonInitScript(cfg *initfile.File) string {
	initScriptName := cfg.Section("cocoon").Key("startup").Validate(func(in string) string {
		if len(in) == 0 {
			return "cocoon_init.cmd"
		}
		return in
	})
	return GetAbsolutePath(initScriptName)

}

func getCocoonLogFilename(cfg *initfile.File) string {
	logName := cfg.Section("cocoon").Key("log.file").Validate(func(in string) string {
		if len(in) == 0 {
			return "CocoonProcess"
		}
		return in
	})
	return logName
}

func getCocoonLogLevel(cfg *initfile.File) string {
	logLevel := cfg.Section("cocoon").Key("log.level").Validate(func(in string) string {
		if len(in) == 0 {
			return "error"
		}
		return in
	})
	return logLevel
}

func getCocoonUsepipe(cfg *initfile.File) bool {
	usePipe := cfg.Section("cocoon").Key("usepipe").Validate(func(in string) string {
		if len(in) == 0 {
			return "false"
		}
		return in
	})
	return strings.EqualFold(usePipe, "true") || strings.EqualFold(usePipe, "yes")
}

func getChrystalisVersionPath(cfg *initfile.File) string {
	basedir := cfg.Section("chrysalis").Key("dir.base").Validate(func(in string) string {
		if len(in) == 0 {
			return "runtime"
		}
		return in
	})

	versiondir := cfg.Section("chrysalis").Key("dir.version").String()
	return filepath.Join(basedir, versiondir)
}

func getChrystalisPath(cfg *initfile.File, is64bit bool) string {
	versiondir := getChrystalisVersionPath(cfg)

	archsuffix := is64bitToString(is64bit)
	archdir := cfg.Section("chrysalis").Key("dir." + archsuffix).Validate(func(in string) string {
		if len(in) == 0 {
			return archsuffix
		}
		return in
	})

	return GetAbsolutePath(filepath.Join(versiondir, archdir))
}

func getChrystalisInitScript(cfg *initfile.File) string {
	versiondir := getChrystalisVersionPath(cfg)
	initScriptName := cfg.Section("chrysalis").Key("initscript").Validate(func(in string) string {
		if len(in) == 0 {
			return "init.cmd"
		}
		return in
	})
	return GetAbsolutePath(filepath.Join(versiondir, initScriptName))
}

func getLarvaPath(cfg *initfile.File) string {
	basedir := cfg.Section("larva").Key("appdir").Validate(func(in string) string {
		if len(in) == 0 {
			return "."
		}
		return in
	})

	return GetAbsolutePath(basedir)
}

func getLarvaStartupScript(cfg *initfile.File) string {
	larvaPath := getLarvaPath(cfg)
	initScriptName := cfg.Section("larva").Key("startup").Validate(func(in string) string {
		if len(in) == 0 {
			return "larva.cmd"
		}
		return in
	})
	return GetAbsolutePath(filepath.Join(larvaPath, initScriptName))
}

// Cocoon configuration structure
type Cocoon struct {
	ArchStr           string
	Path              string `validate:"fileExists"`
	Startup           string
	ChrystalisPath    string `validate:"fileExists"`
	ChrystalisStartup string
	LarvaPath         string `validate:"fileExists"`
	LarvaStartup      string `validate:"fileExists"`
	UsePipe           bool
}

func (cocoon Cocoon) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("Arch: %s\n", cocoon.ArchStr))
	buffer.WriteString(fmt.Sprintf("Cocoon path: %s\n", cocoon.Path))
	buffer.WriteString(fmt.Sprintf("Cocoon init script: %s\n", cocoon.Startup))
	buffer.WriteString(fmt.Sprintf("Cocoon use pipes: %v\n", cocoon.UsePipe))
	buffer.WriteString(fmt.Sprintf("Chrystalis Path: %s\n", cocoon.ChrystalisPath))
	buffer.WriteString(fmt.Sprintf("Chrystalis Init script: %s\n", cocoon.ChrystalisStartup))
	buffer.WriteString(fmt.Sprintf("Larva path: %s\n", cocoon.LarvaPath))
	buffer.WriteString(fmt.Sprintf("Larva startup script: %s\n", cocoon.LarvaStartup))

	return buffer.String()
}

// NewCocoon creates new Coocon object, based on config file content.
func NewCocoon(cfg *initfile.File, is64bit bool) Cocoon {
	return Cocoon{
		ArchStr:           is64bitToString(is64bit),
		Path:              GetMyselfDir(),
		Startup:           getCocoonInitScript(cfg),
		ChrystalisPath:    getChrystalisPath(cfg, is64bit),
		ChrystalisStartup: getChrystalisInitScript(cfg),
		LarvaPath:         getLarvaPath(cfg),
		LarvaStartup:      getLarvaStartupScript(cfg),
		UsePipe:           getCocoonUsepipe(cfg),
	}
}

// DefaultCocoon creates new Cocoon object with default values.
func DefaultCocoon(startupCmdFile string, is64bit bool) Cocoon {
	return Cocoon{
		ArchStr:           is64bitToString(is64bit),
		Path:              GetMyselfDir(),
		Startup:           "",
		ChrystalisPath:    "",
		ChrystalisStartup: "",
		LarvaPath:         GetMyselfDir(),
		LarvaStartup:      startupCmdFile,
		UsePipe:           false,
	}
}

func fileExists(v interface{}, param string) error {
	st := reflect.ValueOf(v)
	if st.Kind() != reflect.String {
		return errors.New("fileExists only validates strings")
	}

	if _, err := os.Stat(st.String()); os.IsNotExist(err) {
		return errors.New("file or path does not exists")
	}

	return nil
}
