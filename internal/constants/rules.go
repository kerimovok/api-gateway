package constants

import "api-gateway/pkg/utils"

var EnvValidationRules = []utils.ValidationRule{
	// Server validation
	{
		Variable: "PORT",
		Default:  "3000",
		Rule:     utils.IsValidPort,
		Message:  "server port is required and must be a valid port number",
	},
	{
		Variable: "GO_ENV",
		Default:  "development",
		Rule:     func(v string) bool { return v == "development" || v == "production" },
		Message:  "GO_ENV must be either 'development' or 'production'",
	},
}
