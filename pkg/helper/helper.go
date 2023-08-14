// Package helpers are general functions shared with all parts of the web application.
package helper

import (
	"net"
	"os"
)

const (
	// Eraseline is an ANSI escape control to erase the active line of the terminal.
	Eraseline = "\x1b[2K"
	// byteUnits is a list of units used for formatting byte sizes.
	byteUnits = "kMGTPE"
)

// GetLocalIPs returns a list of local IP addresses.
// credit: https://gosamples.dev/local-ip-address/
func GetLocalIPs() ([]net.IP, error) {
	var ips []net.IP
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, addr := range addresses {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP)
			}
		}
	}
	return ips, nil
}

// GetLocalHosts returns a list of local hostnames.
func GetLocalHosts() ([]string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	hosts := []string{}
	hosts = append(hosts, hostname)
	// confirm localhost is resolvable
	if _, err = net.LookupHost("localhost"); err != nil {
		return nil, err
	}
	hosts = append(hosts, "localhost")
	return hosts, nil
}
