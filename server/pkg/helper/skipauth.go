package helper

import (
	"log"
	"net"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	subnets     []*net.IPNet
	subnetsOnce sync.Once
)

func SkipAuth(c *gin.Context) bool {
	remote_ip := net.ParseIP(c.RemoteIP())
	if remote_ip == nil {
		return false
	}

	// get ipnets of local interface
	subnetsOnce.Do(func() {
		interfaces, err := net.Interfaces()
		if err != nil {
			log.Printf("[Error] list interface failed: %v", err)
		}
		for _, iface := range interfaces {
			addrs, err := iface.Addrs()
			if err != nil {
				log.Printf("[Error] list addrs of interface %s failed: %v", iface.Name, err)
				continue
			}
			for _, addr := range addrs {
				ip, subnet, err := net.ParseCIDR(addr.String())
				if err != nil {
					log.Printf("[Error] ParseCIDR %s failed: %v", addr.String(), err)
					continue
				}

				if ip.IsLoopback() {
					log.Printf("[Debug] skip loopback addr %s", ip.String())
					continue // Skip loopback addresses
				}

				subnets = append(subnets, subnet)
			}
		}
	})

	// if remote is in local subnet, skip auth
	for _, subnet := range subnets {
		if subnet.Contains(remote_ip) {
			log.Printf("[Info] remote_ip %s is in subnet, skip authRequirement", c.RemoteIP())
			return true
		}
	}
	return false
}
