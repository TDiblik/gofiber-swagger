// This file was initially copied from https://github.com/gofiber/swagger/blob/6ced278517482f1bf4bab7aa82105ed770921dac/config.go.
// Shoutout to the original authors! Since then, we've made several changes, but credit goes to them for the foundation.

package gofiberswagger

import (
	"html/template"
)

// SwaggerUIConfig stores SwaggerUI configuration variables
type SwaggerUIConfig struct {
	// This parameter can be used to name different swagger document instances.
	// default: ""
	InstanceName string `json:"-"`

	// Title pointing to title of HTML page.
	// default: "Swagger UI"
	Title string `json:"-"`

	// URL to fetch external configuration document from.
	// default: ""
	ConfigURL string `json:"configUrl,omitempty"`

	// The URL pointing to API definition (normally swagger.json or swagger.yaml).
	// default: SwaggerPath + "/swagger.yaml" (normally "/swagger/swagger.yaml").
	URL string `json:"url,omitempty"`

	// Enables overriding configuration parameters via URL search params.
	// default: false
	QueryConfigEnabled bool `json:"queryConfigEnabled,omitempty"`

	// The name of a component available via the plugin system to use as the top-level layout for Swagger UI.
	// default: "StandaloneLayout"
	Layout string `json:"layout,omitempty"`

	// An array of plugin functions to use in Swagger UI.
	// default: [SwaggerUIBundle.plugins.DownloadUrl]
	Plugins []template.JS `json:"-"`

	// An array of presets to use in Swagger UI. Usually, you'll want to include ApisPreset if you use this option.
	// default: [SwaggerUIBundle.presets.apis, SwaggerUIStandalonePreset]
	Presets []template.JS `json:"-"`

	// If set to true, enables deep linking for tags and operations.
	// default: true
	DeepLinking bool `json:"deepLinking"`

	// Controls the display of operationId in operations list.
	// default: false
	DisplayOperationId bool `json:"displayOperationId,omitempty"`

	// The default expansion depth for models (set to -1 completely hide the models).
	// default: 1
	DefaultModelsExpandDepth int `json:"defaultModelsExpandDepth,omitempty"`

	// The default expansion depth for the model on the model-example section.
	// default: 1
	DefaultModelExpandDepth int `json:"defaultModelExpandDepth,omitempty"`

	// Controls how the model is shown when the API is first rendered.
	// The user can always switch the rendering for a given model by clicking the 'Model' and 'Example Value' links.
	// default: "example"
	DefaultModelRendering string `json:"defaultModelRendering,omitempty"`

	// Controls the display of the request duration (in milliseconds) for "Try it out" requests.
	// default: false
	DisplayRequestDuration bool `json:"displayRequestDuration,omitempty"`

	// Controls the default expansion setting for the operations and tags.
	// 'list' (default, expands only the tags),
	// 'full' (expands the tags and operations),
	// 'none' (expands nothing)
	DocExpansion string `json:"docExpansion,omitempty"`

	// If set, enables filtering. The top bar will show an edit box that you can use to filter the tagged operations that are shown.
	// Can be Boolean to enable or disable, or a string, in which case filtering will be enabled using that string as the filter expression.
	// Filtering is case sensitive matching the filter expression anywhere inside the tag.
	// default: false
	Filter FilterConfig `json:"-"`

	// If set, limits the number of tagged operations displayed to at most this many. The default is to show all operations.
	// default: 0
	MaxDisplayedTags int `json:"maxDisplayedTags,omitempty"`

	// Controls the display of vendor extension (x-) fields and values for Operations, Parameters, Responses, and Schema.
	// default: false
	ShowExtensions bool `json:"showExtensions,omitempty"`

	// Controls the display of extensions (pattern, maxLength, minLength, maximum, minimum) fields and values for Parameters.
	// default: false
	ShowCommonExtensions bool `json:"showCommonExtensions,omitempty"`

	// Apply a sort to the tag list of each API. It can be 'alpha' (sort by paths alphanumerically) or a function (see Array.prototype.sort().
	// to learn how to write a sort function). Two tag name strings are passed to the sorter for each pass.
	// default: "" -> Default is the order determined by Swagger UI.
	TagsSorter template.JS `json:"-"`

	// Provides a mechanism to be notified when Swagger UI has finished rendering a newly provided definition.
	// default: "" -> Function=NOOP
	OnComplete template.JS `json:"-"`

	// An object with the activate and theme properties.
	SyntaxHighlight *SyntaxHighlightConfig `json:"-"`

	// Controls whether the "Try it out" section should be enabled by default.
	// default: false
	TryItOutEnabled bool `json:"tryItOutEnabled,omitempty"`

	// Enables the request snippet section. When disabled, the legacy curl snippet will be used.
	// default: false
	RequestSnippetsEnabled bool `json:"requestSnippetsEnabled,omitempty"`

	// OAuth redirect URL.
	// default: ""
	OAuth2RedirectUrl string `json:"oauth2RedirectUrl,omitempty"`

	// MUST be a function. Function to intercept remote definition, "Try it out", and OAuth 2.0 requests.
	// Accepts one argument requestInterceptor(request) and must return the modified request, or a Promise that resolves to the modified request.
	// default: ""
	RequestInterceptor template.JS `json:"-"`

	// If set, MUST be an array of command line options available to the curl command. This can be set on the mutated request in the requestInterceptor function.
	// For example request.curlOptions = ["-g", "--limit-rate 20k"]
	// default: nil
	RequestCurlOptions []string `json:"request.curlOptions,omitempty"`

	// MUST be a function. Function to intercept remote definition, "Try it out", and OAuth 2.0 responses.
	// Accepts one argument responseInterceptor(response) and must return the modified response, or a Promise that resolves to the modified response.
	// default: ""
	ResponseInterceptor template.JS `json:"-"`

	// If set to true, uses the mutated request returned from a requestInterceptor to produce the curl command in the UI,
	// otherwise the request before the requestInterceptor was applied is used.
	// default: true
	ShowMutatedRequest bool `json:"showMutatedRequest"`

	// List of HTTP methods that have the "Try it out" feature enabled. An empty array disables "Try it out" for all operations.
	// This does not filter the operations from the display.
	// Possible values are ["get", "put", "post", "delete", "options", "head", "patch", "trace"]
	// default: nil
	SupportedSubmitMethods []string `json:"supportedSubmitMethods,omitempty"`

	// By default, Swagger UI attempts to validate specs against swagger.io's online validator. You can use this parameter to set a different validator URL.
	// For example for locally deployed validators (https://github.com/swagger-api/validator-badge).
	// Setting it to either none, 127.0.0.1 or localhost will disable validation.
	// default: ""
	ValidatorUrl string `json:"validatorUrl,omitempty"`

	// If set to true, enables passing credentials, as defined in the Fetch standard, in CORS requests that are sent by the browser.
	// Note that Swagger UI cannot currently set cookies cross-domain (see https://github.com/swagger-api/swagger-js/issues/1163).
	// as a result, you will have to rely on browser-supplied cookies (which this setting enables sending) that Swagger UI cannot control.
	// default: false
	WithCredentials bool `json:"withCredentials,omitempty"`

	// Function to set default values to each property in model. Accepts one argument modelPropertyMacro(property), property is immutable.
	// default: ""
	ModelPropertyMacro template.JS `json:"-"`

	// Function to set default value to parameters. Accepts two arguments parameterMacro(operation, parameter).
	// Operation and parameter are objects passed for context, both remain immutable.
	// default: ""
	ParameterMacro template.JS `json:"-"`

	// If set to true, it persists authorization data and it would not be lost on browser close/refresh.
	// default: false
	PersistAuthorization bool `json:"persistAuthorization,omitempty"`

	// Configuration information for OAuth2, optional if using OAuth2
	OAuth *OAuthConfig `json:"-"`

	// (authDefinitionKey, username, password) => action
	// Programmatically set values for a Basic authorization scheme.
	// default: ""
	PreauthorizeBasic template.JS `json:"-"`

	// (authDefinitionKey, apiKeyValue) => action
	// Programmatically set values for an API key or Bearer authorization scheme.
	// In case of OpenAPI 3.0 Bearer scheme, apiKeyValue must contain just the token itself without the Bearer prefix.
	// default: ""
	PreauthorizeApiKey template.JS `json:"-"`

	// Applies custom CSS styles.
	// default: ""
	CustomStyle template.CSS `json:"-"`

	// Applies custom JavaScript scripts.
	// default ""
	CustomScript template.JS `json:"-"`
}

