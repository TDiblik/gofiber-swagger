package gofiberswagger

import (
	"github.com/getkin/kin-openapi/openapi3"
)

type Operation = openapi3.Operation
type Origin = openapi3.Origin
type Location = openapi3.Location
type Parameters = openapi3.Parameters
type ParameterRef = openapi3.ParameterRef
type Parameter = openapi3.Parameter
type Examples = openapi3.Examples
type ExampleRef = openapi3.ExampleRef
type Example = openapi3.Example
type Content = openapi3.Content
type MediaType = openapi3.MediaType
type Encoding = openapi3.Encoding
type Headers = openapi3.Headers
type HeaderRef = openapi3.HeaderRef
type Header = openapi3.Header
type SchemaRef = openapi3.SchemaRef
type Schema = openapi3.Schema
type SchemaRefs = openapi3.SchemaRefs
type Types = openapi3.Types
type ExternalDocs = openapi3.ExternalDocs
type RequestBodyRef = openapi3.RequestBodyRef
type RequestBody = openapi3.RequestBody
type Responses = openapi3.Responses
type ResponseRef = openapi3.ResponseRef
type Callbacks = openapi3.Callbacks
type Links = openapi3.Links
type ParametersMap = openapi3.ParametersMap
type RequestBodies = openapi3.RequestBodies
type ResponseBodies = openapi3.ResponseBodies
type Schemas = openapi3.Schemas
type Components = openapi3.Components
type Response = openapi3.Response
type SecuritySchemes = openapi3.SecuritySchemes
type SecuritySchemeRef = openapi3.SecuritySchemeRef
type SecurityScheme = openapi3.SecurityScheme
type OAuthFlows = openapi3.OAuthFlows
type OAuthFlow = openapi3.OAuthFlow
type StringMap[V any] = openapi3.StringMap[V]
type Info = openapi3.Info
type Contact = openapi3.Contact
type License = openapi3.License
type Paths = openapi3.Paths
type PathItem = openapi3.PathItem
type Servers = openapi3.Servers
type Server = openapi3.Server
type ServerVariable = openapi3.ServerVariable
type SecurityRequirements = openapi3.SecurityRequirements
type SecurityRequirement = openapi3.SecurityRequirement
type Tags = openapi3.Tags
type Tag = openapi3.Tag
type NewResponsesOption = openapi3.NewResponsesOption
type AdditionalProperties = openapi3.AdditionalProperties
type XML = openapi3.XML
type Discriminator = openapi3.Discriminator

// ----- Request Body ----- //
var DefaultRequestBodyConsumes = []string{
	"application/json", "application/xml", "application/x-www-form-urlencoded", "multipart/form-data", // all supported by the `c.Bind().Body()` function
}

func NewRequestBody[T any]() *RequestBodyRef {
	return NewRequestBodyExtended[T]("", false)
}
func NewRequestBodyExtended[T any](description string, required bool) *RequestBodyRef {
	return NewRequestBodyFullyCustom[T](description, required, DefaultRequestBodyConsumes)
}

func NewRequestBodyJSON[T any]() *RequestBodyRef {
	return NewRequestBodyJSONExtended[T]("", false)
}
func NewRequestBodyJSONExtended[T any](description string, required bool) *RequestBodyRef {
	return NewRequestBodyFullyCustom[T](description, required, []string{"application/json"})
}

func NewRequestBodyFormData[T any]() *RequestBodyRef {
	return NewRequestBodyFormDataExtended[T]("", false)
}
func NewRequestBodyFormDataExtended[T any](description string, required bool) *RequestBodyRef {
	return NewRequestBodyFullyCustom[T](description, required, []string{"multipart/form-data"})
}

func NewRequestBodyFormUrlEncodedData[T any]() *RequestBodyRef {
	return NewRequestBodyFormUrlEncodedDataExtended[T]("", false)
}
func NewRequestBodyFormUrlEncodedDataExtended[T any](description string, required bool) *RequestBodyRef {
	return NewRequestBodyFullyCustom[T](description, required, []string{"application/x-www-form-urlencoded"})
}

func NewRequestBodyXML[T any]() *RequestBodyRef {
	return NewRequestBodyXMLExtended[T]("", false)
}
func NewRequestBodyXMLExtended[T any](description string, required bool) *RequestBodyRef {
	return NewRequestBodyFullyCustom[T](description, required, []string{"application/xml"})
}

