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
	ATTACH_PARENT_PROCESS = ^uint32(0) // (DWORD)-1

	MB_OK                = 0x00000000
	MB_OKCANCEL          = 0x00000001
	MB_ABORTRETRYIGNORE  = 0x00000002
	MB_YESNOCANCEL       = 0x00000003
	MB_YESNO             = 0x00000004
	MB_RETRYCANCEL       = 0x00000005
	MB_CANCELTRYCONTINUE = 0x00000006
	MB_ICONHAND          = 0x00000010
	MB_ICONQUESTION      = 0x00000020
	MB_ICONEXCLAMATION   = 0x00000030
	MB_ICONASTERISK      = 0x00000040
	MB_USERICON          = 0x00000080
	MB_ICONWARNING       = MB_ICONEXCLAMATION
	MB_ICONERROR         = MB_ICONHAND
	MB_ICONINFORMATION   = MB_ICONASTERISK
	MB_ICONSTOP          = MB_ICONHAND

	MB_DEFBUTTON1 = 0x00000000
	MB_DEFBUTTON2 = 0x00000100
	MB_DEFBUTTON3 = 0x00000200
	MB_DEFBUTTON4 = 0x00000300

	CREATE_NO_WINDOW = 0x08000000
)

var (
	modkernel32       = syscall.NewLazyDLL("kernel32.dll")
	procAttachConsole = modkernel32.NewProc("AttachConsole")
	procFreeConsole   = modkernel32.NewProc("FreeConsole")
	user32            = syscall.NewLazyDLL("user32.dll")
	procMessageBox    = user32.NewProc("MessageBoxW")
)

func MessageBox(caption, text string, style uintptr) (result int) {
	ret, _, _ := procMessageBox.Call(0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(text))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(caption))),
		style)
	//	if callErr != nil {
	//		abort("Call MessageBox", callErr)
	//	}
	result = int(ret)
	return
}

func AttachConsole() (ok bool) {
	return attachConsole(ATTACH_PARENT_PROCESS)
}

func attachConsole(dwParentProcess uint32) (ok bool) {
	//r0, _, _ := syscall.Syscall(procAttachConsole.Addr(), 1, uintptr(dwParentProcess), 0, 0)
	r0, _, _ := procAttachConsole.Call(uintptr(dwParentProcess))
	ok = bool(r0 != 0)
	return
}

func FreeConsole() (ok bool) {
	//r0, _, _ := syscall.Syscall(procFreeConsole.Addr(), 0, 0, 0, 0)
	r0, _, _ := procFreeConsole.Call()
	ok = bool(r0 != 0)
	return
}

func Is64bitOS() bool {
	var is64bit bool = true

	procIsWow64Process := modkernel32.NewProc("IsWow64Process")
	//	fmt.Printf("%v\n", procIsWow64Process)
	handle, err := syscall.GetCurrentProcess()
	//	fmt.Printf("handle: %v\n", handle)
	if err != nil {
		is64bit = false
	} else {
		var result bool
		x, _, _ := procIsWow64Process.Call(uintptr(handle), uintptr(unsafe.Pointer(&result)))
		//fmt.Printf("%v\n", result)
		if x == 0 || !result {
			is64bit = false
		}

	}
	return is64bit
}

var zeroSysProcAttr syscall.SysProcAttr

func StartCmdScript(scriptName string, attr *os.ProcAttr) {
	if attr.Sys == nil {
		attr.Sys = &zeroSysProcAttr
	}
	attr.Sys.HideWindow = true
	attr.Sys.CreationFlags = attr.Sys.CreationFlags | CREATE_NO_WINDOW
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

func StartCmdScripts(scriptNames []string, attr *os.ProcAttr) {
	if attr.Sys == nil {
		attr.Sys = &zeroSysProcAttr
	}
	attr.Sys.HideWindow = true
	attr.Sys.CreationFlags = attr.Sys.CreationFlags | CREATE_NO_WINDOW

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
