package gofiberswagger

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
)

func TestNewRouter(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	router := NewRouter(app)

	assert.NotNil(t, router)
	assert.Equal(t, "", router.internalGroup)
	assert.NotNil(t, router.Router)
}

func TestNewRouterFromRouter(t *testing.T) {
	t.Parallel()

	app := fiber.New()
	group := app.Group("/test")
	router := NewRouterFromRouter(group)

	assert.NotNil(t, router)
	assert.Equal(t, "", router.internalGroup)
	assert.NotNil(t, router.Router)
}

func TestSwaggerRouter_Methods(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		method string
	}{
		{"GET"},
		{"POST"},
		{"PUT"},
		{"DELETE"},
		{"PATCH"},
		{"HEAD"},
		{"OPTIONS"},
		{"CONNECT"},
		{"TRACE"},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			t.Parallel()

			// setup
			app := fiber.New()
			router := NewRouter(app)
			path := fmt.Sprintf("/%s", tc.method)
			handler := func(c fiber.Ctx) error {
				return c.SendString(tc.method)
			}
			docs := &RouteInfo{
				Summary: "Test endpoint",
			}

			// execute
			switch tc.method {
			case "GET":
				router.Get(path, docs, handler)
			case "POST":
				router.Post(path, docs, handler)
			case "PUT":
				router.Put(path, docs, handler)
			case "DELETE":
				router.Delete(path, docs, handler)
			case "PATCH":
				router.Patch(path, docs, handler)
			case "HEAD":
				router.Head(path, docs, handler)
			case "OPTIONS":
				router.Options(path, docs, handler)
			case "CONNECT":
				router.Connect(path, docs, handler)
			case "TRACE":
				router.Trace(path, docs, handler)
			}

			// verify route is registered
			req := httptest.NewRequest(tc.method, path, nil)
			resp, err := app.Test(req)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.StatusCode)

			// verify docs are registered
			registeredDocs := getAcquiredRoutesInfo(tc.method, path)
			assert.NotNil(t, registeredDocs)
			assert.Equal(t, "Test endpoint", registeredDocs.Summary)
		})
	}
}

func TestSwaggerRouter_Group(t *testing.T) {
	t.Parallel()

	// setup
	app := fiber.New()
	router := NewRouter(app)

	// execute
	group := router.Group("/test")
	group.Get("/endpoint", &RouteInfo{Summary: "Group endpoint"}, func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	// verify route is registered
	req := httptest.NewRequest("GET", "/test/endpoint", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// verify docs are registered with group tag
	registeredDocs := getAcquiredRoutesInfo("GET", "/test/endpoint")
	assert.NotNil(t, registeredDocs)
	assert.Equal(t, "Group endpoint", registeredDocs.Summary)
	assert.Contains(t, registeredDocs.Tags, "/test")
}
