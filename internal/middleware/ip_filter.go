package middleware

import (
	"fmt"
	"net"
	"sync"

	"api-gateway/internal/config"
	internalUtils "api-gateway/internal/utils"
	pkgUtils "api-gateway/pkg/utils"

	"github.com/gofiber/fiber/v2"
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
			return pkgUtils.ErrorResponse(c, fiber.StatusNotFound, "Service not found", err)
		}

		if serviceConfig == nil {
			return pkgUtils.ErrorResponse(c, fiber.StatusNotFound, "Service not found", nil)
		}

		// Get client IP using the utility function
		clientIP := pkgUtils.GetUserIP(c)
		ip := net.ParseIP(clientIP)
		if ip == nil {
			return pkgUtils.ErrorResponse(c, fiber.StatusBadRequest, "Invalid IP address", nil)
		}

		// Check blocklist first (both global and service-specific)
		var combinedBlockList []string
		if cfg.Global != nil {
			combinedBlockList = append(combinedBlockList, cfg.Global.IPBlockList...)
		}
		combinedBlockList = append(combinedBlockList, serviceConfig.IPBlockList...)

		if isIPBlocked(ip, combinedBlockList) {
			return pkgUtils.ErrorResponse(c,
				fiber.StatusForbidden,
				fmt.Sprintf("IP %s is blocked", clientIP),
				nil)
		}

		// Then check allowlist if defined
		var combinedAllowList []string
		if cfg.Global != nil {
			combinedAllowList = append(combinedAllowList, cfg.Global.IPAllowList...)
		}
		combinedAllowList = append(combinedAllowList, serviceConfig.IPAllowList...)

		if hasAllowlist := len(combinedAllowList) > 0; hasAllowlist {
			if !isIPAllowed(ip, combinedAllowList) {
				return pkgUtils.ErrorResponse(c,
					fiber.StatusForbidden,
					fmt.Sprintf("IP %s is not allowed", clientIP),
					nil)
			}
		}

		return c.Next()
	}
}

func isIPAllowed(ip net.IP, allowlist []string) bool {
	return checkIPInList(ip, allowlist)
}

func isIPBlocked(ip net.IP, blocklist []string) bool {
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
