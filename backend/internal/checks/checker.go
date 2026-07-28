package checks

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

func CheckPing(target string) (bool, int) {
	start := time.Now()
	cmd := exec.Command("ping", "-c", "1", "-W", "2", target)
	err := cmd.Run()
	elapsed := int(time.Since(start).Milliseconds())
	if err != nil {
		return false, elapsed
	}
	return true, elapsed
}

func CheckTCP(target string, port int) (bool, int) {
	start := time.Now()
	address := net.JoinHostPort(target, fmt.Sprintf("%d", port))
	conn, err := net.DialTimeout("tcp", address, 5*time.Second)
	elapsed := int(time.Since(start).Milliseconds())
	if err != nil {
		return false, elapsed
	}
	defer conn.Close()
	return true, elapsed
}

func CheckDNS(target, expected string) (bool, int) {
	start := time.Now()
	ips, err := net.LookupIP(target)
	elapsed := int(time.Since(start).Milliseconds())
	if err != nil {
		return false, elapsed
	}
	if expected == "" {
		return len(ips) > 0, elapsed
	}
	for _, ip := range ips {
		if ip.String() == expected {
			return true, elapsed
		}
	}
	return false, elapsed
}

func CheckHTTP(targetURL, expectedBody string) (bool, int) {
	start := time.Now()
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(targetURL)
	elapsed := int(time.Since(start).Milliseconds())
	if err != nil {
		return false, elapsed
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return false, elapsed
	}
	if expectedBody != "" {
		buf := make([]byte, 8192)
		n, _ := resp.Body.Read(buf)
		if !strings.Contains(string(buf[:n]), expectedBody) {
			return false, elapsed
		}
	}
	return true, elapsed
}

func CheckSSLExpiry(target string, port int) (bool, int, time.Time) {
	start := time.Now()
	address := net.JoinHostPort(target, fmt.Sprintf("%d", port))
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: 10 * time.Second}, "tcp", address, &tls.Config{
		InsecureSkipVerify: true,
	})
	elapsed := int(time.Since(start).Milliseconds())
	if err != nil {
		return false, elapsed, time.Time{}
	}
	defer conn.Close()
	certs := conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return false, elapsed, time.Time{}
	}
	return true, elapsed, certs[0].NotAfter
}