type FilterConfig struct {
	Enabled    bool
	Expression string
}

func (fc FilterConfig) Value() interface{} {
	if fc.Expression != "" {
		return fc.Expression
	}
	return fc.Enabled
}

type SyntaxHighlightConfig struct {
	// Whether syntax highlighting should be activated or not.
	// default: true
	Activate bool `json:"activate"`
	// Highlight.js syntax coloring theme to use.
	// Possible values are ["agate", "arta", "monokai", "nord", "obsidian", "tomorrow-night"]
	// default: "agate"
	Theme string `json:"theme,omitempty"`
}

func (shc SyntaxHighlightConfig) Value() interface{} {
	if shc.Activate {
		return shc
	}
	return false
}

type OAuthConfig struct {
	// ID of the client sent to the OAuth2 provider.
	// default: ""
	ClientId string `json:"clientId,omitempty"`

	// Never use this parameter in your production environment.
	// It exposes cruicial security information. This feature is intended for dev/test environments only.
	// Secret of the client sent to the OAuth2 provider.
	// default: ""
	ClientSecret string `json:"clientSecret,omitempty"`

	// Application name, displayed in authorization popup.
	// default: ""
	AppName string `json:"appName,omitempty"`

	// Realm query parameter (for oauth1) added to authorizationUrl and tokenUrl.
	// default: ""
	Realm string `json:"realm,omitempty"`

	// String array of initially selected oauth scopes
	// default: nil
	Scopes []string `json:"scopes,omitempty"`

	// Additional query parameters added to authorizationUrl and tokenUrl.
	// default: nil
	AdditionalQueryStringParams map[string]string `json:"additionalQueryStringParams,omitempty"`

	// Unavailable	Only activated for the accessCode flow.
	// During the authorization_code request to the tokenUrl, pass the Client Password using the HTTP Basic Authentication scheme
	// (Authorization header with Basic base64encode(client_id + client_secret)).
	// default: false
	UseBasicAuthenticationWithAccessCodeGrant bool `json:"useBasicAuthenticationWithAccessCodeGrant,omitempty"`

	// Only applies to authorizatonCode flows.
	// Proof Key for Code Exchange brings enhanced security for OAuth public clients.
	// default: false
	UsePkceWithAuthorizationCodeGrant bool `json:"usePkceWithAuthorizationCodeGrant,omitempty"`
}

