package middleware

import (
	"api-gateway/internal/config"
	internalUtils "api-gateway/internal/utils"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/kerimovok/go-pkg-utils/httpx"
)

// APIKeyMiddleware validates the API key for a service
func APIKeyMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		cfg := config.GetConfig()
		serviceName := c.Params("service")

		serviceConfig, err := internalUtils.GetServiceConfig(serviceName, &cfg)
		if err != nil {
			response := httpx.NotFound("Service not found")
			return httpx.SendResponse(c, response)
		}

		// Skip if auth config is nil or not enabled
		if serviceConfig.Auth == nil || serviceConfig.Auth.Enabled == nil || !*serviceConfig.Auth.Enabled {
			return c.Next()
		}

		// Validate required auth fields
		if serviceConfig.Auth.Key == "" || serviceConfig.Auth.Value == "" {
			response := httpx.InternalServerError("Invalid auth configuration", fmt.Errorf("missing auth key or value"))
			return httpx.SendResponse(c, response)
		}

		providedAPIKey := c.Get(serviceConfig.Auth.Key)
		if providedAPIKey == "" {
			response := httpx.Unauthorized("API key is missing")
			return httpx.SendResponse(c, response)
		}

		if providedAPIKey != serviceConfig.Auth.Value {
			response := httpx.Forbidden("Invalid API key")
			return httpx.SendResponse(c, response)
		}

		return c.Next()
	}
}
