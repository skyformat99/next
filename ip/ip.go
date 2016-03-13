package ip

import (
	"encoding/binary"
	"net"
	"reflect"

	"github.com/chzyer/flagly"

	"gopkg.in/logex.v1"
)

type IP [4]byte

func (i IP) String() string {
	return i.IP().String()
}

type IPNet struct {
	IP   IP         // network number
	Mask net.IPMask // network mask
}

func (i *IPNet) ToNet() *net.IPNet {
	return &net.IPNet{
		IP:   i.IP.IP(),
		Mask: i.Mask,
	}
}

func (IPNet) ParseArgs(args []string) (reflect.Value, error) {
	n, err := ParseCIDR(args[0])
	if err != nil {
		return flagly.NilValue, err
	}
	return reflect.ValueOf(n), nil
}
func (IPNet) Type() reflect.Type { return reflect.TypeOf(IPNet{}) }

func (in *IPNet) String() string {
	ipn := &net.IPNet{
		IP:   in.IP.IP(),
		Mask: in.Mask,
	}
	return ipn.String()
}

func ParseCIDR(s string) (*IPNet, error) {
	ip, ipnet, err := net.ParseCIDR(s)
	if err != nil {
		return nil, logex.Trace(err)
	}
	return &IPNet{
		IP:   CopyIP(ip.To4()),
		Mask: ipnet.Mask,
	}, nil
}

func ParseIntIP(ip uint32) IP {
	var ret IP
	binary.BigEndian.PutUint32(ret[:], ip)
	return ret
}

func (ip IP) Clone() IP {
	return ip
}

func (ip IP) Int() uint32 {
	return binary.BigEndian.Uint32(ip[:])
}

func (ip IP) IP() net.IP {
	return net.IPv4(ip[0], ip[1], ip[2], ip[3])
}

func ParseIP(s string) IP {
	return CopyIP(net.ParseIP(s))
}

func CopyIP(ip net.IP) IP {
	ret := [4]byte{}
	copy(ret[:], ip.To4())
	return IP(ret)
}