func NewRequestBodyFullyCustom[T any](description string, required bool, consumes []string) *RequestBodyRef {
	request_body := openapi3.NewRequestBody()
	request_body.WithDescription(description)
	request_body.WithRequired(required)
	schema := CreateSchema[T]()
	request_body.WithSchemaRef(schema, consumes)
	return &RequestBodyRef{Value: request_body}
}

// ----- Response ----- //
type ResponseInfo struct {
	Code        string
	Description string
	Response    *ResponseRef
}

func NewResponses(responses ...ResponseInfo) *Responses {
	output := &Responses{}
	for _, response := range responses {
		if response.Response != nil && response.Response.Value != nil && (response.Response.Value.Description == nil || *response.Response.Value.Description == "") && *response.Response.Value.Description != response.Description {
			response.Response.Value.WithDescription(response.Description)
		}
		output.Set(response.Code, response.Response)
	}
	return output
}
func NewResponseInfo[T any](code string, description string) ResponseInfo {
	return ResponseInfo{
		Code:        code,
		Description: description,
		Response:    NewResponseRawJSON[T](description),
	}
}
func NewResponseInfoRaw[T any](code string, description string, mediatype string, additonalMediaTypeInfo *MediaType) ResponseInfo {
	return ResponseInfo{
		Code:        code,
		Description: description,
		Response:    NewResponseRaw[T](mediatype, description, additonalMediaTypeInfo),
	}
}

func NewResponsesRaw(responses map[string]*ResponseRef) *Responses {
	output := &Responses{}
	for k, v := range responses {
		output.Set(k, v)
	}
	return output
}
func NewResponseRawJSON[T any](description string) *ResponseRef {
	response := openapi3.NewResponse()
	schema := CreateSchema[T]()
	response.WithJSONSchemaRef(schema)
	response.WithDescription(description)
	return &ResponseRef{Value: response}
}

func NewResponseRaw[T any](mediaType string, description string, additonalMediaTypeInfo *MediaType) *ResponseRef {
	response := openapi3.NewResponse()

	if additonalMediaTypeInfo == nil {
		additonalMediaTypeInfo = &MediaType{}
	}
	if additonalMediaTypeInfo.Schema == nil {
		schema := CreateSchema[T]()
		additonalMediaTypeInfo.Schema = schema
	}

	response.WithContent(
		Content{
			mediaType: additonalMediaTypeInfo,
		},
	)
	response.WithDescription(description)
	return &ResponseRef{Value: response}
}

// ----- Parameter ----- //
func NewParameters(parameters ...*ParameterRef) Parameters {
	result := Parameters{}
	for _, parameter := range parameters {
		result = append(result, parameter)
	}
	return result
}

func INewPathParameter[T any](name string) *ParameterRef {
	param_raw := openapi3.NewPathParameter(name)
	param_raw.Schema = CreateSchema[T]()
	return &ParameterRef{Value: param_raw}
}

func NewPathParameter(name string) *ParameterRef {
	return NewPathParameterExtended(name, &Schema{
		Type: &Types{"string"},
	})
}

func NewPathParameterWithType(name string, Type string) *ParameterRef {
	return NewPathParameterExtended(name, &Schema{
		Type: &Types{Type},
	})
}

func NewPathParameterExtended(name string, schema *Schema) *ParameterRef {
	return &ParameterRef{Value: openapi3.NewPathParameter(name).WithSchema(schema)}
}

func INewQueryParameter[T any](name string) *ParameterRef {
	param_raw := openapi3.NewQueryParameter(name)
	param_raw.Schema = CreateSchema[T]()
	return &ParameterRef{Value: param_raw}
}

func NewQueryParameter(name string) *ParameterRef {
	return NewQueryParameterExtended(name, &Schema{
		Type: &Types{"string"},
	})
}

func NewQueryParameterRequired(name string) *ParameterRef {
	param := NewQueryParameterExtended(name, &Schema{
		Type:            &Types{"string"},
		AllowEmptyValue: false,
		Nullable:        false,
	})
	param.Value.WithRequired(true)
	return param
}

