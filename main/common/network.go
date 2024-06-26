package common

import (
	e "XrayHelper/main/errors"
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	tagNetwork = "network"
	timeout    = 3000
	dns        = "223.5.5.5:53"
	dns6       = "2400:3200::1"
)

// getHttpClient get a http client with custom dns
func getHttpClient(dns string, timeout time.Duration) *http.Client {
	http.DefaultTransport.(*http.Transport).DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		dialer := &net.Dialer{
			Resolver: &net.Resolver{
				PreferGo: true,
				Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
					d := net.Dialer{Timeout: timeout}
					return d.DialContext(ctx, "udp", dns)
				},
			},
		}
		return dialer.DialContext(ctx, network, addr)
	}
	return &http.Client{}
}

// CheckPort check whether the port is listening
func CheckPort(protocol string, host string, port string) bool {
	addr := net.JoinHostPort(host, port)
	conn, err := net.DialTimeout(protocol, addr, 1*time.Second)
	if err != nil {
		return false
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			return
		}
	}(conn)
	return true
}

// CheckLocalPort check whether the local port is listening
func CheckLocalPort(port string) bool {
	var msg bytes.Buffer
	NewExternal(0, &msg, nil, "netstat", "-tnlp").Run()
	return strings.Contains(msg.String(), port)
}

func IsIPv6(cidr string) bool {
	ip, _, _ := net.ParseCIDR(cidr)
	if ip != nil && ip.To4() == nil {
		return true
	}
	return false
}

func CheckIPv6Connection() bool {
	return CheckPort("udp", dns6, "53")
}

func CheckLocalDevice(dev string) bool {
	devices, err := net.Interfaces()
	if err != nil {
		return false
	}
	for _, device := range devices {
		if dev == device.Name {
			return true
		}
	}
	return false
}

// getExternalIPv6Addr get external ipv6 address, which should bypass
func getExternalIPv6Addr() ([]string, error) {
	var ipv6Addrs []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, e.New("cannot get ip address from local interface, ", err).WithPrefix(tagNetwork)
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && ipnet.IP.IsGlobalUnicast() {
			if ipnet.IP.To4() == nil {
				ipv6Addrs = append(ipv6Addrs, ipnet.IP.String())
			}
		}
	}
	return ipv6Addrs, nil
}

// DownloadFile download file from url, and save to filepath
func DownloadFile(filepath string, url string) error {
	// get file from url
	response, err := getHttpClient(dns, timeout*time.Millisecond).Get(url)
	if err != nil {
		return e.New("cannot get file "+url+", ", err).WithPrefix(tagNetwork)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)
	if response.StatusCode != http.StatusOK {
		return e.New("bad http status "+response.Status+", ", err).WithPrefix(tagNetwork)
	}
	// open saveFile
	saveFile, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_SYNC|os.O_TRUNC, 0755)
	if err != nil {
		return e.New("cannot open file "+filepath+", ", err).WithPrefix(tagNetwork)
	}
	defer func(saveFile *os.File) {
		_ = saveFile.Close()
	}(saveFile)
	_, err = io.Copy(saveFile, response.Body)
	if err != nil {
		return e.New("save file "+filepath+" failed, ", err).WithPrefix(tagNetwork)
	}
	return nil
}

// GetRawData get raw data from a url
func GetRawData(url string) ([]byte, error) {
	response, err := getHttpClient(dns, timeout*time.Millisecond).Get(url)
	if err != nil {
		return nil, e.New("cannot get url "+url+", ", err).WithPrefix(tagNetwork)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(response.Body)
	if response.StatusCode != http.StatusOK {
		return nil, e.New("bad http status "+response.Status+", ", err).WithPrefix(tagNetwork)
	}
	raw, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, e.New("read data failed, ", err).WithPrefix(tagNetwork)
	}
	return raw, nil
}
