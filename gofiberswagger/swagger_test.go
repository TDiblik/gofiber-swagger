package gofiberswagger

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterSwaggerPath(t *testing.T) {
	for _, tc := range []struct {
		name       string
		path       string
		base       string
		uiURL      string
		wantURL    string
		useDefault bool
	}{
		{name: "empty", base: "/swagger", wantURL: "/swagger/swagger.yaml"},
		{name: "explicit default", path: "/swagger", base: "/swagger", wantURL: "/swagger/swagger.yaml"},
		{name: "custom", path: "/docs", base: "/docs", wantURL: "/docs/swagger.yaml"},
		{name: "nested", path: "/api/v1/docs", base: "/api/v1/docs", wantURL: "/api/v1/docs/swagger.yaml"},
		{name: "trailing slash", path: "/docs/", base: "/docs", wantURL: "/docs/swagger.yaml"},
		{name: "no leading slash", path: "docs", base: "/docs", wantURL: "/docs/swagger.yaml"},
		{name: "root", path: "/", base: "", wantURL: "/swagger.yaml"},
		{name: "copied default config", path: "/docs", base: "/docs", wantURL: "/docs/swagger.yaml", useDefault: true},
		{name: "custom schema URL", path: "/docs", base: "/docs", uiURL: "https://example.com/openapi.json", wantURL: "https://example.com/openapi.json"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			config := Config{}
			if tc.useDefault {
				config = DefaultConfig
			}
			config.CreateSwaggerFiles = false
			config.SwaggerPath = tc.path
			if tc.uiURL != "" {
				config.SwaggerUI.URL = tc.uiURL
			}
			config.Swagger = SwaggerConfig{
				OpenAPI: "3.1.1",
				Info:    &Info{Title: "Path test", Version: "1.0"},
				Paths:   openapi3.NewPaths(), Components: &Components{},
			}
			require.NoError(t, Register(app, config))

			for _, suffix := range []string{"", "/", "/index.html", "/swagger", "/swagger.json", "/swagger.yaml"} {
				url := tc.base + suffix
				if url == "" {
					url = "/"
				}
				resp, err := app.Test(httptest.NewRequest("GET", url, nil))
				require.NoError(t, err)
				body, err := io.ReadAll(resp.Body)
				resp.Body.Close()
				require.NoError(t, err)
				assert.Equal(t, 200, resp.StatusCode, url)
				if strings.HasSuffix(suffix, ".json") {
					assert.Contains(t, resp.Header.Get("Content-Type"), "application/json")
					var schema SwaggerConfig
					require.NoError(t, json.Unmarshal(body, &schema))
					require.NotNil(t, schema.Info)
					assert.Equal(t, "Path test", schema.Info.Title)
				} else if strings.HasSuffix(suffix, ".yaml") {
					assert.Contains(t, string(body), "title: Path test")
				} else {
					assert.Contains(t, resp.Header.Get("Content-Type"), "text/html")
					assert.Contains(t, string(body), `"url":"`+tc.wantURL+`"`)
				}
			}
			if tc.base != "/swagger" && tc.base != "" {
				resp, err := app.Test(httptest.NewRequest("GET", "/swagger", nil))
				require.NoError(t, err)
				resp.Body.Close()
				assert.Equal(t, 404, resp.StatusCode)
			}
		})
	}
}

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

	t.Run("should not create duplicate path parameters if custom path parameter is provided", func(t *testing.T) {
		t.Parallel()

		// setup
		app := fiber.New()

		RegisterRoute("GET", "/custom-parameter/:id", &RouteInfo{
			Parameters: NewParameters(
				NewPathParameterExtended("id", &Schema{
					Type:   &Types{"integer"},
					Format: "int64",
				}),
			),
		})

		app.Get("/custom-parameter/:id", func(c fiber.Ctx) error {
			return c.SendString(fmt.Sprintf("test %s", c.Params("id")))
		})

		// register swagger
		config := Config{Swagger: SwaggerConfig{Paths: openapi3.NewPaths(), Components: &Components{}}}
		config.Swagger = swaggerConfigDefault(config.Swagger)
		err := Register(app, config)
		assert.NoError(t, err, "Error while registering swagger")

		operation := config.Swagger.Paths.Find("/custom-parameter/{id}").Get
		assert.NotNil(t, operation, "Operation should not be nil")

		paramCount := 0
		for _, param := range operation.Parameters {
			if param.Value != nil && param.Value.In == openapi3.ParameterInPath && param.Value.Name == "id" {
				paramCount++
				assert.Equal(t, "integer", param.Value.Schema.Value.Type.Slice()[0])
				assert.Equal(t, "int64", param.Value.Schema.Value.Format)
			}
		}

		assert.Equal(t, 1, paramCount, "There should be exactly one path parameter named id")
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
