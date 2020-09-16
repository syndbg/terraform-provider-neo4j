package main

import (
	"github.com/sumup-oss/go-pkgs/logger"

	"github.com/syndbg/terraform-provider-neo4j/provider"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	loggerInstance := logger.NewLogrusLogger()
	loggerInstance.SetLevel(logger.InfoLevel)

	plugin.Serve(
		&plugin.ServeOpts{
			ProviderFunc: provider.WithLogger(loggerInstance),
		},
	)
}
