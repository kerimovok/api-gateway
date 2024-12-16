package middleware

import (
	"api-gateway/internal/config"
	internalUtils "api-gateway/internal/utils"
	pkgUtils "api-gateway/pkg/utils"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
)

var (
	// Cache normalized user agents for better performance
	userAgentCache = make(map[string]string)
	uaCacheMutex   sync.RWMutex
)

// getNormalizedUserAgent returns cached or normalized user agent
func getNormalizedUserAgent(ua string) string {
	normalized := strings.ToLower(ua)

	uaCacheMutex.RLock()
	if cached, exists := userAgentCache[normalized]; exists {
		uaCacheMutex.RUnlock()
		return cached
	}
	uaCacheMutex.RUnlock()

	uaCacheMutex.Lock()
	userAgentCache[normalized] = normalized
	uaCacheMutex.Unlock()

	return normalized
}

func UserAgentFilter() fiber.Handler {
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

		userAgent := c.Get("User-Agent")
		if userAgent == "" {
			return pkgUtils.ErrorResponse(c, fiber.StatusForbidden, "User-Agent header is required", nil)
		}

		normalizedUA := getNormalizedUserAgent(userAgent)

		// Check service-specific blocklist first (more specific rules take precedence)
		if isUserAgentBlocked(normalizedUA, serviceConfig.UserAgentBlocklist) {
			return pkgUtils.ErrorResponse(c,
				fiber.StatusForbidden,
				"User-Agent is blocked for this service",
				nil)
		}

		// Then check global blocklist
		if cfg.Global != nil && isUserAgentBlocked(normalizedUA, cfg.Global.UserAgentBlocklist) {
			return pkgUtils.ErrorResponse(c,
				fiber.StatusForbidden,
				"User-Agent is blocked globally",
				nil)
		}

		// Check service-specific allowlist if defined
		if len(serviceConfig.UserAgentAllowlist) > 0 {
			if !isUserAgentAllowed(normalizedUA, serviceConfig.UserAgentAllowlist) {
				return pkgUtils.ErrorResponse(c,
					fiber.StatusForbidden,
					"User-Agent is not allowed for this service",
					nil)
			}
			return c.Next() // If allowed by service rules, skip global check
		}

		// Check global allowlist if defined
		if cfg.Global != nil && len(cfg.Global.UserAgentAllowlist) > 0 {
			if !isUserAgentAllowed(normalizedUA, cfg.Global.UserAgentAllowlist) {
				return pkgUtils.ErrorResponse(c,
					fiber.StatusForbidden,
					"User-Agent is not allowed globally",
					nil)
			}
		}

		return c.Next()
	}
}

func isUserAgentBlocked(normalizedUA string, blocklist []string) bool {
	for _, blocked := range blocklist {
		if strings.Contains(normalizedUA, strings.ToLower(blocked)) {
			return true
		}
	}
	return false
}

func isUserAgentAllowed(normalizedUA string, allowlist []string) bool {
	for _, allowed := range allowlist {
		if strings.Contains(normalizedUA, strings.ToLower(allowed)) {
			return true
		}
	}
	return false
}
