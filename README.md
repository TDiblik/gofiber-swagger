# Golang Fiber Swagger generation

### What

This library generates swagger documentation based on your codebase. It generates as much as possible, so you can just drop it in. However, if you want to customize it, it's basically a wrapper around [github.com/getkin/kin-openapi](https://github.com/getkin/kin-openapi) (with re-exported types) for the [fiber web framework](https://gofiber.io/), so you can let your wildest openapi dreams come true haha.

### How

It uses context from `fiber.App` to generate routes with parameters and then uses additional context provided (types) by the user of the library to generate `openapi` schemas using reflection.

You can either:

- a) Use the `gofiberswagger.NewRouter` to create a router which acts like the `fiber.Router`, but takes `*RouteInfo` for swagger docs as the second argument.
- b) Use the `gofiberswagger.RegisterRoute` function to manually register a route and it's info.

### Why

I really, really, really, hate defining the swagger docs using [swaggo/swag](https://github.com/swaggo/swag). It's a cool project and you should totally check it out, but it just isn't for me.

### Example

Here's an example that showcases most of the features provided. You can find many more examples in the `/examples/` directory

```go
package main

import (
	"github.com/TDiblik/gofiber-swagger/gofiberswagger"
	"github.com/gofiber/fiber/v3"
)

func main() {
	app := fiber.New()

	// Create wrapper around the fiber router (optional, recommended)
	router := gofiberswagger.NewRouter(app)

	// Normally create your routes. You don't have to provide the swagger options (second argument)
	router.Get("/", nil, GETHelloHandler)

	// You can group your routes normally
	parametersGroup := router.Group("/parameters/")

	// Path parameters get automatically recognized :)
	parametersGroup.Get("/:id", nil, GETHandlerWithId)

	// This is how you specify RequestBody / ResponseBody / different parameters / any other openapi property tied to a request
	parametersGroup.Post("/:id", &gofiberswagger.RouteInfo{
		Parameters: gofiberswagger.NewParameters(
			// We can also specify additonal parameters, for example query parameters
			gofiberswagger.NewQueryParameter("queryParam"),
		),
		RequestBody: gofiberswagger.NewRequestBody[POSTHandlerWithIdRequestBody](),
		Responses: gofiberswagger.NewResponses(
			gofiberswagger.NewResponseInfo[POSTHandlerWithIdResponse]("200", "example response 👀"),
		),
	}, POSTHandlerWithId)

	// You can also manully register routes without touching your existing code!
	app.Get("/abc", GETHelloHandler)
	gofiberswagger.RegisterRoute("GET", "/", &gofiberswagger.RouteInfo{})

	// Register swagger. Without this line, nothing will get generated.
	// For more config customizability, see /examples/custom-config/main.go
    // You can now see your:
    // - UI at /swagger/
    // - json at /swagger/swagger.json
    // - yaml at /swagger/swagger.yaml
	gofiberswagger.Register(app, gofiberswagger.DefaultConfig)

	log.Fatal(app.Listen(":3000"))
}

func GETHelloHandler(c fiber.Ctx) error {
	return c.SendStatus(200)
}
func GETHandlerWithId(c fiber.Ctx) error {
	return c.Status(200).JSON(fiber.Map{"id": c.Params("id", "no id provided!")})
}

// ----- Handler with custom request body it's types ----- //
type POSTHandlerWithIdRequestBody struct {
	A int32    `json:"a" validate:"required,min=1,max=10"`
	B string   `json:"b" validate:"required"`
	C []string `json:"c" validate:"require,min=1"`
}
type EmbeddedField struct {
	A int32
	B string
	C []string
}
type POSTHandlerWithIdResponse struct {
	Status        string        `json:"status"`
	Id            string        `json:"id"`
	QueryParam    string        `json:"query_param"`
	EmbeddedField EmbeddedField `json:"embedded_field"`
}

func POSTHandlerWithId(c fiber.Ctx) error {
	id := c.Params("id", "no id provided!")
	queryParam := c.Query("queryParam", "no queryParam provided!")
	request_body := new(POSTHandlerWithIdRequestBody)
	if err := c.Bind().Body(request_body); err != nil {
		return c.Status(400).JSON(fiber.Map{"status": "error", "msg": "Invalid request body"})
	}

	// in real application, you'd want to validate the struct here,
	// however that would overcomplicate our basic example

	response := POSTHandlerWithIdResponse{
		Status:     "ok",
		Id:         id,
		QueryParam: queryParam,
		EmbeddedField: EmbeddedField{
			A: request_body.A,
			B: request_body.B,
			C: request_body.C,
		},
	}
	return c.Status(200).JSON(response)
}
```

### Custom Swagger path

Set `SwaggerPath` to serve the Swagger UI and schemas under a custom path:

```go
config := gofiberswagger.DefaultConfig
config.SwaggerPath = "/docs"
if err := gofiberswagger.Register(app, config); err != nil {
	log.Fatal(err)
}
```

The UI is available at `/docs/` and `/docs/index.html`, with schemas at
`/docs/swagger.json` and `/docs/swagger.yaml`. If `SwaggerPath` is empty, it defaults
to `/swagger`. Leading and trailing slashes are normalized.

The default `SwaggerUI.URL` automatically follows `SwaggerPath`. Set it to a custom
schema URL to load a different API definition.

### Notes

Even though this library is in the early stages of development, from my personal experience, it's quite stable 🤷‍♂️.
