// MIT License
//
// Copyright (c) 2018-2019 Alexander Malyshev <alexript@outlook.com>
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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/natefinch/npipe"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
	initfile "gopkg.in/ini.v1"
	validator "gopkg.in/validator.v2"
)

var (
	application    = kingpin.New(os.Args[0], "Cocoon")
	metamorphose   = application.Command("metamorphose", "Metamorphose larva to cocoon with chrysalis")
	morphCommand   = metamorphose.Command("morph", "Execute metamorpgose")
	cocoonStartup  = morphCommand.Flag("cocoon-startup", "Set cocoon preparation script name").String()
	cocoonLoglevel = morphCommand.Flag("cocoon-loglevel", "Set cocoon log level").Enum("info", "warning", "error")
	cocoonLogname  = morphCommand.Flag("cocoon-logname", "Set cocoon application name").String()
	cocoonUsepipe  = morphCommand.Flag("cocoon-usepipe", "Use pipe for communication from larva to cocoon").Enum("yes", "no", "true", "false")
	chrysalisDir   = morphCommand.Flag("chrysalis-dir", "Set chrysalis dir name").String()
	larvaStartup   = morphCommand.Flag("larva-startup", "Set larva run script").String()
	injectCommand  = metamorphose.Command("inject", "Inject new chrysalis into cocoon")
	injectName     = injectCommand.Arg("name", "New chrysalis name").Required().String()
	injectZip      = injectCommand.Arg("injected", "ZIP file with new chrysalis").Required().String()
	dropRuntimes   = injectCommand.Arg("dropOther", "Delete old chrysalises on success").Enum("yes", "no", "true", "false")

	larvaProcess *os.Process

	cocoonEnvPrefix = []string{
		"COCOON_",
		"JAVA_HOME",
		"JRE_HOME",
		"JDK_HOME",
		"JRE_ARCH_DIR",
		"RUNTIME_DIR",
		"STABLE_RUNTIME",
		"CURRENT_RUNTIME",
		"_ROOT",
	}
)

func appendScript(scriptName string, slice []string) []string {
	if _, err := os.Stat(scriptName); os.IsNotExist(err) {
		return slice
	}
	if len(slice) == 0 {
		return append(slice, scriptName)
	}
	return append(slice, "&&", scriptName)
}

func readConfiguration() *initfile.File {
	myName, _ := GetMyselfName()
	cfg, err := LoadConfig()
	if err != nil {
		showError(fmt.Sprintf("Fail to read file: %v", err))
		os.Exit(1)
	}
	params := os.Args[1:]
	exitFlag := 0

	if ShouldMetamorph(params) {
		appCommand, cmdErr := application.Parse(params)
		LogWarning("Metamorphose: " + appCommand)
		if cmdErr == nil {
			var cfgChanged = false
			switch appCommand {
			case "metamorphose morph":

				cfgChanged = cfgChanged || MetamorphoseCocoonStartup(*cocoonStartup, cfg)
				cfgChanged = cfgChanged || MetamorphoseCocoonLoglevel(*cocoonLoglevel, cfg)
				cfgChanged = cfgChanged || MetamorphoseCocoonLogname(*cocoonLogname, cfg)
				cfgChanged = cfgChanged || MetamorphoseCocoonUsepipe(*cocoonUsepipe, cfg)
				cfgChanged = cfgChanged || MetamorphoseChrysalisDir(*chrysalisDir, cfg)
				cfgChanged = cfgChanged || MetamorphoseLarvaStartup(*larvaStartup, cfg)
				break
			case "metamorphose inject":
				cfgChanged = cfgChanged || MetamorphoseInjectChrysalis(*injectName, *injectZip, *dropRuntimes, cfg)
				break
			}
			if cfgChanged {
				MetamorphoseDate(cfg)
				_ = cfg.SaveTo(GetConfigFileName())
				exitFlag = 1
			} else {
				LogError(fmt.Sprintf("Metamorphose: %v failed\n %v", appCommand, os.Args))
				exitFlag = 2
			}
		} else {
			kingpin.Usage()
		}
	}

	logFileName := getCocoonLogFilename(cfg)
	logLevel := getCocoonLogLevel(cfg)

	Initlog(logFileName, myName)
	SetLogLevel(ParseLogLevel(logLevel))

	if exitFlag > 0 {
		os.Exit(exitFlag - 1)
	}
	return cfg
}

func Stop() {
	if larvaProcess == nil {
		return
	}
	notifyChilds("event", "die")
	if larvaProcess != nil {
		larvaProcess.Kill()
	}
	larvaProcess = nil
}

