package localaddrmanager

import (
	"errors"
	"fmt"
	"net"

	"github.com/B9O2/NStruct/Shield"
)

type LocalAddrManager struct {
	s    *Shield.Shield
	used map[int]bool
}

func (lam *LocalAddrManager) AllocatePort() (port int, err error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func (lam *LocalAddrManager) FreePort(port int) {
	lam.used[port] = false
}

func (lam *LocalAddrManager) GetIPFromNetInterface(inter net.Interface) (net.IP, bool) {
	if (inter.Flags & net.FlagUp) != 0 {
		addrs, _ := inter.Addrs()
		for _, address := range addrs {
			if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					return ipnet.IP, true
				}
			}
		}
	}
	return net.ParseIP("127.0.0.1"), false
}

func (lam *LocalAddrManager) GetLocalIP(interName string) (net.IP, error) {

	if interName == "" {
		netInterfaces, err := net.Interfaces()
		if err != nil {
			return net.ParseIP("127.0.0.1"), err
		}
		for _, inter := range netInterfaces {
			if res, ok := lam.GetIPFromNetInterface(inter); ok {
				return res, nil
			}
		}
	} else {
		inter, err := net.InterfaceByName(interName)
		if err != nil {
			return net.ParseIP("127.0.0.1"), err
		}
		if res, ok := lam.GetIPFromNetInterface(*inter); ok {
			return res, nil
		}
	}

	return net.ParseIP("127.0.0.1"), errors.New("no available ip")
}

func (lam *LocalAddrManager) GetLocalAddr(interName string) (*net.TCPAddr, error) {
	if port, err := lam.AllocatePort(); err != nil {
		return nil, err
	} else {
		ip, err := lam.GetLocalIP(interName)
		if err != nil {
			return nil, err
		}
		if addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", ip, port)); err != nil {
			return nil, err
		} else {
			return addr, nil
		}
	}
}

func (lam *LocalAddrManager) Close() {
	lam.s.Close()
}

func NewLocalAddrManager() *LocalAddrManager {
	return &LocalAddrManager{
		s:    Shield.NewShield(),
		used: make(map[int]bool),
	}
}
