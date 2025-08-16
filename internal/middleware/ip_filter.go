package middleware

import (
	"fmt"
	"net"
	"sync"

	"api-gateway/internal/config"
	internalUtils "api-gateway/internal/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/kerimovok/go-pkg-utils/httpx"
	pkgNet "github.com/kerimovok/go-pkg-utils/net"
)

var (
	// Cache parsed CIDRs to improve performance
	parsedCIDRs = make(map[string]*net.IPNet)
	cidrMutex   sync.RWMutex
)

// getCIDR returns cached IPNet or parses and caches new one
func getCIDR(cidr string) *net.IPNet {
	cidrMutex.RLock()
	if ipNet, exists := parsedCIDRs[cidr]; exists {
		cidrMutex.RUnlock()
		return ipNet
	}
	cidrMutex.RUnlock()

	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil
	}

	cidrMutex.Lock()
	parsedCIDRs[cidr] = ipNet
	cidrMutex.Unlock()

	return ipNet
}

func IPFilterMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		cfg := config.GetConfig()
		serviceName := c.Params("service")
		serviceConfig, err := internalUtils.GetServiceConfig(serviceName, &cfg)
		if err != nil {
			response := httpx.NotFound("Service not found")
			return httpx.SendResponse(c, response)
		}

		if serviceConfig == nil {
			response := httpx.NotFound("Service not found")
			return httpx.SendResponse(c, response)
		}

		// Get client IP using the utility function
		clientIP := pkgNet.GetUserIP(c)
		ip := net.ParseIP(clientIP)
		if ip == nil {
			response := httpx.BadRequest("Invalid IP address", nil)
			return httpx.SendResponse(c, response)
		}

		// Check blocklist first (both global and service-specific)
		var combinedBlockList []string
		if cfg.Global != nil && len(cfg.Global.IPBlockList) > 0 {
			combinedBlockList = append(combinedBlockList, cfg.Global.IPBlockList...)
		}
		if len(serviceConfig.IPBlockList) > 0 {
			combinedBlockList = append(combinedBlockList, serviceConfig.IPBlockList...)
		}

		if isIPBlocked(ip, combinedBlockList) {
			response := httpx.Forbidden(fmt.Sprintf("IP %s is blocked", clientIP))
			return httpx.SendResponse(c, response)
		}

		// Then check allowlist if defined
		var combinedAllowList []string
		if cfg.Global != nil && len(cfg.Global.IPAllowList) > 0 {
			combinedAllowList = append(combinedAllowList, cfg.Global.IPAllowList...)
		}
		if len(serviceConfig.IPAllowList) > 0 {
			combinedAllowList = append(combinedAllowList, serviceConfig.IPAllowList...)
		}

		if hasAllowlist := len(combinedAllowList) > 0; hasAllowlist {
			if !isIPAllowed(ip, combinedAllowList) {
				response := httpx.Forbidden(fmt.Sprintf("IP %s is not allowed", clientIP))
				return httpx.SendResponse(c, response)
			}
		}

		return c.Next()
	}
}

func isIPAllowed(ip net.IP, allowlist []string) bool {
	if allowlist == nil {
		return false
	}
	return checkIPInList(ip, allowlist)
}

func isIPBlocked(ip net.IP, blocklist []string) bool {
	if blocklist == nil {
		return false
	}
	return checkIPInList(ip, blocklist)
}

func checkIPInList(ip net.IP, ipList []string) bool {
	for _, cidrOrIP := range ipList {
		// Check if it's an exact IP match
		if cidrOrIP == ip.String() {
			return true
		}

		// Check CIDR match
		if ipNet := getCIDR(cidrOrIP); ipNet != nil {
			if ipNet.Contains(ip) {
				return true
			}
		}
	}
	return false
}
