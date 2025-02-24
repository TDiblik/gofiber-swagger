package main

import (
	"github.com/TDiblik/gofiber-swagger/gofiberswagger"
	"github.com/gofiber/fiber/v3"
)

func main() {
	app := fiber.New()

	// equivalent to:
	// router := gofiberswagger.NewRouter(app)
	// router.Get("/", nil, HelloHandler)
	// router.Get("/abc", nil, HelloHandler)
	// router.Get("/bca", nil, HelloHandler)

	app.Get("/", HelloHandler)
	app.Get("/abc", HelloHandler)
	app.Get("/bca", HelloHandler)
	gofiberswagger.RegisterPath("GET", "/", &gofiberswagger.RouteInfo{})
	gofiberswagger.RegisterPath("GET", "/abc", &gofiberswagger.RouteInfo{})
	gofiberswagger.RegisterPath("GET", "/bca", &gofiberswagger.RouteInfo{})

	gofiberswagger.Register(app, gofiberswagger.DefaultConfig)

	app.Listen(":3000")
}

// ----- Hello Handler and it's types ----- //
func HelloHandler(c fiber.Ctx) error {
	return c.SendStatus(200)
}
