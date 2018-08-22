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

// cocoon.exe metamorphose morph
//	--cocoon-startup=cocoon.cmd
//	--cocoon-loglevel=info
//  --cocoon-logname=MyApplication
//	--cocoon-usepipe=yes
//	--chrysalis-dir=jre8u172
//	--larva-startup=run.cmd

// cocoon.exe metamorphose inject jre8u1777 d:\Temp\downloaded\jre8_2018-04-25.zip yes

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	initfile "gopkg.in/ini.v1"
)

func ShouldMetamorf(params []string) bool {
	for _, a := range params {
		if a == "metamorphose" {
			return true
		}
	}
	return false
}

func logMorphInfo(section, key, value string) {
	LogWarning(fmt.Sprintf("Applyed new metamorphose: [%v] %v -> %v", section, key, value))
}

func MetamorphoseCocoonStartup(value string, cfg *initfile.File) bool {
	if value == "" {
		return false
	}
	cfg.Section("cocoon").Key("startup").SetValue(value)
	logMorphInfo("cocoon", "startup", value)
	return true

}

func MetamorphoseCocoonLoglevel(value string, cfg *initfile.File) bool {
	if value == "" {
		return false
	}
	cfg.Section("cocoon").Key("log.level").SetValue(value)
	logMorphInfo("cocoon", "log.level", value)
	return true

}

func MetamorphoseCocoonLogname(value string, cfg *initfile.File) bool {
	if value == "" {
		return false
	}
	cfg.Section("cocoon").Key("log.file").SetValue(value)
	logMorphInfo("cocoon", "log.file", value)
	return true

}

func MetamorphoseCocoonUsepipe(value string, cfg *initfile.File) bool {
	if value == "" {
		return false
	}
	cfg.Section("cocoon").Key("usepipe").SetValue(value)
	logMorphInfo("cocoon", "usepipe", value)
	return true
}

func MetamorphoseChrysalisDir(value string, cfg *initfile.File) bool {
	if value == "" {
		return false
	}
	baseDir := cfg.Section("chrysalis").Key("dir.base").String()
	descriptionFile := filepath.Join(baseDir, value, "chrysalis.ini")
	if _, err := os.Stat(descriptionFile); err == nil {
		description, descErr := initfile.Load(descriptionFile)
		if descErr == nil {
			cfg.Section("chrysalis").Key("dir.64bit").SetValue(description.Section("").Key("dir.64bit").String())
			cfg.Section("chrysalis").Key("dir.32bit").SetValue(description.Section("").Key("dir.32bit").String())
			cfg.Section("chrysalis").Key("initscript").SetValue(description.Section("").Key("initscript").String())
		}
	}
	cfg.Section("chrysalis").Key("dir.version").SetValue(value)
	logMorphInfo("chrysalis", "dir.version", value)
	return true

}

func MetamorphoseLarvaStartup(value string, cfg *initfile.File) bool {
	if value != "" {
		cfg.Section("larva").Key("startup").SetValue(value)
		logMorphInfo("larva", "startup", value)
		return true
	}
	return false
}

func MetamorphoseInjectChrysalis(injectName, injectZip, dropRuntimes string, cfg *initfile.File) bool {
	// if injectZip exists -> unpack into runtime folder
	// if unpacked sucessfully -> apply new runtime into config
	// on success and if dropRuntimes 'yes' or 'true' -> delete other runtimes

	deleteOther := strings.EqualFold(dropRuntimes, "true") || strings.EqualFold(dropRuntimes, "yes")

	basedir := cfg.Section("chrysalis").Key("dir.base").Validate(func(in string) string {
		if len(in) == 0 {
			return "runtime"
		}
		return in
	})

	path, errA := filepath.Abs(basedir)
	if errA != nil {
		return false
	}
	zipfile, errB := filepath.Abs(injectZip)
	if errB != nil {
		return false
	}

	if zi, err1 := os.Stat(zipfile); err1 == nil {
		if !zi.IsDir() {
			if di, err2 := os.Stat(path); err2 == nil {
				if di.IsDir() {
					// now we can upack <zipfile> into <path> and register with 'dir.version' = <injectName>
					if metamorphUnpack(zipfile, path, injectName) {
						if MetamorphoseChrysalisDir(injectName, cfg) {
							if deleteOther {
								metamorphDeleteCrysalisesExcept(path, injectName)
								// we don't care about delete success
							}
							return true
						}
					}
				}
			}
		}
	}

	return false
}

func metamorphDeleteCrysalisesExcept(path, exceptionName string) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return
	}
	dirs := make([]string, 0)
	for _, v := range files {
		if v.IsDir() && v.Name() != exceptionName {
			dirs = append(dirs, filepath.Join(path, v.Name()))
		}
	}

	if len(dirs) < 1 {
		return
	}

	for _, dirName := range dirs {
		os.RemoveAll(dirName)
	}
}

func metamorphUnpack(zipfile, path, unpackedFolderName string) bool {
	target := filepath.Join(path, unpackedFolderName)
	if _, err1 := os.Stat(target); err1 == nil {
		return false
	}
	err := Unzip(zipfile, target)
	if err != nil {
		LogError(err)
	}
	return err == nil
}

func MetamorphoseDate(cfg *initfile.File) {
	cfg.Section("metamorphosis").Key("date").SetValue(time.Now().String())
}
