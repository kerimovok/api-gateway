package main

import (
	"api-gateway/internal/config"
	"api-gateway/internal/constants"
	"api-gateway/internal/handlers"
	"api-gateway/internal/middleware"
	pkgUtils "api-gateway/pkg/utils"
	"api-gateway/pkg/validator"
	"net/http"

	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/google/uuid"
)

func init() {
	// Initialize configuration
	if err := config.InitConfig(); err != nil {
		pkgUtils.LogFatal("failed to initialize config", err)
	}

	// Validate environment variables
	if err := pkgUtils.ValidateConfig(constants.EnvValidationRules); err != nil {
		pkgUtils.LogFatal("configuration validation failed", err)
	}

	// Initialize validator
	validator.InitValidator()
}

func setupApp() *fiber.App {
	app := fiber.New(fiber.Config{})

	// Middleware
	app.Use(helmet.New())
	app.Use(cors.New())
	app.Use(compress.New())
	app.Use(healthcheck.New())
	app.Use(requestid.New(requestid.Config{
		Generator: func() string {
			return uuid.New().String()
		},
	}))

	// Enable logging middleware based on global configuration
	app.Use(func(c *fiber.Ctx) error {
		cfg := config.GetConfig()

		// Only enable logging if global logging is enabled
		if cfg.Global != nil && cfg.Global.Logging != nil && *cfg.Global.Logging {
			return logger.New()(c)
		}

		return c.Next()
	})

	return app
}

func main() {
	app := setupApp()

	// Create channel for shutdown signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Set up a dynamic route to proxy requests
	app.All("/:service/*",
		middleware.IPFilterMiddleware(),
		middleware.UserAgentFilter(),
		middleware.APIKeyMiddleware(),
		middleware.RateLimitMiddleware(),
		middleware.CacheMiddleware(),
		handlers.ProxyHandler())

	// Start server in a goroutine
	go func() {
		if err := app.Listen(":" + pkgUtils.GetEnv("PORT")); err != nil && err != http.ErrServerClosed {
			pkgUtils.LogFatal("failed to start server", err)
		}
	}()

	// Wait for shutdown signal
	<-quit
	pkgUtils.LogInfo("Shutting down server...")

	// Stop the config watcher
	config.StopConfig()

	// Gracefully shutdown the server
	if err := app.Shutdown(); err != nil {
		pkgUtils.LogError("error shutting down server", err)
	}
}
