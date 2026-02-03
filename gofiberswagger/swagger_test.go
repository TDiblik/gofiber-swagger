package gofiberswagger

import (
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
)

func TestGenerateIndexPage(t *testing.T) {
	t.Parallel()

	// case 1: simple execution
	cfg := swaggerUIConfigDefault(SwaggerUIConfig{})
	index, err := generateIndexPage(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, index)
	assert.Contains(t, string(index), "SwaggerUIBundle")

	// case 2: verify that values are being changed
	cfg = swaggerUIConfigDefault(SwaggerUIConfig{
		Title:       "Test",
		DeepLinking: true,
	})
	index, err = generateIndexPage(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, index)
	assert.Contains(t, string(index), "<title>Test</title>")
	assert.Contains(t, string(index), "\"deepLinking\":true")
}

func TestGenerateOpenApiSchema(t *testing.T) {
	t.Parallel()
	schema := openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:       "Title",
			Description: "This is the description",
			Version:     "v1.0",
		},
		Servers: []*openapi3.Server{
			{
				URL: "http://example.com",
			},
		},
	}

	schema.Tags = append(schema.Tags, &openapi3.Tag{
		Name:        "Test Tag",
		Description: "This is the description of the test tag",
		ExternalDocs: &openapi3.ExternalDocs{
			Description: "Find out more about our store",
			URL:         "http://example.com",
		},
	})
	schema.AddServer(&openapi3.Server{
		URL:         "http://localhost:8080",
		Description: "Development server",
	})

	as_json, as_yaml, err := generateOpenApiSchema(schema)
	assert.NoError(t, err, "Error while generating the swagger schema")
	assert.NotNil(t, as_json, "The generated json schema is nil")
	assert.NotNil(t, as_yaml, "The generated yaml schema is nil")

	assert.Contains(t, string(as_json), "Title")
	assert.Contains(t, string(as_json), "Test Tag")
	assert.Contains(t, string(as_json), "http://localhost:8080")
	assert.Contains(t, string(as_json), "http://example.com")

	assert.Contains(t, string(as_yaml), "Title")
	assert.Contains(t, string(as_yaml), "Test Tag")
	assert.Contains(t, string(as_yaml), "http://localhost:8080")
	assert.Contains(t, string(as_yaml), "http://example.com")
}

func TestCreateSwaggerFiles(t *testing.T) {
	t.Parallel()

	temp_dir := t.TempDir()
	index_page := []byte("index")
	schema_as_json := []byte("json")
	schema_as_yaml := []byte("yaml")

	err := createSwaggerFiles(temp_dir, index_page, schema_as_json, schema_as_yaml)
	assert.NoError(t, err, "Error while creating the swagger files")

	assert.FileExists(t, filepath.Join(temp_dir, "index.html"))
	assert.FileExists(t, filepath.Join(temp_dir, "swagger.json"))
	assert.FileExists(t, filepath.Join(temp_dir, "swagger.yaml"))

	index_content, err := os.ReadFile(filepath.Join(temp_dir, "index.html"))
	assert.NoError(t, err)
	assert.Equal(t, index_page, index_content)

	json_content, err := os.ReadFile(filepath.Join(temp_dir, "swagger.json"))
	assert.NoError(t, err)
	assert.Equal(t, schema_as_json, json_content)

	yaml_content, err := os.ReadFile(filepath.Join(temp_dir, "swagger.yaml"))
	assert.NoError(t, err)
	assert.Equal(t, schema_as_yaml, yaml_content)
}

func TestRegister(t *testing.T) {
	t.Parallel()

	t.Run("should serve swagger files on /swagger", func(t *testing.T) {
		t.Parallel()
		// fiber instance
		app := fiber.New()

		// register swagger
		err := Register(app, Config{})
		assert.NoError(t, err, "Error while registering swagger")

		// test swagger routes
		resp, err := app.Test(httptest.NewRequest("GET", "/swagger", nil))
		assert.NoError(t, err, "Error while testing the swagger route")
		assert.Equal(t, 200, resp.StatusCode, "The swagger route should return 200")
	})

	t.Run("should create swagger files if enabled", func(t *testing.T) {
		t.Parallel()

		// setup
		app := fiber.New()
		tempDir := t.TempDir()

		// register swagger with file creation enabled
		err := Register(app, Config{
			CreateSwaggerFiles: true,
			SwaggerFilesPath:   tempDir,
		})
		assert.NoError(t, err, "Error while registering swagger")

		// assertions
		assert.FileExists(t, filepath.Join(tempDir, "index.html"))
		assert.FileExists(t, filepath.Join(tempDir, "swagger.json"))
		assert.FileExists(t, filepath.Join(tempDir, "swagger.yaml"))
	})

	t.Run("should add method to tags if enabled", func(t *testing.T) {
		t.Parallel()

		// setup
		app := fiber.New()
		app.Get("/test", func(c fiber.Ctx) error {
			return c.SendString("test")
		})

		// register swagger with method to tags enabled
		err := Register(app, Config{
			AppendMethodToTags: true,
		})
		assert.NoError(t, err, "Error while registering swagger")
	})

	t.Run("should require auth if enabled", func(t *testing.T) {
		t.Parallel()

		// setup
		app := fiber.New()
		app.Get("/test", func(c fiber.Ctx) error {
			return c.SendString("test")
		})

		// register swagger with auth required
		err := Register(app, Config{
			AutomaticallyRequireAuth: true,
			RequiredAuth: &openapi3.SecurityRequirements{
				{
					"bearerAuth": []string{},
				},
			},
		})
		assert.NoError(t, err, "Error while registering swagger")
	})

	t.Run("should handle path parameters", func(t *testing.T) {
		t.Parallel()

		// setup
		app := fiber.New()
		app.Get("/test/:id", func(c fiber.Ctx) error {
			return c.SendString(fmt.Sprintf("test %s", c.Params("id")))
		})

		// register swagger
		err := Register(app, Config{})
		assert.NoError(t, err, "Error while registering swagger")
	})

	t.Run("should handle wildcard path parameters", func(t *testing.T) {
		t.Parallel()

		// setup
		app := fiber.New()
		app.Get("/test/*", func(c fiber.Ctx) error {
			return c.SendString(fmt.Sprintf("test %s", c.Params("*")))
		})

		// register swagger
		err := Register(app, Config{})
		assert.NoError(t, err, "Error while registering swagger")
	})
}