var DefaultUIConfig = SwaggerUIConfig{
	URL:    "/swagger/swagger.yaml",
	Title:  "Swagger UI",
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
	SyntaxHighlight: &SyntaxHighlightConfig{
		Activate: true,
		Theme:    "agate",
	},
	ShowMutatedRequest:     true,
	DisplayRequestDuration: true,
	PersistAuthorization:   true,
	RequestSnippetsEnabled: true,
	TryItOutEnabled:        true,
}

func swaggerUIConfigDefault(ui_config SwaggerUIConfig) SwaggerUIConfig {
	cfg := ui_config

	if cfg.URL == "" {
		cfg.URL = DefaultUIConfig.URL
	}

	if cfg.Title == "" {
		cfg.Title = DefaultUIConfig.Title
	}

	if cfg.Layout == "" {
		cfg.Layout = DefaultUIConfig.Layout
	}

	if cfg.DefaultModelRendering == "" {
		cfg.DefaultModelRendering = DefaultUIConfig.DefaultModelRendering
	}

	if cfg.DocExpansion == "" {
		cfg.DocExpansion = DefaultUIConfig.DocExpansion
	}

	if cfg.Plugins == nil {
		cfg.Plugins = DefaultUIConfig.Plugins
	}

	if cfg.Presets == nil {
		cfg.Presets = DefaultUIConfig.Presets
	}

	if cfg.SyntaxHighlight == nil {
		cfg.SyntaxHighlight = DefaultUIConfig.SyntaxHighlight
	}

	return cfg
}

