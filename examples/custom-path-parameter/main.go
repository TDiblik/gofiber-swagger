package main

import (
	"log"

	"github.com/TDiblik/gofiber-swagger/gofiberswagger"
	"github.com/gofiber/fiber/v3"
)

func main() {
	app := fiber.New()
	router := gofiberswagger.NewRouter(app)

	router.Post("/update/:id", &gofiberswagger.RouteInfo{
		Parameters: gofiberswagger.NewParameters(
			gofiberswagger.NewPathParameterExtended("id", &gofiberswagger.Schema{
				Type:   &gofiberswagger.Types{"integer"},
				Format: "int64",
			}),
		),
	}, func(c fiber.Ctx) error {
		return c.SendString("Update ID: " + c.Params("id"))
	})

	gofiberswagger.Register(app, gofiberswagger.DefaultConfig)

	log.Println("Server started on http://localhost:3000")
	log.Println("Swagger UI available at http://localhost:3000/swagger/")
	log.Fatal(app.Listen(":3000"))
}
