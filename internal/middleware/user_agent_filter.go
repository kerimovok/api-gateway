package middleware

import (
	"api-gateway/internal/config"
	internalUtils "api-gateway/internal/utils"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kerimovok/go-pkg-utils/httpx"
)

var (
	// Cache normalized user agents for better performance
	userAgentCache  = make(map[string]string)
	uaCacheMutex    sync.RWMutex
	lastCleanup     time.Time
	cleanupInterval = 5 * time.Minute
	maxCacheSize    = 1000
)

// getNormalizedUserAgent returns cached or normalized user agent
func getNormalizedUserAgent(ua string) string {
	// Check cache first
	uaCacheMutex.RLock()
	if cached, exists := userAgentCache[ua]; exists {
		uaCacheMutex.RUnlock()
		return cached
	}
	uaCacheMutex.RUnlock()

	// Normalize and cache
	normalized := strings.ToLower(ua)

	uaCacheMutex.Lock()
	defer uaCacheMutex.Unlock()

	// Cleanup cache if needed
	cleanupCacheIfNeeded()

	// Store in cache
	userAgentCache[ua] = normalized
	return normalized
}

// cleanupCacheIfNeeded removes old entries if cache is too large
func cleanupCacheIfNeeded() {
	now := time.Now()
	if now.Sub(lastCleanup) < cleanupInterval && len(userAgentCache) < maxCacheSize {
		return
	}

	// Clear cache if it's too large
	if len(userAgentCache) >= maxCacheSize {
		userAgentCache = make(map[string]string)
	}
	lastCleanup = now
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
		if len(serviceConfig.UserAgentBlocklist) > 0 && isUserAgentBlocked(normalizedUA, serviceConfig.UserAgentBlocklist) {
			response := httpx.Forbidden("User-Agent is blocked for this service")
			return httpx.SendResponse(c, response)
		}

		// Then check global blocklist
		if cfg.Global != nil && len(cfg.Global.UserAgentBlocklist) > 0 && isUserAgentBlocked(normalizedUA, cfg.Global.UserAgentBlocklist) {
			response := httpx.Forbidden("User-Agent is blocked globally")
			return httpx.SendResponse(c, response)
		}

		// Check service-specific allowlist if defined
		if len(serviceConfig.UserAgentAllowlist) > 0 {
			if !isUserAgentAllowed(normalizedUA, serviceConfig.UserAgentAllowlist) {
				response := httpx.Forbidden("User-Agent is not allowed for this service")
				return httpx.SendResponse(c, response)
			}
			// If allowed by service rules, proceed to next middleware
			return c.Next()
		}

		// Check global allowlist if defined
		if cfg.Global != nil && len(cfg.Global.UserAgentAllowlist) > 0 {
			if !isUserAgentAllowed(normalizedUA, cfg.Global.UserAgentAllowlist) {
				response := httpx.Forbidden("User-Agent is not allowed globally")
				return httpx.SendResponse(c, response)
			}
		}

		// If no allowlists are defined, allow the request to proceed
		return c.Next()
	}
}

func isUserAgentBlocked(normalizedUA string, blocklist []string) bool {
	if blocklist == nil {
		return false
	}
	for _, blocked := range blocklist {
		// Convert blocklist item to lowercase once
		normalizedBlocked := strings.ToLower(blocked)
		if strings.Contains(normalizedUA, normalizedBlocked) {
			return true
		}
	}
	return false
}

func isUserAgentAllowed(normalizedUA string, allowlist []string) bool {
	if allowlist == nil {
		return false
	}
	for _, allowed := range allowlist {
		// Convert allowlist item to lowercase once
		normalizedAllowed := strings.ToLower(allowed)
		if strings.Contains(normalizedUA, normalizedAllowed) {
			return true
		}
	}
	return false
}
