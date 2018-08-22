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
	"bufio"
	"fmt"
	"net"
	"syscall"

	"github.com/natefinch/npipe"
)

// GetNpipeName constructs npipe name by pid
func GetNpipeName() string {
	pid := syscall.Getpid()
	return fmt.Sprintf("\\\\.\\pipe\\cocoon_%v", pid)
}

// ListenNpipe starts npipe listener
func ListenNpipe() *npipe.PipeListener {
	ln, err := npipe.Listen(GetNpipeName())
	if err != nil {
		return nil
	}

	go func(ln *npipe.PipeListener) {
		for {
			conn, err := ln.Accept()
			if err != nil {
				// handle error
				continue
			}

			// handle connection like any other net.Conn
			go func(conn net.Conn) {
				r := bufio.NewReader(conn)
				msg, err := r.ReadString('\n')
				if err != nil {
					// handle error
					return
				}
				go DefaultMessageBox("NPipe message", msg)

			}(conn)
		}
	}(ln)
	return ln
}
