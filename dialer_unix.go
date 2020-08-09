// +build darwin linux

package main

import (
	"net"
)

func dialDockerSocket() (net.Conn, error) {
	return net.Dial("unix", "/var/run/docker.sock")
}
