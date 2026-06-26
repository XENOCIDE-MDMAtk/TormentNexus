package tools

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"
)

// HandleCheckPort checks if a specific TCP port is open on a host.
func HandleCheckPort(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	port, _ :=getInt(args, "port")
	timeout, _ :=getInt(args, "timeout")

	if timeout == 0 {
		timeout = 5
	}

	address := fmt.Sprintf("%s:%d", host, port)
	
	conn, dialErr := net.DialTimeout("tcp", address, time.Duration(timeout)*time.Second)
	if dialErr != nil {
		return err(fmt.Sprintf("Failed to connect to %s: %v", address, dialErr))
}

	defer conn.Close()

	return ok(fmt.Sprintf("Successfully connected to %s", address))
}

// HandleScanPorts scans a range of TCP ports on a host to find open ones.
func HandleScanPorts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	host, _ :=getString(args, "host")
	startPort, _ :=getInt(args, "start_port")
	endPort, _ :=getInt(args, "end_port")
	timeout, _ :=getInt(args, "timeout")

	if timeout == 0 {
		timeout = 1 // Shorter default timeout for scanning
	}

	var openPorts []string

	for port := startPort; port <= endPort; port++ {
		address := fmt.Sprintf("%s:%d", host, port)
		conn, dialErr := net.DialTimeout("tcp", address, time.Duration(timeout)*time.Second)
		if dialErr == nil {
			conn.Close()
			openPorts = append(openPorts, fmt.Sprintf("%d", port))

	}

	if len(openPorts) == 0 {
		return ok(fmt.Sprintf("No open ports found on %s in range %d-%d", host, startPort, endPort))
}

	return ok(fmt.Sprintf("Open ports on %s: %s", host, strings.Join(openPorts, ", ")))
}
}