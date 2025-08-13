package constants

import (
	"github.com/kerimovok/go-pkg-utils/config"
	"github.com/kerimovok/go-pkg-utils/validator"
)

var EnvValidationRules = []validator.ValidationRule{
	// Server validation
	{
		Variable: "PORT",
		Default:  "3000",
		Rule:     config.IsValidPort,
		Message:  "server port is required and must be a valid port number",
	},
	{
		Variable: "GO_ENV",
		Default:  "development",
		Rule:     func(v string) bool { return v == "development" || v == "production" },
		Message:  "GO_ENV must be either 'development' or 'production'",
	},
}
