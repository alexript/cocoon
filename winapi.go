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
	"fmt"
	"os"

	"syscall"
	"unsafe"
)

const (
	attachParentProcess = ^uint32(0) // (DWORD)-1

	mbOk                = 0x00000000
	mbOkCancel          = 0x00000001
	mbAbortRetryIgnore  = 0x00000002
	mbYesNoCancel       = 0x00000003
	mbYesNo             = 0x00000004
	mbRetryCancel       = 0x00000005
	mbCancelTryContinue = 0x00000006
	mbIconHand          = 0x00000010
	mbIconQuestion      = 0x00000020
	mbIconExclamation   = 0x00000030
	mbIconAsterisk      = 0x00000040
	mbUserIcon          = 0x00000080
	mbIconWarning       = mbIconExclamation
	mbIconError         = mbIconHand
	mbIconInformation   = mbIconAsterisk
	mbIconStop          = mbIconHand

	mbDefButton1 = 0x00000000
	mbDefButton2 = 0x00000100
	mbDefButton3 = 0x00000200
	mbDefButton4 = 0x00000300

	createNoWindow = 0x08000000
)

var (
	modkernel32       = syscall.NewLazyDLL("kernel32.dll")
	procAttachConsole = modkernel32.NewProc("AttachConsole")
	procFreeConsole   = modkernel32.NewProc("FreeConsole")
	user32            = syscall.NewLazyDLL("user32.dll")
	procMessageBox    = user32.NewProc("MessageBoxW")
)

// DefaultMessageBox is win32 MessageBox in information mode
func DefaultMessageBox(caption, text string) (result int) {
	return messageBox(caption, text, mbOk|mbIconInformation)
}

// ErrorMessageBox is win32 MessageBox in error mode
func ErrorMessageBox(text string) (result int) {
	return messageBox("Error", text, mbOk|mbIconError)
}

// messageBox call win32 MessageBox and return int result
func messageBox(caption, text string, style uintptr) (result int) {
	ret, _, _ := procMessageBox.Call(0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(caption))),
		style)
	result = int(ret)
	return
}

// AttachConsole call win32 AttachConsole
func AttachConsole() (ok bool) {
	return attachConsole(attachParentProcess)
}

func attachConsole(dwParentProcess uint32) (ok bool) {
	r0, _, _ := procAttachConsole.Call(uintptr(dwParentProcess))
	ok = bool(r0 != 0)
	return
}

// FreeConsole call win32 FreeConsole
func FreeConsole() (ok bool) {
	r0, _, _ := procFreeConsole.Call()
	ok = bool(r0 != 0)
	return
}

// Is64bitOS try to call win32 IsWow64Process. Our application must be 32bit to have correct result.
func Is64bitOS() bool {
	is64bit := true

	procIsWow64Process := modkernel32.NewProc("IsWow64Process")
	handle, err := syscall.GetCurrentProcess()
	if err != nil {
		is64bit = false
	} else {
		var result bool
		x, _, _ := procIsWow64Process.Call(uintptr(handle), uintptr(unsafe.Pointer(&result)))
		if x == 0 || !result {
			is64bit = false
		}

	}
	return is64bit
}

var zeroSysProcAttr syscall.SysProcAttr

// StartCmdScript starts %COMSPEC% (cmd.exe expected) with scriptName as /C argument value.
func StartCmdScript(scriptName string, attr *os.ProcAttr) {
	if attr.Sys == nil {
		attr.Sys = &zeroSysProcAttr
	}
	attr.Sys.HideWindow = true
	attr.Sys.CreationFlags = attr.Sys.CreationFlags | createNoWindow
	if _, err := os.Stat(scriptName); err == nil {
		LogInfo(fmt.Sprintf("execute script: %v", scriptName))
		p, perr := os.StartProcess(os.Getenv("COMSPEC"), []string{"/C", scriptName}, attr)
		if perr != nil {
			LogError(perr)
		} else {
			p.Wait()
		}

	} else {
		LogError(err)
	}
}

func makeCmdLine(args []string) string {
	var s string
	for _, v := range args {
		if s != "" {
			s += " "
		}
		s += syscall.EscapeArg(v)
	}
	return `/C "` + s + `"`
}

// StartCmdScripts creates sequence '"script1" && "script2" && ...' and call %COMSPEC% (cmd.exe) with this sequence as /C value
func StartCmdScripts(scriptNames []string, attr *os.ProcAttr) {
	if attr.Sys == nil {
		attr.Sys = &zeroSysProcAttr
	}
	attr.Sys.HideWindow = true
	attr.Sys.CreationFlags = attr.Sys.CreationFlags | createNoWindow

	cmdLine := makeCmdLine(scriptNames)
	attr.Sys.CmdLine = cmdLine

	LogInfo(fmt.Sprintf("execute scripts touple: %v", cmdLine))
	p, perr := os.StartProcess(os.Getenv("COMSPEC"), nil, attr)
	if perr != nil {
		LogError(perr)
	} else {
		p.Wait()
	}

}
