package multicast

import (
	"net"
	"strconv"
	"sync"
	"time"
)

const (
	MulticastIPv4 = "224.0.0.9:32416"
)

//go:generate go install github.com/lemon-mint/vstruct/cli/vstruct@latest
//go:generate vstruct go multicast addrinfo.vstruct
//go:generate go mod tidy

type PeerInfo struct {
	Addr string
}

type Multicast struct {
	mu      sync.Mutex
	started bool
	stop    chan struct{}
	ln      *net.UDPConn
	addr    *net.UDPAddr
	C       chan PeerInfo
	myaddr  AddrInfo
}

func New(addr AddrInfo) *Multicast {
	return &Multicast{
		stop:   make(chan struct{}),
		C:      make(chan PeerInfo),
		myaddr: addr,
	}
}

func (m *Multicast) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.started {
		return nil
	}
	m.started = true
	addr, err := net.ResolveUDPAddr("udp", MulticastIPv4)
	if err != nil {
		return err
	}
	m.addr = addr
	ln, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		return err
	}
	m.ln = ln
	go m.startBroadcast()
	go m.startListen()
	return nil
}

func (m *Multicast) startBroadcast() {
	t := time.NewTicker(time.Second * 10)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			_, err := m.ln.WriteToUDP(m.myaddr, m.addr)
			if err != nil {
				return
			}
			println("Broadcast:", m.myaddr.String())
		case _, ok := <-m.stop:
			if !ok {
				// closed
				return
			}
		}
	}
}

func (m *Multicast) startListen() {
	buf := make([]byte, 1024)
	defer println("Multicast listener stopped")
	for {
		n, _, err := m.ln.ReadFromUDP(buf)
		if err != nil {
			return
		}
		data := AddrInfo(buf[:n])
		if !data.Vstruct_Validate() {
			println("Invalid data:", data)
			continue
		}
		println("Peer:", data.String())
		m.C <- PeerInfo{
			Addr: data.Ip() + ":" + strconv.Itoa(int(data.Port())),
		}
	}
}
func (m *Multicast) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.started {
		return
	}
	close(m.stop)
	m.ln.Close()
	close(m.C)
	m.started = false
}
