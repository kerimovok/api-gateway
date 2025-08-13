package middleware

import (
	"api-gateway/internal/config"
	"api-gateway/internal/utils"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/storage/memory"
	"github.com/kerimovok/go-pkg-utils/httpx"
)

var (
	// Create a single store instance with default settings
	cacheStore = memory.New(memory.Config{
		GCInterval: 10 * time.Second,
	})
)

// CacheMiddleware applies caching based on service configuration
func CacheMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		cfg := config.GetConfig()
		serviceName := c.Params("service")
		serviceConfig, err := utils.GetServiceConfig(serviceName, &cfg)
		if err != nil {
			response := httpx.NotFound("Service not found")
			return httpx.SendResponse(c, response)
		}

		if serviceConfig == nil {
			return c.Next()
		}

		// Determine cache config to use (service-specific or global)
		var cacheConfig *config.CacheConfig
		if serviceConfig.Cache != nil {
			cacheConfig = serviceConfig.Cache
		} else if cfg.Global != nil && cfg.Global.Cache != nil {
			cacheConfig = cfg.Global.Cache
		}

		// Skip caching if config is nil or explicitly disabled
		if cacheConfig == nil || cacheConfig.Enabled == nil || !*cacheConfig.Enabled {
			return c.Next()
		}

		return cache.New(cache.Config{
			Next: func(c *fiber.Ctx) bool {
				return c.Method() != "GET" // Only cache GET requests
			},
			Expiration: cacheConfig.Duration,
			Storage:    cacheStore,
			KeyGenerator: func(c *fiber.Ctx) string {
				return serviceName + "_" + c.Path() + string(c.OriginalURL())
			},
		})(c)
	}
}
