package main

import (
	"log"
	"mime/multipart"

	"github.com/TDiblik/gofiber-swagger/gofiberswagger"
	"github.com/gofiber/fiber/v3"
)

func main() {
	app := fiber.New()

	router := gofiberswagger.NewRouter(app)
	router.Get("/", nil, HelloHandler)
	router.Post("/upload", &gofiberswagger.RouteInfo{
		RequestBody: gofiberswagger.NewRequestBodyFormData[UploadRequest](),
		Responses: gofiberswagger.NewResponses(
			gofiberswagger.NewResponseInfo[struct {
				status string
				file   multipart.FileHeader
			}]("200", "OK"),
		),
	}, UploadHandler)

	// You can now see your:
	// - UI at /swagger/
	// - json at /swagger/swagger.json
	// - yaml at /swagger/swagger.yaml
	gofiberswagger.Register(app, gofiberswagger.DefaultConfig)

	log.Fatal(app.Listen(":3000"))
}

// ----- Hello Handler and it's types ----- //
func HelloHandler(c fiber.Ctx) error {
	return c.SendStatus(200)
}

type UploadRequest struct {
	File1 *multipart.FileHeader    `form:"file1" validate:"required"`
	File2 multipart.FileHeader     `form:"file2"`
	Files *[]*multipart.FileHeader `form:"files"`
}

func UploadHandler(c fiber.Ctx) error {
	file, err := c.FormFile("file1")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"status": "error", "msg": "This API endpoint requires \"file1\" submitted as a form file."})
	}

	return c.Status(200).JSON(fiber.Map{"status": "ok", "file": file})
}
