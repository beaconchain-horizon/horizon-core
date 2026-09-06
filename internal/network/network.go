package network

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

type SubnetInfo struct {
	NetworkAddress string
	Broadcast      string
	FirstIP        string
	LastIP         string
	TotalIPs       int
	Netmask        string
	CIDR           int
}

func CalculateSubnet(cidr string) (*SubnetInfo, error) {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}
	ones, bits := ipnet.Mask.Size()
	totalIPs := 1 << (bits - ones)
	ip := ipnet.IP.To4()
	if ip == nil {
		return nil, fmt.Errorf("IPv6 not supported yet")
	}
	network := ip.Mask(ipnet.Mask)
	broadcast := make(net.IP, len(ipnet.IP))
	copy(broadcast, ipnet.IP)
	for i := range broadcast {
		broadcast[i] = broadcast[i] | ^ipnet.Mask[i]
	}
	firstIP := make(net.IP, len(ipnet.IP))
	copy(firstIP, network)
	firstIP[3]++
	lastIP := make(net.IP, len(broadcast))
	copy(lastIP, broadcast)
	lastIP[3]--
	return &SubnetInfo{
		NetworkAddress: network.String(),
		Broadcast:      broadcast.String(),
		FirstIP:        firstIP.String(),
		LastIP:         lastIP.String(),
		TotalIPs:       totalIPs,
		Netmask:        net.IP(ipnet.Mask).String(),
		CIDR:           ones,
	}, nil
}

func LookupOUI(mac string) string {
	mac = strings.ToUpper(strings.ReplaceAll(strings.ReplaceAll(mac, "-", ":"), ".", ":"))
	parts := strings.Split(mac, ":")
	if len(parts) < 3 {
		return "invalid"
	}
	db := map[string]string{
		"00:1A:2B": "Cisco",
		"00:0C:29": "VMware",
		"00:1E:8C": "Apple",
	}
	if v, ok := db[strings.Join(parts[:3], ":")]; ok {
		return v
	}
	return "Unknown"
}
