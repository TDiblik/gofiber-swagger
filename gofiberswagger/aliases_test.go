package gofiberswagger

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	A string
	B int
}

func TestNewRequestBody(t *testing.T) {
	t.Parallel()

	t.Run("JSON", func(t *testing.T) {
		t.Parallel()
		reqBody := NewRequestBodyJSON[TestStruct]()
		assert.NotNil(t, reqBody)
		assert.NotNil(t, reqBody.Value)
		assert.NotNil(t, reqBody.Value.Content)
		assert.Contains(t, reqBody.Value.Content, "application/json")
	})

	t.Run("FormData", func(t *testing.T) {
		t.Parallel()
		reqBody := NewRequestBodyFormData[TestStruct]()
		assert.NotNil(t, reqBody)
		assert.NotNil(t, reqBody.Value)
		assert.NotNil(t, reqBody.Value.Content)
		assert.Contains(t, reqBody.Value.Content, "multipart/form-data")
	})

	t.Run("FormUrlEncoded", func(t *testing.T) {
		t.Parallel()
		reqBody := NewRequestBodyFormUrlEncodedData[TestStruct]()
		assert.NotNil(t, reqBody)
		assert.NotNil(t, reqBody.Value)
		assert.NotNil(t, reqBody.Value.Content)
		assert.Contains(t, reqBody.Value.Content, "application/x-www-form-urlencoded")
	})

	t.Run("XML", func(t *testing.T) {
		t.Parallel()
		reqBody := NewRequestBodyXML[TestStruct]()
		assert.NotNil(t, reqBody)
		assert.NotNil(t, reqBody.Value)
		assert.NotNil(t, reqBody.Value.Content)
		assert.Contains(t, reqBody.Value.Content, "application/xml")
	})
}

func TestNewResponses(t *testing.T) {
	t.Parallel()

	responses := NewResponses(
		NewResponseInfo[TestStruct](strconv.Itoa(http.StatusCreated), "Created"),
		NewResponseInfo[string](strconv.Itoa(http.StatusNotFound), "Not Found"),
	)
	assert.NotNil(t, responses)
	assert.NotNil(t, responses.Value(strconv.Itoa(http.StatusCreated)))
	assert.NotNil(t, responses.Value(strconv.Itoa(http.StatusNotFound)))
}

func TestNewParameters(t *testing.T) {
	t.Parallel()

	params := NewParameters(
		NewPathParameter("id"),
		NewQueryParameter("name"),
		NewHeaderParameter("X-Request-ID"),
		NewCookieParameter("session"),
	)

	assert.NotNil(t, params)
	assert.Len(t, params, 4)

	assert.Equal(t, "id", params[0].Value.Name)
	assert.Equal(t, "path", params[0].Value.In)

	assert.Equal(t, "name", params[1].Value.Name)
	assert.Equal(t, "query", params[1].Value.In)

	assert.Equal(t, "X-Request-ID", params[2].Value.Name)
	assert.Equal(t, "header", params[2].Value.In)

	assert.Equal(t, "session", params[3].Value.Name)
	assert.Equal(t, "cookie", params[3].Value.In)
}

func TestSchemaHelpers(t *testing.T) {
	t.Parallel()

	t.Run("Bool", func(t *testing.T) {
		t.Parallel()
		s := NewBoolSchema()
		assert.Equal(t, "boolean", (*s.Type)[0])
	})

	t.Run("String", func(t *testing.T) {
		t.Parallel()
		s := NewStringSchema()
		assert.Equal(t, "string", (*s.Type)[0])
	})

	t.Run("Int32", func(t *testing.T) {
		t.Parallel()
		s := NewInt32Schema()
		assert.Equal(t, "integer", (*s.Type)[0])
		assert.Equal(t, "int32", s.Format)
	})
}
