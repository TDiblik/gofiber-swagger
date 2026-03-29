package main

import (
	"log"

	"github.com/TDiblik/gofiber-swagger/gofiberswagger"
	"github.com/gofiber/fiber/v3"
)

func main() {
	app := fiber.New()

	router := gofiberswagger.NewRouter(app)

	// these do not require auth
	router.Get("/", nil, HelloHandler)

	// Option A: specify Security field with your SecurityRequirements for each route
	router.Get("/behind-auth", &gofiberswagger.RouteInfo{
		Security: &gofiberswagger.SecurityRequirements{{
			"User API Token": {},
		}},
	}, HelloHandler)

	// You can now see your:
	// - UI at /swagger/
	// - json at /swagger/swagger.json
	// - yaml at /swagger/swagger.yaml
	gofiberswagger.Register(app, gofiberswagger.Config{
		Swagger: gofiberswagger.SwaggerConfig{
			OpenAPI: gofiberswagger.DefaultSwaggerConfig.OpenAPI,
			Info:    gofiberswagger.DefaultSwaggerConfig.Info,
			Components: &gofiberswagger.Components{
				// you have to create
				SecuritySchemes: map[string]*gofiberswagger.SecuritySchemeRef{
					"User API Token": {
						Value: &gofiberswagger.SecurityScheme{
							Type: "apiKey",
							Name: "x-user-token", // header name
							In:   "header",
						},
					},
				},
			},
		},
		// Option B: uncomment the following 2 lines and automatically require auth for all routes
		// AutomaticallyRequireAuth: true,
		// RequiredAuth:             security_requirements_docs,
		SwaggerUI: gofiberswagger.DefaultUIConfig,
	})

	log.Fatal(app.Listen(":3000"))
}

func HelloHandler(c fiber.Ctx) error {
	return c.SendStatus(200)
}
