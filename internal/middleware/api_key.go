package middleware

import (
	"api-gateway/internal/config"
	internalUtils "api-gateway/internal/utils"
	pkgUtils "api-gateway/pkg/utils"

	"github.com/gofiber/fiber/v2"
)

// APIKeyMiddleware validates the API key for a service
func APIKeyMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		cfg := config.GetConfig()
		serviceName := c.Params("service")

		serviceConfig, err := internalUtils.GetServiceConfig(serviceName, &cfg)
		if err != nil {
			return pkgUtils.ErrorResponse(c, fiber.StatusNotFound, "Service not found", err)
		}

		// Skip if auth is not enabled
		if serviceConfig.Auth.Enabled == nil || !*serviceConfig.Auth.Enabled {
			return c.Next()
		}

		providedAPIKey := c.Get(serviceConfig.Auth.Key)
		if providedAPIKey == "" {
			return pkgUtils.ErrorResponse(c, fiber.StatusUnauthorized, "API key is missing", nil)
		}

		if providedAPIKey != serviceConfig.Auth.Value {
			return pkgUtils.ErrorResponse(c, fiber.StatusForbidden, "Invalid API key", nil)
		}

		return c.Next()
	}
}