func NewQueryParameterWithType(name string, Type string) *ParameterRef {
	return NewQueryParameterExtended(name, &Schema{
		Type: &Types{Type},
	})
}

func NewQueryParameterExtended(name string, schema *Schema) *ParameterRef {
	return &ParameterRef{Value: openapi3.NewQueryParameter(name).WithSchema(schema)}
}

func INewHeaderParameter[T any](name string) *ParameterRef {
	param_raw := openapi3.NewHeaderParameter(name)
	param_raw.Schema = CreateSchema[T]()
	return &ParameterRef{Value: param_raw}
}

func NewHeaderParameter(name string) *ParameterRef {
	return NewHeaderParameterExtended(name, &Schema{
		Type: &Types{"string"},
	})
}

func NewHeaderParameterRequired(name string) *ParameterRef {
	param := NewHeaderParameterExtended(name, &Schema{
		Type:            &Types{"string"},
		AllowEmptyValue: false,
		Nullable:        false,
	})
	param.Value.WithRequired(true)
	return param
}

func NewHeaderParameterWithType(name string, Type string) *ParameterRef {
	return NewHeaderParameterExtended(name, &Schema{
		Type: &Types{Type},
	})
}

func NewHeaderParameterExtended(name string, schema *Schema) *ParameterRef {
	return &ParameterRef{Value: openapi3.NewHeaderParameter(name).WithSchema(schema)}
}

func INewCookieParameter[T any](name string) *ParameterRef {
	param_raw := openapi3.NewCookieParameter(name)
	param_raw.Schema = CreateSchema[T]()
	return &ParameterRef{Value: param_raw}
}

func NewCookieParameter(name string) *ParameterRef {
	return NewCookieParameterExtended(name, &Schema{
		Type: &Types{"string"},
	})
}

func NewCookieParameterRequired(name string) *ParameterRef {
	param := NewCookieParameterExtended(name, &Schema{
		Type:            &Types{"string"},
		AllowEmptyValue: false,
		Nullable:        false,
	})
	param.Value.WithRequired(true)
	return param
}

func NewCookieParameterWithType(name string, Type string) *ParameterRef {
	return NewCookieParameterExtended(name, &Schema{
		Type: &Types{Type},
	})
}

func NewCookieParameterExtended(name string, schema *Schema) *ParameterRef {
	return &ParameterRef{Value: openapi3.NewCookieParameter(name).WithSchema(schema)}
}

func NewPathParameterRaw(name string) *Parameter {
	return openapi3.NewPathParameter(name)
}

func NewQueryParameterRaw(name string) *Parameter {
	return openapi3.NewQueryParameter(name)
}

func NewHeaderParameterRaw(name string) *Parameter {
	return openapi3.NewHeaderParameter(name)
}

func NewCookieParameterRaw(name string) *Parameter {
	return openapi3.NewCookieParameter(name)
}

// ----- Schema ----- //
func NewOneOfSchema(schemas ...*Schema) *Schema {
	return openapi3.NewOneOfSchema(schemas...)
}

func NewAnyOfSchema(schemas ...*Schema) *Schema {
	return openapi3.NewAnyOfSchema(schemas...)
}

func NewAllOfSchema(schemas ...*Schema) *Schema {
	return openapi3.NewAllOfSchema(schemas...)
}

func NewBoolSchema() *Schema {
	return openapi3.NewBoolSchema()
}

func NewFloat64Schema() *Schema {
	return openapi3.NewFloat64Schema()
}

func NewIntegerSchema() *Schema {
	return openapi3.NewIntegerSchema()
}

func NewInt32Schema() *Schema {
	return openapi3.NewInt32Schema()
}

func NewInt64Schema() *Schema {
	return openapi3.NewInt64Schema()
}

func NewStringSchema() *Schema {
	return openapi3.NewStringSchema()
}

func NewDateTimeSchema() *Schema {
	return openapi3.NewDateTimeSchema()
}

func NewUUIDSchema() *Schema {
	return openapi3.NewUUIDSchema()
}

func NewBytesSchema() *Schema {
	return openapi3.NewBytesSchema()
}

func NewArraySchema() *Schema {
	return openapi3.NewArraySchema()
}

func NewObjectSchema() *Schema {
	return openapi3.NewObjectSchema()
}
