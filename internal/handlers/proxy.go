package handlers

import (
	"api-gateway/internal/config"
	"api-gateway/pkg/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
)

// ProxyHandler forwards requests to the upstream service
func ProxyHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		cfg := config.GetConfig()
		serviceName := c.Params("service")

		// Find the corresponding service configuration
		service, exists := cfg.Services[serviceName]
		if !exists {
			return utils.ErrorResponse(c, fiber.StatusNotFound, "Service not found", nil)
		}

		// Forward the request to the upstream URL
		targetURL := service.URL + "/" + c.Params("*")
		if err := proxy.Forward(targetURL)(c); err != nil {
			return utils.ErrorResponse(c, fiber.StatusBadGateway, "Failed to proxy request", err)
		}

		return nil
	}
}
