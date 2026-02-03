package gofiberswagger

import (
	"bytes"
	"errors"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/gofiber/fiber/v3"
	"gopkg.in/yaml.v3"
)

func Register(app *fiber.App, config Config) error {
	config.Swagger = swaggerConfigDefault(config.Swagger)
	config.SwaggerUI = swaggerUIConfigDefault(config.SwaggerUI)

	for k, v := range acquiredSchemas {
		if config.Swagger.Components.Schemas[k] == nil {
			config.Swagger.Components.Schemas[k] = v
		}
	}

	routes := app.GetRoutes(config.FilterOutAppUse)
	for _, route := range routes {
		operation := getAcquiredRoutesInfo(route.Method, route.Path)
		if operation == nil {
			operation = &RouteInfo{}
		}

		corrected_path := route.Path
		for _, param_name := range route.Params {
			parameter := NewPathParameter(param_name)
			parameter.Value = parameter.Value.WithSchema(NewStringSchema())
			operation.AddParameter(parameter.Value)

			corrected_path = strings.Replace(corrected_path, ":"+param_name, "{"+param_name+"}", 1)
			if param_name[0] == '*' || param_name[0] == '+' {
				char_to_replace := "*"
				if param_name[0] == '+' {
					char_to_replace = "+"
				}

				nth, err := strconv.ParseUint(strings.ReplaceAll(param_name, char_to_replace, ""), 10, 64)
				if err != nil {
					return errors.Join(errors.New("unable to parse out the nth position of the param_name \""+param_name+"\""), err)
				}
				corrected_path = replaceNthOccurrence(corrected_path, char_to_replace, "{"+param_name+"}", int(nth))
			}
		}
		if config.AppendMethodToTags {
			operation.Tags = append(operation.Tags, route.Method)
		}

		if config.AutomaticallyRequireAuth && config.RequiredAuth != nil {
			if operation.Security == nil {
				operation.Security = &openapi3.SecurityRequirements{}
			}
			for _, v := range *config.RequiredAuth {
				operation.Security.With(v)
			}
		}

		if operation.Responses == nil {
			operation.Responses = &Responses{}
		}

		path_item := config.Swagger.Paths.Find(corrected_path)
		if path_item == nil {
			path_item = &openapi3.PathItem{}
		}
		switch route.Method {
		case "POST":
			path_item.Post = operation
		case "CONNECT":
			path_item.Connect = operation
		case "DELETE":
			path_item.Delete = operation
		case "GET":
			path_item.Get = operation
		case "HEAD":
			path_item.Head = operation
		case "OPTIONS":
			path_item.Options = operation
		case "PATCH":
			path_item.Patch = operation
		case "PUT":
			path_item.Put = operation
		case "TRACE":
			path_item.Trace = operation
		default:
			log.Println("gofiber-swagger: unable to translate operation \"", route.Method, "\", skipping...")
		}
		config.Swagger.Paths.Set(corrected_path, path_item)
	}

	index_page, err := generateIndexPage(swaggerUIConfigDefault(config.SwaggerUI))
	if err != nil {
		return err
	}
	schema_as_json, schema_as_yaml, err := generateOpenApiSchema(config.Swagger)
	if err != nil {
		return err
	}

	if config.CreateSwaggerFiles && !fiber.IsChild() {
		if config.SwaggerFilesPath == "" {
			return errors.New("gofiber-swagger: CreateSwaggerFiles was set to true, however SwaggerFilesPaths was left empty")
		}
		createSwaggerFiles(config.SwaggerFilesPath, index_page, schema_as_json, schema_as_yaml)
	}

	swagger_routes := app.Group("/swagger")
	index_handler := func(c fiber.Ctx) error {
		return c.Type("html").Send(index_page)
	}
	swagger_routes.Get("/", index_handler)
	swagger_routes.Get("/index.html", index_handler)
	swagger_routes.Get("/swagger", index_handler)
	swagger_routes.Get("/swagger.json", func(c fiber.Ctx) error {
		return c.Type("json").Send(schema_as_json)
	})
	swagger_routes.Get("/swagger.yaml", func(c fiber.Ctx) error {
		return c.Type("yaml").Send(schema_as_yaml)
	})

	return nil
}

func generateIndexPage(ui_config SwaggerUIConfig) (index_page []byte, err error) {
	index_tpl, err := template.New("swagger_index.html").Parse(indexPageTmpl)
	if err != nil {
		return nil, errors.Join(errors.New("gofiber-swagger: error while parsing the swagger index template -> "), err)
	}
	index_tpl_buf := bytes.NewBufferString("")
	err = index_tpl.Execute(index_tpl_buf, ui_config)
	if err != nil {
		return nil, errors.Join(errors.New("gofiber-swagger: error while executing the swagger index template -> "), err)
	}
	return index_tpl_buf.Bytes(), nil
}

func generateOpenApiSchema(schema openapi3.T) (as_json, as_yaml []byte, err error) {
	schema_as_yaml_raw, err := schema.MarshalYAML()
	if err != nil {
		return nil, nil, errors.Join(errors.New("gofiber-swagger: error while creating the yaml schema -> "), err)
	}
	schema_as_yaml, err := yaml.Marshal(schema_as_yaml_raw)
	if err != nil {
		return nil, nil, errors.Join(errors.New("gofiber-swagger: error while converting the yaml schema to yaml -> "), err)
	}

	schema_as_json, err := schema.MarshalJSON()
	if err != nil {
		return nil, nil, errors.Join(errors.New("gofiber-swagger: error while creating the json schema -> "), err)
	}

	return schema_as_json, schema_as_yaml, nil
}

func createSwaggerFiles(target_folder_path string, index_page []byte, schema_as_json []byte, schema_as_yaml []byte) error {
	var creation_perms os.FileMode = 0o766

	if err := os.MkdirAll(target_folder_path, creation_perms); err != nil {
		return errors.Join(errors.New("unable to create file directory for swagger files"), err)
	}

	if err := os.WriteFile(filepath.Join(target_folder_path, "index.html"), index_page, creation_perms); err != nil {
		return errors.Join(errors.New("unable to create index.html for swagger files"), err)
	}

	if err := os.WriteFile(filepath.Join(target_folder_path, "swagger.json"), schema_as_json, creation_perms); err != nil {
		return errors.Join(errors.New("unable to create swagger.json for swagger files"), err)
	}

	if err := os.WriteFile(filepath.Join(target_folder_path, "swagger.yaml"), schema_as_yaml, creation_perms); err != nil {
		return errors.Join(errors.New("unable to create swagger.yaml for swagger files"), err)
	}

	return nil
}
