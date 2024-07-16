package netx

import (
	"net"
	"net/url"
	"time"
)

func Ping(urlStr string) error {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return err
	}
	address := parsedURL.Hostname()
	port := parsedURL.Port()
	if len(port) == 0 {
		if parsedURL.Scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	address = net.JoinHostPort(address, port)
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}
