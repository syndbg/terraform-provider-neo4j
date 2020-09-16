package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sumup-oss/go-pkgs/logger"
	"github.com/sumup-oss/go-pkgs/testutils"
)

// TODO: Split to multiple files
func TestProvider_IntegrationTest(t *testing.T) {
	t.Run(
		"with a neo4j instance with basic auth",
		func(t *testing.T) {

			t.Run(
				"it applies and destroyed successfully `neo4j_user` resource",
				func(t *testing.T) {
					loggerInstance := logger.NewLogrusLogger()

					username := testutils.RandString(12)
					// NOTE: Don't enforce the `TF_ACC` environment variable requirement,
					// but still run it as an integration test.
					resource.UnitTest(
						t,
						resource.TestCase{
							ProviderFactories: map[string]func() (*schema.Provider, error){
								"neo4j": func() (*schema.Provider, error) {
									return provider(loggerInstance), nil
								},
							},
							Steps: []resource.TestStep{
								{
									Config: fmt.Sprintf(`
resource "neo4j_user" "user" {
  username = "%s"
}
`, username),
									Check: resource.ComposeTestCheckFunc(
										resource.TestCheckResourceAttr(
											"neo4j_user.user",
											"username",
											username,
										),
									),
								},
							},
						},
					)
				},
			)
			t.Parallel()

			loggerInstance := logger.NewLogrusLogger()
			p := provider(loggerInstance)

			actualErr := p.InternalValidate()
			require.Nil(t, actualErr)
		},
	)

	t.Run(
		"it has the `neo4j_user` resource",
		func(t *testing.T) {
			t.Parallel()

			loggerInstance := logger.NewLogrusLogger()
			p := provider(loggerInstance)

			actual := p.Resources()
			assert.Len(t, actual, 1)
			assert.Equal(t, "neo4j_user", actual[0].Name)
			assert.True(t, actual[0].SchemaAvailable)
			// TODO: Make it importable
			assert.False(t, actual[0].Importable)
		},
	)
}