func isCocoonEnv(envstring string) bool {
	result := false
	for _, v := range cocoonEnvPrefix {
		if strings.HasPrefix(envstring, v) {
			result = true
			break
		}
	}
	return result
}

func filterOutEnviron(orig []string) []string {
	filtered := []string{}
	for _, v := range orig {
		if !isCocoonEnv(v) {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

// Start cocoon container
func Start(startupCmdFile, logFileName string, cocoon *Cocoon) {
	Stop()
	isConsoleAttached := AttachConsole()

	defer FreeConsole()

	myName, _ := GetMyselfName()

	hasConfig := HasConfig()

	logLevel := "error"

	var cfg *initfile.File
	if hasConfig {
		cfg = readConfiguration()
	} else {
		Initlog(logFileName, myName)
		SetLogLevel(ParseLogLevel(logLevel))
	}

	params := os.Args[1:]

	f := func(r rune) bool {
		return r == ' '
	}
	for k, v := range params {
		if strings.IndexFunc(v, f) != -1 {
			params[k] = "\"" + params[k] + "\""
		}
	}

	defer CloseLog()

	LogInfo(fmt.Sprintf("Should be console attached: %v", isConsoleAttached))

	is64bit := Is64bitOS()

	LogInfo(fmt.Sprintf("Is 64bit environment: %v", is64bit))

	if cocoon == nil {
		*cocoon = DefaultCocoon(startupCmdFile, is64bit)
		if hasConfig && cfg != nil {
			InitCocoon()
			*cocoon = NewCocoon(cfg, is64bit)
		}
	}

	LogInfo(fmt.Sprintf("Cocoon info:\n%v\n", cocoon))

	var stdout *os.File
	var stderr *os.File
	var stdError error

	outputsPath := cocoon.LogPath
	if len(outputsPath) < 1 {
		outputsPath = cocoon.LarvaPath
	}

	if hasConfig {
		stdout, stderr, stdError = GetOutputs(outputsPath, logFileName)
	} else {
		stdout, stderr, stdError = GetOutputs(outputsPath, filepath.Base(myName))
	}

	if stdError != nil {
		abort("std redirector", stdError)
	}

	os.Stdout = stdout
	os.Stderr = stderr

	if hasConfig {
		if errs := validator.Validate(*cocoon); errs != nil {
			showError(fmt.Sprintf("Cocoon errors: %v\n", errs))
			os.Exit(1)
		}
	}

	arguments := append(params, fmt.Sprintf("--cocoon-pid=%v", syscall.Getpid()))

	var pipeListener *npipe.PipeListener
	if cocoon.UsePipe {

		pipeListener = ListenNpipe()
		if pipeListener != nil {
			LogInfo(fmt.Sprintf("Start NPipe listener on: %v", GetNpipeName()))
			defer pipeListener.Close()
			arguments = append(arguments, fmt.Sprintf("--cocoon-npipe=%v", GetNpipeName()))
		}

	}

	LogInfo(fmt.Sprintf("Cocoon arguments: %v", params))

	procAttr := new(os.ProcAttr)
	procAttr.Files = []*os.File{os.Stdin, stdout, stderr}
	procAttr.Env = append(filterOutEnviron(os.Environ()), []string{

		fmt.Sprintf("COCOON_PID=%v", syscall.Getpid()),
		"COCOON_ARCH=" + cocoon.ArchStr,
		"COCOON_PATH=" + cocoon.Path,
		"COCOON_RUNTIME=" + cocoon.ChrystalisPath,
		"COCOON_APPDIR=" + cocoon.LarvaPath,
		"COCOON_ARGUMENTS=" + strings.Join(arguments[:], " "),
		"COCOON_EXE=" + myName,
	}...)
	if cocoon.UsePipe {
		procAttr.Env = append(procAttr.Env, "COCOON_PIPE="+GetNpipeName())
	}
	procAttr.Dir = cocoon.LarvaPath

	var scripts []string
	if hasConfig {
		scripts = appendScript(cocoon.Startup, scripts)
		scripts = appendScript(cocoon.ChrystalisStartup, scripts)
	}
	scripts = appendScript(cocoon.LarvaStartup, scripts)

	_ = stdout.Sync()
	_ = stderr.Sync()
	larvaProcess := StartCmdScripts(scripts, procAttr)
	if larvaProcess != nil {
		larvaProcess.Wait()
	}
	_ = stdout.Sync()
	_ = stderr.Sync()
}
