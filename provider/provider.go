package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/palantir/stacktrace"
	"github.com/sumup-oss/go-pkgs/logger"
	"github.com/sumup-oss/go-pkgs/testutils"
)

func WithLogger(logger logger.Logger) plugin.ProviderFunc {
	return func() *schema.Provider {
		return provider(logger)
	}
}

func provider(logger logger.Logger) *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("NEO4J_USERNAME", ""),
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("NEO4J_PASSWORD", ""),
				Sensitive:   true,
			},
			"realm": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("NEO4J_REALM", ""),
				Sensitive:   false,
			},
			"connection_uri": {
				Type:             schema.TypeString,
				Required:         true,
				DefaultFunc:      schema.EnvDefaultFunc("NEO4J_CONNECTION_URI", ""),
				ValidateDiagFunc: validateNotBlank,
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"neo4j_user": {
				Schema: map[string]*schema.Schema{
					"username": {
						Type:             schema.TypeString,
						Required:         true,
						ValidateDiagFunc: validateNotBlank,
						ForceNew:         true,
					},
					"password": {
						Type:      schema.TypeString,
						Sensitive: true,
						Computed:  true,
						Description: "This is only the initial password, afterwards it's going to be changed by the user. " +
							"Blank password means it'll be auto-generated",
					},
				},
				CreateContext: func(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
					client := meta.(neo4j.Driver)
					sessionConfig := neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}
					session, err := client.NewSession(sessionConfig)
					if err != nil {
						return diag.FromErr(err)
					}
					defer session.Close()

					username := data.Get("username").(string)
					// TODO: Change to requiring a password always that uses `sumup-oss/vaulted` to encrypt it.
					password := testutils.RandString(16)

					err = data.Set("password", password)
					if err != nil {
						return diag.FromErr(err)
					}

					_, err = session.Run(
						"CREATE USER $user IF NOT EXISTS SET PASSWORD $password SET PASSWORD CHANGE REQUIRED",
						map[string]interface{}{
							"user":     username,
							"password": password,
						},
					)
					if err != nil {
						return diag.FromErr(err)
					}

					result, err := session.Run(
						"SHOW USERS WHERE user = $user",
						map[string]interface{}{
							"user": username,
						},
					)
					if err != nil {
						return nil
					}

					record, err := result.Single()
					if err != nil {
						return diag.FromErr(err)
					}
					if record == nil {
						return diag.FromErr(
							stacktrace.NewError(
								"user %s was created, but not actually returned by neo4j. consistency error suspected",
								username,
							),
						)
					}

					data.SetId(username)
					return nil
				},
				ReadContext: func(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
					client := meta.(neo4j.Driver)
					sessionConfig := neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}
					session, err := client.NewSession(sessionConfig)
					if err != nil {
						return diag.FromErr(err)
					}
					defer session.Close()

					username := data.Get("username").(string)
					result, err := session.Run(
						"SHOW USERS WHERE user = $user",
						map[string]interface{}{
							"user": username,
						},
					)
					if err != nil {
						// NOTE: Might be a network error, don't remove the ID from state.
						return diag.FromErr(err)
					}

					record, err := result.Single()
					if err != nil {
						// NOTE: Might be a network error, don't remove the ID from state.
						return diag.FromErr(err)
					}

					if record == nil {
						data.SetId("")
					}

					return nil
				},
				DeleteContext: func(ctx context.Context, data *schema.ResourceData, meta interface{}) diag.Diagnostics {
					client := meta.(neo4j.Driver)
					sessionConfig := neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite}
					session, err := client.NewSession(sessionConfig)
					if err != nil {
						return diag.FromErr(err)
					}
					defer session.Close()

					_, err = session.Run(
						"DROP USER $username",
						map[string]interface{}{
							"username": data.Get("username").(string),
						},
					)
					if err != nil {
						return diag.FromErr(err)
					}

					// NOTE: `SetId` is called automatically if return value is nil
					return nil
				},
			},
		},
		DataSourcesMap: map[string]*schema.Resource{},
		ConfigureContextFunc: func(ctx context.Context, data *schema.ResourceData) (interface{}, diag.Diagnostics) {
			// TODO: Make client options configurable and add connection abstraction that handles the shutdown
			configForNeo4j40 := func(conf *neo4j.Config) {
				conf.UserAgent = "terraform-provider-neo4j"
				conf.Log = neo4j.ConsoleLogger(neo4j.DEBUG)
			}
			client, err := neo4j.NewDriver(
				data.Get("connection_uri").(string),
				neo4j.BasicAuth(
					data.Get("username").(string),
					data.Get("password").(string),
					data.Get("realm").(string),
				),
				configForNeo4j40,
			)
			if err != nil {
				return nil, diag.FromErr(fmt.Errorf("failed to create new driver, err: %s", err))
			}

			return client, diag.Diagnostics{}
		},
	}
}
