// +build windows

package main

import (
	"github.com/Microsoft/go-winio"
	"net"
)

func dialDockerSocket() (net.Conn, error) {
	return winio.DialPipe(`\\.\pipe\docker_engine`, nil)
}
