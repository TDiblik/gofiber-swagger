package gofiberswagger

import (
	"database/sql"
	"mime/multipart"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type BasicTypes struct {
	Int     int
	Int8    int8
	Int16   int16
	Int32   int32
	Int64   int64
	Uint    uint
	Uint8   uint8
	Uint16  uint16
	Uint32  uint32
	Uint64  uint64
	Float32 float32
	Float64 float64
	Bool    bool
	String  string
}

func TestSchema_BasicTypes(t *testing.T) {
	t.Parallel()

	schema := CreateSchema[BasicTypes]()
	assert.NotNil(t, schema)
	assert.NotNil(t, schema.Value)
	assert.Equal(t, "object", (*schema.Value.Type)[0])
	assert.Equal(t, "BasicTypes", schema.Value.Title)
	assert.Len(t, schema.Value.Properties, 14)

	// Int
	intSchema := schema.Value.Properties["Int"]
	assert.NotNil(t, intSchema)
	assert.Equal(t, "integer", (*intSchema.Value.Type)[0])

	// String
	stringSchema := schema.Value.Properties["String"]
	assert.NotNil(t, stringSchema)
	assert.Equal(t, "string", (*stringSchema.Value.Type)[0])

	// Bool
	boolSchema := schema.Value.Properties["Bool"]
	assert.NotNil(t, boolSchema)
	assert.Equal(t, "boolean", (*boolSchema.Value.Type)[0])
}

type ComplexTypes struct {
	Time        time.Time
	UUID        uuid.UUID
	File        multipart.FileHeader
	NullString  sql.NullString
	NullInt64   sql.NullInt64
	NullFloat64 sql.NullFloat64
	NullBool    sql.NullBool
	NullTime    sql.NullTime
	Bytes       []byte
	Map         map[string]any
	Slice       []string
	Array       [2]string
	Struct      BasicTypes
	Pointer     *BasicTypes
}

func TestSchema_ComplexTypes(t *testing.T) {
	t.Parallel()

	schema := CreateSchema[ComplexTypes]()
	assert.NotNil(t, schema)
	assert.NotNil(t, schema.Value)
	assert.Equal(t, "object", (*schema.Value.Type)[0])
	assert.Equal(t, "ComplexTypes", schema.Value.Title)
	assert.Len(t, schema.Value.Properties, 14)

	// Time
	timeSchema := schema.Value.Properties["Time"]
	assert.NotNil(t, timeSchema)
	assert.Equal(t, "string", (*timeSchema.Value.Type)[0])
	assert.Equal(t, "date-time", timeSchema.Value.Format)

	// UUID
	uuidSchema := schema.Value.Properties["UUID"]
	assert.NotNil(t, uuidSchema)
	assert.Equal(t, "string", (*uuidSchema.Value.Type)[0])
	assert.Equal(t, "uuid", uuidSchema.Value.Format)

	// File
	fileSchema := schema.Value.Properties["File"]
	assert.NotNil(t, fileSchema)
	assert.Equal(t, "string", (*fileSchema.Value.Type)[0])
	assert.Equal(t, "binary", fileSchema.Value.Format)

	// Slice
	sliceSchema := schema.Value.Properties["Slice"]
	assert.NotNil(t, sliceSchema)
	assert.Equal(t, "array", (*sliceSchema.Value.Type)[0])
	assert.NotNil(t, sliceSchema.Value.Items)
	assert.Equal(t, "string", (*sliceSchema.Value.Items.Value.Type)[0])

	// Struct
	structSchema := schema.Value.Properties["Struct"]
	assert.NotNil(t, structSchema)
	assert.NotNil(t, structSchema.Ref)
}

type TestEnum string

const (
	TestEnumA TestEnum = "A"
	TestEnumB TestEnum = "B"
)

func (TestEnum) EnumValues() []any {
	return []any{TestEnumA, TestEnumB}
}

type WithEnums struct {
	Enum TestEnum
}

func TestSchema_WithEnums(t *testing.T) {
	t.Parallel()

	schema := CreateSchema[WithEnums]()
	assert.NotNil(t, schema)
	assert.NotNil(t, schema.Value)
	assert.Equal(t, "object", (*schema.Value.Type)[0])

	enumSchema := schema.Value.Properties["Enum"]
	assert.NotNil(t, enumSchema)
	assert.NotNil(t, enumSchema.Value.Enum)
	assert.Len(t, enumSchema.Value.Enum, 2)
	assert.Equal(t, TestEnum("A"), enumSchema.Value.Enum[0])
	assert.Equal(t, TestEnum("B"), enumSchema.Value.Enum[1])
}

type WithTags struct {
	Required    string `json:"required_field" validate:"required"`
	OmitEmpty   string `json:"omitempty_field,omitempty"`
	MinMax      int    `validate:"min=1,max=10"`
	MinMaxStr   string `validate:"min=1,max=10"`
	MinMaxSlice []int  `validate:"min=1,max=10"`
	Ignored     string `json:"-"`
	IgnoredXML  string `xml:"-"`
	XMLAttr     string `xml:"xml_attr,attr"`
	OneOf       string `validate:"oneof=A B"`
}

func TestSchema_WithTags(t *testing.T) {
	t.Parallel()

	schema := CreateSchema[WithTags]()
	assert.NotNil(t, schema)
	assert.NotNil(t, schema.Value)
	assert.Len(t, schema.Value.Properties, 7)
	assert.Len(t, schema.Value.Required, 1)
	assert.Equal(t, "required_field", schema.Value.Required[0])

	// Required
	requiredSchema := schema.Value.Properties["required_field"]
	assert.NotNil(t, requiredSchema)
	assert.False(t, requiredSchema.Value.Nullable)

	// OmitEmpty
	omitemptySchema := schema.Value.Properties["omitempty_field"]
	assert.NotNil(t, omitemptySchema)
	assert.True(t, omitemptySchema.Value.Nullable)

	// MinMax
	minmaxSchema := schema.Value.Properties["MinMax"]
	assert.NotNil(t, minmaxSchema)
	assert.Equal(t, float64(1), *minmaxSchema.Value.Min)
	assert.Equal(t, float64(10), *minmaxSchema.Value.Max)

	// MinMaxStr
	minmaxStrSchema := schema.Value.Properties["MinMaxStr"]
	assert.NotNil(t, minmaxStrSchema)
	assert.Equal(t, uint64(1), minmaxStrSchema.Value.MinLength)
	assert.Equal(t, uint64(10), *minmaxStrSchema.Value.MaxLength)

	// MinMaxSlice
	minmaxSliceSchema := schema.Value.Properties["MinMaxSlice"]
	assert.NotNil(t, minmaxSliceSchema)
	assert.Equal(t, uint64(1), minmaxSliceSchema.Value.MinItems)
	assert.Equal(t, uint64(10), *minmaxSliceSchema.Value.MaxItems)

	// OneOf
	oneofSchema := schema.Value.Properties["OneOf"]
	assert.NotNil(t, oneofSchema)
	assert.Len(t, oneofSchema.Value.Enum, 2)
	assert.Equal(t, "A", oneofSchema.Value.Enum[0])
	assert.Equal(t, "B", oneofSchema.Value.Enum[1])
}
