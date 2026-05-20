package main

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: tap <host> <ports>\n")
		fmt.Fprintf(os.Stderr, "Example: tap localhost 80,443,8000-9000\n")
		os.Exit(1)
	}

	host := os.Args[1]
	ports, err := parsePorts(os.Args[2])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid ports: %v\n", err)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 100)

	for _, port := range ports {
		wg.Add(1)
		sem <- struct{}{}
		go func(p int) {
			defer wg.Done()
			defer func() { <-sem }()
			addr := net.JoinHostPort(host, strconv.Itoa(p))
			conn, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
			if err == nil {
				name := serviceName(p)
				if name != "" {
					fmt.Printf("%d open (%s)\n", p, name)
				} else {
					fmt.Printf("%d open\n", p)
				}
				conn.Close()
			}
		}(port)
	}

	wg.Wait()
}

func serviceName(port int) string {
	m := map[int]string{
		20: "ftp-data", 21: "ftp", 22: "ssh", 23: "telnet",
		25: "smtp", 53: "dns", 80: "http", 110: "pop3",
		143: "imap", 443: "https", 465: "smtps", 587: "smtp",
		993: "imaps", 995: "pop3s", 1433: "mssql", 1521: "oracle",
		2049: "nfs", 2375: "docker", 2376: "docker-tls",
		3306: "mysql", 3389: "rdp", 5432: "postgres",
		6379: "redis", 6443: "https-alt", 8080: "http-alt",
		8443: "https-alt", 9092: "kafka", 27017: "mongodb",
	}
	return m[port]
}

func parsePorts(s string) ([]int, error) {
	var ports []int
	parts := strings.Split(s, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "-") {
			ends := strings.SplitN(part, "-", 2)
			start, err := strconv.Atoi(strings.TrimSpace(ends[0]))
			if err != nil {
				return nil, err
			}
			end, err := strconv.Atoi(strings.TrimSpace(ends[1]))
			if err != nil {
				return nil, err
			}
			for i := start; i <= end; i++ {
				ports = append(ports, i)
			}
		} else {
			p, err := strconv.Atoi(part)
			if err != nil {
				return nil, err
			}
			ports = append(ports, p)
		}
	}
	return ports, nil
}
