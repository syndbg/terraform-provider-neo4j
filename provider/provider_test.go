package provider

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sumup-oss/go-pkgs/logger"
)

func TestWithLogger(t *testing.T) {
	t.Parallel()

	loggerInstance := logger.NewLogrusLogger()
	actual := WithLogger(loggerInstance)()
	require.NotNil(t, actual)
}

func TestProvider(t *testing.T) {
	t.Run(
		"it is a valid terraform provider",
		func(t *testing.T) {
			t.Parallel()

			loggerInstance := logger.NewLogrusLogger()
			p := provider(loggerInstance)

			actualErr := p.InternalValidate()
			fmt.Println(actualErr)
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
