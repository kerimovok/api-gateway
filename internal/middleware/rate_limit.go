package middleware

import (
	"api-gateway/internal/config"
	internalUtils "api-gateway/internal/utils"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/storage/memory"
	"github.com/kerimovok/go-pkg-utils/httpx"
	pkgNet "github.com/kerimovok/go-pkg-utils/net"
)

var (
	store    = memory.New() // Single store instance for all limiters
	limiters = make(map[string]*limiter.Config)
	mu       sync.RWMutex
)

func getLimiter(service string, maxRequests int, duration time.Duration) fiber.Handler {
	mu.RLock()
	if lim, exists := limiters[service]; exists {
		mu.RUnlock()
		return limiter.New(*lim)
	}
	mu.RUnlock()

	mu.Lock()
	defer mu.Unlock()

	cfg := limiter.Config{
		Max:        maxRequests,
		Expiration: duration,
		Storage:    store,
		KeyGenerator: func(c *fiber.Ctx) string {
			return service + "_" + pkgNet.GetUserIP(c) // Use consistent IP detection
		},
		LimitReached: func(c *fiber.Ctx) error {
			response := httpx.TooManyRequests("Rate limit exceeded")
			return httpx.SendResponse(c, response)
		},
	}
	limiters[service] = &cfg
	return limiter.New(cfg)
}

func RateLimitMiddleware() fiber.Handler {
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

		// Check if rate limiting is enabled in service or global config
		var rateLimit *config.RateLimitConfig
		if serviceConfig.RateLimit != nil && serviceConfig.RateLimit.Enabled != nil && *serviceConfig.RateLimit.Enabled {
			rateLimit = serviceConfig.RateLimit
		} else if cfg.Global != nil && cfg.Global.RateLimit != nil &&
			cfg.Global.RateLimit.Enabled != nil && *cfg.Global.RateLimit.Enabled {
			rateLimit = cfg.Global.RateLimit
		}

		// Skip if rate limiting is not enabled or config is nil
		if rateLimit == nil {
			return c.Next()
		}

		// Skip if required fields are not properly set
		if rateLimit.MaxRequests <= 0 || rateLimit.Duration <= 0 {
			return c.Next()
		}

		return getLimiter(serviceName, rateLimit.MaxRequests, rateLimit.Duration)(c)
	}
}
