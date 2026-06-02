package main

import (
	"log"

	"github.com/TDiblik/gofiber-swagger/gofiberswagger"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
)

// ExampleRequest struct demonstrating the swagger tags
type ExampleRequest struct {
	// This field will be completely ignored in the generated swagger schema.
	InternalSecret string `json:"internal_secret" swaggerignore:"true"`

	// This field's type will be overridden to "integer" in the swagger schema.
	OverriddenString string `json:"overridden_string" swaggertype:"integer"`

	// This field's type will be overridden to an array of integers.
	OverriddenArray []string `json:"overridden_array" swaggertype:"[]integer"`
}

// ExampleResponse struct demonstrating the swagger tags
type ExampleResponse struct {
	// Even though this is an array of strings in Go, we override it to be an "object" in the Swagger documentation.
	Data []string `json:"data" swaggertype:"object"`
}

func main() {
	app := fiber.New()

	app.Use(cors.New())
	app.Use(logger.New())

	// Create wrapper around the fiber router
	router := gofiberswagger.NewRouter(app)

	// Register a route
	router.Post("/tags-example", &gofiberswagger.RouteInfo{
		RequestBody: gofiberswagger.NewRequestBody[ExampleRequest](),
		Responses: gofiberswagger.NewResponses(
			gofiberswagger.NewResponseInfo[ExampleResponse]("200", "Example response showing tag effects"),
		),
	}, TagsExampleHandler)

	// Register swagger. Without this line, nothing will get generated.
	// You can now see your:
	// - UI at /swagger/
	// - json at /swagger/swagger.json
	// - yaml at /swagger/swagger.yaml
	gofiberswagger.Register(app, gofiberswagger.DefaultConfig)

	log.Println("Listening on http://localhost:3000")
	log.Println("Swagger UI available at http://localhost:3000/swagger/")
	log.Fatal(app.Listen(":3000"))
}

func TagsExampleHandler(c fiber.Ctx) error {
	var req ExampleRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	res := ExampleResponse{
		Data: []string{"this", "is", "an", "array", "but", "swagger", "thinks", "its", "an", "object"},
	}

	return c.Status(200).JSON(res)
}
