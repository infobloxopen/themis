package pdp

import (
	"encoding/binary"
	"fmt"
	"github.com/miekg/bitradix"
	"net"
)

type SetOfNetworks struct {
	root *bitradix.Radix32
}

type NetworkLeafItem struct {
	Network *net.IPNet
	Leaf    interface{}
}

var (
	masks = []net.IPMask{
		net.CIDRMask(0x00, 32),
		net.CIDRMask(0x01, 32),
		net.CIDRMask(0x02, 32),
		net.CIDRMask(0x03, 32),
		net.CIDRMask(0x04, 32),
		net.CIDRMask(0x05, 32),
		net.CIDRMask(0x06, 32),
		net.CIDRMask(0x07, 32),
		net.CIDRMask(0x08, 32),
		net.CIDRMask(0x09, 32),
		net.CIDRMask(0x0a, 32),
		net.CIDRMask(0x0b, 32),
		net.CIDRMask(0x0c, 32),
		net.CIDRMask(0x0d, 32),
		net.CIDRMask(0x0e, 32),
		net.CIDRMask(0x0f, 32),
		net.CIDRMask(0x10, 32),
		net.CIDRMask(0x11, 32),
		net.CIDRMask(0x12, 32),
		net.CIDRMask(0x13, 32),
		net.CIDRMask(0x14, 32),
		net.CIDRMask(0x15, 32),
		net.CIDRMask(0x16, 32),
		net.CIDRMask(0x17, 32),
		net.CIDRMask(0x18, 32),
		net.CIDRMask(0x19, 32),
		net.CIDRMask(0x1a, 32),
		net.CIDRMask(0x1b, 32),
		net.CIDRMask(0x1c, 32),
		net.CIDRMask(0x1d, 32),
		net.CIDRMask(0x1e, 32),
		net.CIDRMask(0x1f, 32),
		net.CIDRMask(0x20, 32)}

	ErrorIPv6NotImplemented = fmt.Errorf("IPv6 and mixed network sets haven't been implemented yet")
)

func MakeNetwork(s string) (*net.IPNet, error) {
	addr := net.ParseIP(s)
	if addr != nil {
		if addr.To4() == nil {
			return nil, ErrorIPv6NotImplemented
		}

		s = fmt.Sprintf("%s/32", s)
	}

	_, n, err := net.ParseCIDR(s)
	if err != nil {
		return nil, fmt.Errorf("can't treat \"%s\" as IPv4 address or network", s)
	}

	if n.IP.To4() == nil {
		return nil, ErrorIPv6NotImplemented
	}

	return n, nil
}

func NewSetOfNetworks() *SetOfNetworks {
	return &SetOfNetworks{bitradix.New32()}
}

func netToKey32(n *net.IPNet) (uint32, int) {
	ones, _ := n.Mask.Size()
	return uint32(n.IP[0])<<24 | uint32(n.IP[1])<<16 | uint32(n.IP[2])<<8 | uint32(n.IP[3]), ones
}

func addrToKey32(a net.IP) (uint32, int) {
	return uint32(a[0])<<24 | uint32(a[1])<<16 | uint32(a[2])<<8 | uint32(a[3]), 32
}

func (s *SetOfNetworks) addToSetOfNetworks(n *net.IPNet, v interface{}) {
	key, bits := netToKey32(n)
	s.root.Insert(key, bits, v)
}

func (s *SetOfNetworks) GetByNet(n *net.IPNet) interface{} {
	if n.IP.To4() == nil {
		return nil
	}

	key, bits := netToKey32(n)
	node := s.root.Find(key, bits)
	if node != nil {
		return node.Value
	}

	return nil
}

func (s *SetOfNetworks) GetByAddr(a net.IP) interface{} {
	a4 := a.To4()
	if a4 == nil {
		return nil
	}

	key, bits := addrToKey32(a4)
	node := s.root.Find(key, bits)
	if node != nil {
		return node.Value
	}

	return nil
}

func (s *SetOfNetworks) Contains(a net.IP) bool {
	return s.GetByAddr(a) != nil
}

func (s *SetOfNetworks) Iterate() chan NetworkLeafItem {
	ch := make(chan NetworkLeafItem)
	go func() {
		defer close(ch)

		s.root.Do(func(r *bitradix.Radix32, i int) {
			if r.Leaf() && r.Value != nil {
				addr := make(net.IP, 4)
				binary.BigEndian.PutUint32(addr, r.Key())
				ch <- NetworkLeafItem{&net.IPNet{addr, masks[r.Bits()]}, r.Value}
			}
		})
	}()

	return ch
}