const indexPageTmpl string = `
<!-- HTML for static distribution bundle build -->
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <title>{{.Title}}</title>
    <link href="https://fonts.googleapis.com/css?family=Open+Sans:400,700|Source+Code+Pro:300,600|Titillium+Web:400,600,700" rel="stylesheet">
  	<link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5.20.0/swagger-ui.css" />
    <link rel="icon" type="image/png" href="./favicon-32x32.png" sizes="32x32" />
    <link rel="icon" type="image/png" href="./favicon-16x16.png" sizes="16x16" />
    {{- if .CustomStyle}}
      <style>
        body { margin: 0; }
        {{.CustomStyle}}
      </style>
    {{- end}}
    {{- if .CustomScript}}
      <script>
        {{.CustomScript}}
      </script>
    {{- end}}
  </head>
  <body>
    <svg xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" style="position:absolute;width:0;height:0">
      <defs>
        <symbol viewBox="0 0 20 20" id="unlocked">
              <path d="M15.8 8H14V5.6C14 2.703 12.665 1 10 1 7.334 1 6 2.703 6 5.6V6h2v-.801C8 3.754 8.797 3 10 3c1.203 0 2 .754 2 2.199V8H4c-.553 0-1 .646-1 1.199V17c0 .549.428 1.139.951 1.307l1.197.387C5.672 18.861 6.55 19 7.1 19h5.8c.549 0 1.428-.139 1.951-.307l1.196-.387c.524-.167.953-.757.953-1.306V9.199C17 8.646 16.352 8 15.8 8z"></path>
        </symbol>
        <symbol viewBox="0 0 20 20" id="locked">
          <path d="M15.8 8H14V5.6C14 2.703 12.665 1 10 1 7.334 1 6 2.703 6 5.6V8H4c-.553 0-1 .646-1 1.199V17c0 .549.428 1.139.951 1.307l1.197.387C5.672 18.861 6.55 19 7.1 19h5.8c.549 0 1.428-.139 1.951-.307l1.196-.387c.524-.167.953-.757.953-1.306V9.199C17 8.646 16.352 8 15.8 8zM12 8H8V5.199C8 3.754 8.797 3 10 3c1.203 0 2 .754 2 2.199V8z"/>
        </symbol>
        <symbol viewBox="0 0 20 20" id="close">
          <path d="M14.348 14.849c-.469.469-1.229.469-1.697 0L10 11.819l-2.651 3.029c-.469.469-1.229.469-1.697 0-.469-.469-.469-1.229 0-1.697l2.758-3.15-2.759-3.152c-.469-.469-.469-1.228 0-1.697.469-.469 1.228-.469 1.697 0L10 8.183l2.651-3.031c.469-.469 1.228-.469 1.697 0 .469.469.469 1.229 0 1.697l-2.758 3.152 2.758 3.15c.469.469.469 1.229 0 1.698z"/>
        </symbol>
        <symbol viewBox="0 0 20 20" id="large-arrow">
          <path d="M13.25 10L6.109 2.58c-.268-.27-.268-.707 0-.979.268-.27.701-.27.969 0l7.83 7.908c.268.271.268.709 0 .979l-7.83 7.908c-.268.271-.701.27-.969 0-.268-.269-.268-.707 0-.979L13.25 10z"/>
        </symbol>
        <symbol viewBox="0 0 20 20" id="large-arrow-down">
          <path d="M17.418 6.109c.272-.268.709-.268.979 0s.271.701 0 .969l-7.908 7.83c-.27.268-.707.268-.979 0l-7.908-7.83c-.27-.268-.27-.701 0-.969.271-.268.709-.268.979 0L10 13.25l7.418-7.141z"/>
        </symbol>
        <symbol viewBox="0 0 24 24" id="jump-to">
          <path d="M19 7v4H5.83l3.58-3.59L8 6l-6 6 6 6 1.41-1.41L5.83 13H21V7z"/>
        </symbol>
        <symbol viewBox="0 0 24 24" id="expand">
          <path d="M10 18h4v-2h-4v2zM3 6v2h18V6H3zm3 7h12v-2H6v2z"/>
        </symbol>
      </defs>
    </svg>
    <div id="swagger-ui"></div>
 	<script src="https://unpkg.com/swagger-ui-dist@5.20.0/swagger-ui-bundle.js" crossorigin></script>
 	<script src="https://unpkg.com/swagger-ui-dist@5.20.0/swagger-ui-standalone-preset.js" crossorigin></script>
    <script>
    window.onload = function() {
      config = {{.}};
      config.dom_id = '#swagger-ui';
      config.plugins = [
        {{- range $plugin := .Plugins }}
          {{$plugin}},
        {{- end}}
      ];
      config.presets = [
        {{- range $preset := .Presets }}
          {{$preset}},
        {{- end}}
      ];
      config.filter = {{.Filter.Value}}
      config.syntaxHighlight = {{.SyntaxHighlight.Value}}
      {{if .TagsSorter}}
        config.tagsSorter = {{.TagsSorter}}
      {{end}}
      {{if .OnComplete}}
        config.onComplete = {{.OnComplete}}
      {{end}}
      {{if .RequestInterceptor}}
        config.requestInterceptor = {{.RequestInterceptor}}
      {{end}}
      {{if .ResponseInterceptor}}
        config.responseInterceptor = {{.ResponseInterceptor}}
      {{end}}
      {{if .ModelPropertyMacro}}
        config.modelPropertyMacro = {{.ModelPropertyMacro}}
      {{end}}
      {{if .ParameterMacro}}
        config.parameterMacro = {{.ParameterMacro}}
      {{end}}

      const ui = SwaggerUIBundle(config);

      {{if .OAuth}}
        ui.initOAuth({{.OAuth}});
      {{end}}
      {{if .PreauthorizeBasic}}
        ui.preauthorizeBasic({{.PreauthorizeBasic}});
      {{end}}
      {{if .PreauthorizeApiKey}}
        ui.preauthorizeApiKey({{.PreauthorizeApiKey}});
      {{end}}

      window.ui = ui
    }
    </script>
  </body>
</html>
`
