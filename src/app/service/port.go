package service

import (
	"fmt"
	"net"
	"strconv"
)

const (
	DefaultPortStart = 50100
	DefaultPortEnd   = 50199
)

// PortAllocator manages port allocation for plugin services
type PortAllocator struct {
	portStart int
	portEnd   int
}

// NewPortAllocator creates a port allocator with the given range
func NewPortAllocator(start, end int) *PortAllocator {
	if start == 0 {
		start = DefaultPortStart
	}
	if end == 0 {
		end = DefaultPortEnd
	}
	return &PortAllocator{portStart: start, portEnd: end}
}

// AllocatePort finds the first available port not in the state file and not in use
func (pa *PortAllocator) AllocatePort(state *ProcessState) (int, error) {
	usedPorts := make(map[int]bool)
	for _, inst := range state.Instances {
		usedPorts[inst.Port] = true
	}

	for port := pa.portStart; port <= pa.portEnd; port++ {
		if usedPorts[port] {
			continue
		}
		if isPortFree(port) {
			return port, nil
		}
	}

	return 0, fmt.Errorf("no available ports in range %d-%d", pa.portStart, pa.portEnd)
}

// isPortFree checks if a TCP port is actually available
func isPortFree(port int) bool {
	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return false
	}
	ln.Close()
	return true
}
