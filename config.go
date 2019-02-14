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
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	initfile "gopkg.in/ini.v1"
)

// GetConfigFileName returns config file name
func GetConfigFileName() string {
	name, err := GetMyselfName()
	if err != nil {
		name = "cocoon.exe"
	}
	return name + ".ini"
}

func findRuntimeVersion(folder string) string {
	files, err := ioutil.ReadDir(folder)
	if err != nil {
		return ""
	}
	dirs := make([]os.FileInfo, 0)
	for _, v := range files {
		if v.IsDir() {
			dirs = append(dirs, v)
		}
	}

	if len(dirs) < 1 {
		return ""
	}

	if len(dirs) == 1 {
		return filepath.Base(dirs[0].Name())
	}

	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].ModTime().Unix() > dirs[j].ModTime().Unix()
	})
	return filepath.Base(dirs[0].Name())
}

func findRuntimeScript(defaultRuntimeFolder, defaultVersion, defaultName string) string {
	runtimeDir := filepath.Join(GetMyselfDir(), defaultRuntimeFolder, defaultVersion)
	files, err := ioutil.ReadDir(runtimeDir)
	if err != nil {
		return defaultName
	}

	scripts := make([]os.FileInfo, 0)
	for _, v := range files {
		ext := path.Ext(v.Name())
		if strings.EqualFold(ext, ".cmd") {
			scripts = append(scripts, v)
		}
	}

	if len(scripts) < 1 {
		return defaultName
	}
	return scripts[0].Name()
}

func findLarvaScript() string {
	return filepath.Base(GetCocoonAssetName(".cmd"))
}

func findLogFilename() string {
	result := "CocoonProcess"

	return result
}

func createDefaultConfig(configFileName string) {
	cfg := initfile.Empty()

	cfg.Section("cocoon").Key("startup").SetValue("cocoon_init.cmd")
	cfg.Section("cocoon").Key("log.file").SetValue(findLogFilename())
	cfg.Section("cocoon").Key("log.level").SetValue("error")
	cfg.Section("cocoon").Key("usepipe").SetValue("no")

	defaultRuntimeFolder := "runtime"
	defaultVersion := findRuntimeVersion(defaultRuntimeFolder)

	cfg.Section("chrysalis").Key("dir.base").SetValue(defaultRuntimeFolder)
	cfg.Section("chrysalis").Key("dir.version").SetValue(defaultVersion)
	cfg.Section("chrysalis").Key("dir.64bit").SetValue("x64")
	cfg.Section("chrysalis").Key("dir.32bit").SetValue("i586")
	cfg.Section("chrysalis").Key("initscript").SetValue(findRuntimeScript(defaultRuntimeFolder, defaultVersion, "init.cmd"))

	cfg.Section("larva").Key("appdir").SetValue(".")
	cfg.Section("larva").Key("startup").SetValue(findLarvaScript())

	MetamorphoseDate(cfg)

	err := cfg.SaveTo(configFileName)
	if err != nil {
		abort("createDefaultConfig", err)
	}

}

// HasConfig checks is config file is available
func HasConfig() bool {
	configFileName := GetConfigFileName()

	if _, err := os.Stat(configFileName); os.IsNotExist(err) {
		return false
	}

	return true
}

// LoadConfig if config file is not exists then creates default and load config file
func LoadConfig() (*initfile.File, error) {
	configFileName := GetConfigFileName()

	if _, err := os.Stat(configFileName); os.IsNotExist(err) {
		createDefaultConfig(configFileName)
	}

	return initfile.Load(configFileName)
}
