package gofiberswagger

import (
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
)

func TestSwaggerConfigDefault(t *testing.T) {
	t.Parallel()

	t.Run("empty config", func(t *testing.T) {
		t.Parallel()
		cfg := swaggerConfigDefault(SwaggerConfig{})
		assert.Equal(t, DefaultSwaggerConfig.Info.Title, cfg.Info.Title)
		assert.Equal(t, DefaultSwaggerConfig.Info.Version, cfg.Info.Version)
		assert.NotNil(t, cfg.Paths)
		assert.NotNil(t, cfg.Components)
		assert.NotNil(t, cfg.Components.Schemas)
	})

	t.Run("partial config", func(t *testing.T) {
		t.Parallel()
		cfg := swaggerConfigDefault(SwaggerConfig{
			Info: &openapi3.Info{
				Title: "My API",
			},
		})
		assert.Equal(t, "My API", cfg.Info.Title)
		assert.Equal(t, DefaultSwaggerConfig.Info.Version, cfg.Info.Version)
	})
}
