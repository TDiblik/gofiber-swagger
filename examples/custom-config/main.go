package main

import (
	"html/template"
	"log"

	"github.com/TDiblik/gofiber-swagger/gofiberswagger"
	"github.com/gofiber/fiber/v3"
)

func main() {
	app := fiber.New()

	// Create wrapper around the fiber router
	router := gofiberswagger.NewRouter(app)
	router.Get("/", nil, HelloHandler)

	// Provide the custom options you wanna.
	// You can go all out and basically use ANY openapi property!
	gofiberswagger.Register(app, gofiberswagger.Config{
		SwaggerPath: "/docs",
		Swagger: gofiberswagger.SwaggerConfig{
			OpenAPI: "3.0.0",
			Info: &gofiberswagger.Info{
				Title:   "Title inside the generated files.",
				Version: "0.0.1",
			},
		},
		SwaggerUI: gofiberswagger.SwaggerUIConfig{
			// URL defaults to /docs/swagger.yaml based on SwaggerPath.
			Title:  "Swagger UI - title of the swagger UI page",
			Layout: "StandaloneLayout",
			Plugins: []template.JS{
				template.JS("SwaggerUIBundle.plugins.DownloadUrl"),
			},
			Presets: []template.JS{
				template.JS("SwaggerUIBundle.presets.apis"),
				template.JS("SwaggerUIStandalonePreset"),
			},
			DeepLinking:              true,
			DefaultModelsExpandDepth: 1,
			DefaultModelExpandDepth:  1,
			DefaultModelRendering:    "example",
			DocExpansion:             "list",
			SyntaxHighlight: &gofiberswagger.SyntaxHighlightConfig{
				Activate: true,
				Theme:    "agate",
			},
			ShowMutatedRequest:     true,
			DisplayRequestDuration: true,
			PersistAuthorization:   true,
		},
		CreateSwaggerFiles:       true,
		SwaggerFilesPath:         "./generated/swagger",
		AppendMethodToTags:       false,
		FilterOutAppUse:          true,
		RequiredAuth:             nil,
		AutomaticallyRequireAuth: false,
	})

	// You can now see your:
	// - UI at /docs/
	// - json at /docs/swagger.json
	// - yaml at /docs/swagger.yaml

	log.Fatal(app.Listen(":3000"))
}

// ----- Hello Handler and it's types ----- //
func HelloHandler(c fiber.Ctx) error {
	return c.SendStatus(200)
}
