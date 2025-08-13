package handlers

import (
	"api-gateway/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/proxy"
	"github.com/kerimovok/go-pkg-utils/httpx"
)

// ProxyHandler forwards requests to the upstream service
func ProxyHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		cfg := config.GetConfig()
		serviceName := c.Params("service")

		// Find the corresponding service configuration
		service, exists := cfg.Services[serviceName]
		if !exists {
			response := httpx.NotFound("Service not found")
			return httpx.SendResponse(c, response)
		}

		// Forward the request to the upstream URL
		targetURL := service.URL + "/" + c.Params("*")
		if err := proxy.Forward(targetURL)(c); err != nil {
			response := httpx.BadGateway("Failed to proxy request")
			return httpx.SendResponse(c, response)
		}

		return nil
	}
}
