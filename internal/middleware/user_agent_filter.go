package middleware

import (
	"api-gateway/internal/config"
	internalUtils "api-gateway/internal/utils"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/kerimovok/go-pkg-utils/httpx"
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
			response := httpx.NotFound("Service not found")
			return httpx.SendResponse(c, response)
		}

		if serviceConfig == nil {
			response := httpx.NotFound("Service not found")
			return httpx.SendResponse(c, response)
		}

		userAgent := c.Get("User-Agent")
		if userAgent == "" {
			response := httpx.Forbidden("User-Agent header is required")
			return httpx.SendResponse(c, response)
		}

		normalizedUA := getNormalizedUserAgent(userAgent)

		// Check service-specific blocklist first (more specific rules take precedence)
		if isUserAgentBlocked(normalizedUA, serviceConfig.UserAgentBlocklist) {
			response := httpx.Forbidden("User-Agent is blocked for this service")
			return httpx.SendResponse(c, response)
		}

		// Then check global blocklist
		if cfg.Global != nil && isUserAgentBlocked(normalizedUA, cfg.Global.UserAgentBlocklist) {
			response := httpx.Forbidden("User-Agent is blocked globally")
			return httpx.SendResponse(c, response)
		}

		// Check service-specific allowlist if defined
		if len(serviceConfig.UserAgentAllowlist) > 0 {
			if !isUserAgentAllowed(normalizedUA, serviceConfig.UserAgentAllowlist) {
				response := httpx.Forbidden("User-Agent is not allowed for this service")
				return httpx.SendResponse(c, response)
			}
			return c.Next() // If allowed by service rules, skip global check
		}

		// Check global allowlist if defined
		if cfg.Global != nil && len(cfg.Global.UserAgentAllowlist) > 0 {
			if !isUserAgentAllowed(normalizedUA, cfg.Global.UserAgentAllowlist) {
				response := httpx.Forbidden("User-Agent is not allowed globally")
				return httpx.SendResponse(c, response)
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
