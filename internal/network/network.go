package network

import (
	"fmt"
	"net"
	"strings"
)

type SubnetInfo struct {
	NetworkAddress string `json:"network_address"`
	Broadcast      string `json:"broadcast"`
	FirstIP        string `json:"first_ip"`
	LastIP         string `json:"last_ip"`
	TotalIPs       uint64 `json:"total_ips"`
	Netmask        string `json:"netmask"`
	CIDR           int    `json:"cidr"`
}

func CalculateSubnet(cidr string) (*SubnetInfo, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}
	ip4 := ip.To4()
	if ip4 == nil {
		return nil, fmt.Errorf("IPv6 is not supported")
	}
	ones, bits := ipnet.Mask.Size()
	if bits != 32 {
		return nil, fmt.Errorf("only IPv4 is supported")
	}
	network := ip4.Mask(ipnet.Mask).To4()
	broadcast := make(net.IP, 4)
	for i := 0; i < 4; i++ {
		broadcast[i] = network[i] | ^ipnet.Mask[i]
	}
	total := uint64(1) << uint(bits-ones)
	first, last := append(net.IP(nil), network...), append(net.IP(nil), broadcast...)
	if ones < 31 {
		first[3]++
		last[3]--
	}
	return &SubnetInfo{
		NetworkAddress: network.String(),
		Broadcast:      broadcast.String(),
		FirstIP:        first.String(),
		LastIP:         last.String(),
		TotalIPs:       total,
		Netmask:        net.IP(ipnet.Mask).String(),
		CIDR:           ones,
	}, nil
}

func LookupOUI(mac string) string {
	clean := strings.ToUpper(strings.NewReplacer("-", ":", ".", ":").Replace(mac))
	parts := strings.Split(clean, ":")
	if len(parts) < 3 {
		return "invalid"
	}
	oui := strings.Join(parts[:3], ":")
	switch oui {
	case "00:1A:2B":
		return "Cisco"
	case "00:0C:29":
		return "VMware"
	case "00:1E:8C":
		return "Apple"
	default:
		return "Unknown"
	}
}
