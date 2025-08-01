package gofiberswagger

import (
	"math/rand/v2"
	"reflect"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

var acquiredSchemas map[string]*SchemaRef

func setToAcquiredSchemas(ref string, schema *SchemaRef) {
	if acquiredSchemas == nil {
		acquiredSchemas = make(map[string]*SchemaRef)
	}
	if schema != nil {
		acquiredSchemas[ref] = schema
	}
}
func getFromAcquiredSchemas(ref string) *SchemaRef {
	if acquiredSchemas == nil {
		return nil
	}

	return acquiredSchemas[ref]
}

func CreateSchema[T any]() *SchemaRef {
	var t T
	return generateSchema(reflect.TypeOf(t))
}

func boolPtr(b bool) *bool {
	return &b
}

func resolveGenericType(fieldType reflect.Type, parentType reflect.Type) reflect.Type {
	// If your fieldType is interface{} or any, and your parent is a known generic wrapper:
	if fieldType.Kind() == reflect.Interface {
		if concrete, ok := genericMapping[parentType]; ok {
			return concrete
		}
	}
	return nil
}

var genericMapping = map[reflect.Type]reflect.Type{} // or sync.Map

// Called during NewResponseInfo[T]
func recordGenericType(wrapperT reflect.Type, itemT reflect.Type) {
	genericMapping[wrapperT] = itemT
}

func generateSchema(t reflect.Type) *SchemaRef {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	// Handle slices (e.g. Response[[]User])
	if t.Kind() == reflect.Slice {
		return &SchemaRef{Value: &Schema{
			Type:  &Types{"array"},
			Items: generateSchema(t.Elem()),
		}}
	}

	// Handle map[string]X
	if t.Kind() == reflect.Map && t.Key().Kind() == reflect.String {
		return &SchemaRef{Value: &Schema{
			Type: &Types{"object"},
			AdditionalProperties: AdditionalProperties{
				Has:    boolPtr(true),
				Schema: generateSchema(t.Elem()),
			},
		}}
	}

	tName := t.Name()
	if tName == "" {
		var genPartOfName string
		if genPart, err := uuid.NewUUID(); err == nil {
			genPartOfName = genPart.String()
		} else {
			genPartOfName = strconv.Itoa(rand.Int())
		}
		tName = "generated-" + genPartOfName
	}

	ref := strings.ReplaceAll(strings.ReplaceAll(t.PkgPath(), "/", "_"), ".", "_") + tName
	ref_path := "#/components/schemas/" + ref
	possibleSchema := getFromAcquiredSchemas(ref)
	if possibleSchema != nil {
		if t.Kind() == reflect.Struct {
			return &SchemaRef{
				Ref:        ref_path,
				Extensions: possibleSchema.Extensions,
				Origin:     possibleSchema.Origin,
				Value:      possibleSchema.Value,
			}
		}
		return possibleSchema
	}

	schema := getDefaultSchema(t)

	// Handle empty struct{}
	if t.Kind() == reflect.Struct && t.NumField() == 0 {
		schema.Type = &Types{"object"}
		return &SchemaRef{
			Value: schema,
		}
	}

	if t.Kind() == reflect.Struct {
		schema.Title = tName
		schema.Type = &Types{"object"}

		// set placeholder that will get overwritten to prevent recursion
		setToAcquiredSchemas(ref, &SchemaRef{Value: &Schema{}})

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)

			jsonTag, jsonTagExists := field.Tag.Lookup("json")
			if jsonTag == "-" {
				continue
			}

			xmlTag, xmlTagExists := field.Tag.Lookup("xml")
			if xmlTag == "-" || (xmlTagExists && field.Name == "XMLName") {
				continue
			}

			isNullable := false
			fieldType := field.Type
			for fieldType.Kind() == reflect.Pointer {
				fieldType = fieldType.Elem()
				isNullable = true
			}
			fieldTypeName := fieldType.Name()
			fieldTypePkgPath := fieldType.PkgPath()
			fieldKind := fieldType.Kind()

			// create schema for the field
			var result *SchemaRef = nil

			// Handle interface{} or generics
			if fieldKind == reflect.Interface {
				if concrete := resolveGenericType(fieldType, t); concrete != nil {
					result = generateSchema(concrete)
				} else {
					result = &SchemaRef{Value: &Schema{Type: &Types{"object"}}}
				}
			} else {
				switch {
				// skip channels and functions
				case fieldKind == reflect.Func, fieldKind == reflect.Chan:
					continue

				// handle time.Time type
				case fieldKind == reflect.Struct && fieldType == timeType:
					result = &SchemaRef{Value: &Schema{
						Type:   &Types{"string"},
						Format: "date-time",
					}}

				// handle file uploads
				case fieldKind == reflect.Struct && fieldTypeName == "FileHeader" && fieldTypePkgPath == "mime/multipart":
					result = &SchemaRef{Value: &Schema{
						Type:   &Types{"string"},
						Format: "binary",
					}}

				// handle uuid.UUID
				case fieldKind == reflect.Array && fieldTypeName == "UUID" && fieldType.Elem().Kind() == reflect.Uint8:
					result = &SchemaRef{Value: &Schema{
						Type:   &Types{"string"},
						Format: "uuid",
					}}

				// handle uuid.NullUUID and it's alias wrappers
				case fieldKind == reflect.Struct && (isNullType(fieldType, "NullUUID", "UUID") || isNullTypeWrapper(fieldType, "NullUUID", "UUID")):
					isNullable = true
					result = &SchemaRef{Value: &Schema{
						Type:   &Types{"string"},
						Format: "uuid",
					}}

				// handle sql.NullBool and it's alias wrappers
				case fieldKind == reflect.Struct && (isNullType(fieldType, "NullBool", "Bool") || isNullTypeWrapper(fieldType, "NullBool", "Bool")):
					isNullable = true
					result = &SchemaRef{Value: &Schema{
						Type: &Types{"boolean"},
					}}

				// handle sql.NullByte and it's alias wrappers
				case fieldKind == reflect.Struct && (isNullType(fieldType, "NullByte", "Byte") || isNullTypeWrapper(fieldType, "NullByte", "Byte")):
					isNullable = true
					result = &SchemaRef{Value: &Schema{
						Type:   &Types{"string"},
						Format: "byte",
					}}

				// handle sql.NullInt16 and it's alias wrappers
				case fieldKind == reflect.Struct && (isNullType(fieldType, "NullInt16", "Int16") || isNullTypeWrapper(fieldType, "NullInt16", "Int16")):
					isNullable = true
					result = &SchemaRef{Value: &Schema{
						Type:         &Types{"integer"},
						Min:          &minInt16,
						Max:          &maxInt16,
						ExclusiveMin: false,
						ExclusiveMax: false,
					}}

				// handle sql.NullInt32 and it's alias wrappers
				case fieldKind == reflect.Struct && (isNullType(fieldType, "NullInt32", "Int32") || isNullTypeWrapper(fieldType, "NullInt32", "Int32")):
					isNullable = true
					result = &SchemaRef{Value: &Schema{
						Type:         &Types{"integer"},
						Format:       "int32",
						Min:          &minInt32,
						Max:          &maxInt32,
						ExclusiveMin: false,
						ExclusiveMax: false,
					}}

				// handle sql.NullInt64 and it's alias wrappers
				case fieldKind == reflect.Struct && (isNullType(fieldType, "NullInt64", "Int64") || isNullTypeWrapper(fieldType, "NullInt64", "Int64")):
					isNullable = true
					result = &SchemaRef{Value: &Schema{
						Type:         &Types{"integer"},
						Format:       "int64",
						Min:          &minInt64,
						Max:          &maxInt64,
						ExclusiveMin: false,
						ExclusiveMax: false,
					}}

				// handle sql.NullFloat64 and it's alias wrappers
				case fieldKind == reflect.Struct && (isNullType(fieldType, "NullFloat64", "Float64") || isNullTypeWrapper(fieldType, "NullFloat64", "Float64")):
					isNullable = true
					result = &SchemaRef{Value: &Schema{
						Type:         &Types{"number"},
						Format:       "double",
						Min:          &minInt64,
						Max:          &maxInt64,
						ExclusiveMin: false,
						ExclusiveMax: false,
					}}

				// handle sql.NullTime and it's alias wrappers
				case fieldKind == reflect.Struct && (isNullType(fieldType, "NullTime", "Time") || isNullTypeWrapper(fieldType, "NullTime", "Time")):
					isNullable = true
					result = &SchemaRef{Value: &Schema{
						Type:   &Types{"string"},
						Format: "date-time",
					}}

				// handle sql.NullString and it's alias wrappers
				case fieldKind == reflect.Struct && (isNullType(fieldType, "NullString", "String") || isNullTypeWrapper(fieldType, "NullString", "String")):
					isNullable = true
					result = &SchemaRef{Value: &Schema{
						Type: &Types{"string"},
					}}

				// handle bytes
				case fieldKind == reflect.Slice && fieldType.Elem().Kind() == reflect.Uint8:
					if fieldType == rawMessageType {
						result = &SchemaRef{Value: &Schema{}}
					} else {
						result = &SchemaRef{Value: &Schema{
							Type:   &Types{"string"},
							Format: "byte",
						}}
					}

				// handle map[string]object
				case fieldKind == reflect.Map && fieldType.Key().Kind() == reflect.String:
					valueSchema := generateSchema(fieldType.Elem())
					has := true
					result = &SchemaRef{Value: &Schema{
						Type: &Types{"object"},
						AdditionalProperties: AdditionalProperties{
							Has:    &has,
							Schema: valueSchema,
						},
					}}

				// handle general structs
				case fieldKind == reflect.Struct:
					result = generateSchema(fieldType)

				// handle general slices / arrays
				case fieldKind == reflect.Slice, fieldKind == reflect.Array:
					result = &SchemaRef{Value: &Schema{
						Type:  &Types{"array"},
						Items: generateSchema(fieldType.Elem()),
					}}

				// handle general maps / interface{} / any
				case fieldKind == reflect.Map || fieldKind == reflect.Interface:
					result = &SchemaRef{Value: &Schema{
						Type: &Types{"object"},
					}}

				// generated default schema for non-special types (string/int/etc)
				default:
					result = &SchemaRef{
						Value: getDefaultSchema(fieldType),
					}
				}
			}

			if result != nil && result.Value != nil {
				result.Value.Nullable = isNullable
			}

			// handle json tag
			fieldName := field.Name
			jsonTagOptions := strings.Split(jsonTag, ",")
			if jsonTagExists && len(jsonTagOptions) > 0 && jsonTagOptions[0] != "" {
				fieldName = jsonTagOptions[0]
			}
			for i := 1; i < len(jsonTagOptions); i++ {
				option := jsonTagOptions[i]
				switch option {
				case "string":
					result.Value.Type = &Types{"string"}
				case "omitempty":
					result.Value.Nullable = true
					result.Value.Description += " omitempty "
				case "omitzero":
					result.Value.Nullable = true
					result.Value.Description += " omitzero "
				}
			}

			// handle xml tag
			xmlTagOptions := strings.Split(xmlTag, ",")
			if xmlTagExists && len(xmlTagOptions) > 0 && result.Value.XML == nil {
				result.Value.XML = &XML{}
			}
			if xmlTagExists && len(xmlTagOptions) > 0 && xmlTagOptions[0] != "" {
				result.Value.XML.Name = xmlTagOptions[0]
			}
			for i := 1; i < len(xmlTagOptions); i++ {
				option := xmlTagOptions[i]
				switch option {
				case "attr":
					result.Value.XML.Attribute = true
				case "chardata", "cdata", "innerxml", "comment":
					result.Value.Description += " " + option + " "
				case "omitempty":
					result.Value.Nullable = true
					result.Value.Description += " omitempty "
				}
				// todo: handle `name>first` / `a>b>c` syntax
			}

			// handle enum values
			if implementsSwaggerEnum(fieldType) {
				handleEnumValues(result, getSwaggerEnumValues(fieldType), false, fieldType)
			}

			// handle validate tag
			validateTag := field.Tag.Get("validate")
			validateTagOptions := strings.Split(validateTag, ",")
			for _, validation := range validateTagOptions {
				switch {
				case validation == "required":
					schema.Required = append(schema.Required, fieldName)
					result.Value.Nullable = false
					result.Value.AllowEmptyValue = false
				case strings.HasPrefix(validation, "min=") && (fieldKind == reflect.Slice || fieldKind == reflect.Array):
					if minValue, err := strconv.ParseUint(strings.TrimPrefix(validation, "min="), 10, 64); err == nil {
						result.Value.MinItems = minValue
					}
				case strings.HasPrefix(validation, "min=") && fieldKind == reflect.String:
					if minValue, err := strconv.ParseUint(strings.TrimPrefix(validation, "min="), 10, 64); err == nil {
						result.Value.MinLength = minValue
					}
				case strings.HasPrefix(validation, "min="):
					if minValue, err := strconv.ParseFloat(strings.TrimPrefix(validation, "min="), 64); err == nil {
						result.Value.Min = &minValue
						result.Value.Default = minValue
					}
				case strings.HasPrefix(validation, "max=") && (fieldKind == reflect.Slice || fieldKind == reflect.Array):
					if maxValue, err := strconv.ParseUint(strings.TrimPrefix(validation, "max="), 10, 64); err == nil {
						result.Value.MaxItems = &maxValue
					}
				case strings.HasPrefix(validation, "max=") && fieldKind == reflect.String:
					if maxValue, err := strconv.ParseUint(strings.TrimPrefix(validation, "max="), 10, 64); err == nil {
						result.Value.MaxLength = &maxValue
					}
				case strings.HasPrefix(validation, "max="):
					if maxValue, err := strconv.ParseFloat(strings.TrimPrefix(validation, "max="), 64); err == nil {
						result.Value.Max = &maxValue
					}
				case strings.HasPrefix(validation, "minLength="):
					if minLen, err := strconv.ParseUint(strings.TrimPrefix(validation, "minLength="), 10, 64); err == nil {
						result.Value.MinLength = minLen
					}
				case strings.HasPrefix(validation, "maxLength="):
					if maxLen, err := strconv.ParseUint(strings.TrimPrefix(validation, "maxLength="), 10, 64); err == nil {
						result.Value.MaxLength = &maxLen
					}
				case strings.HasPrefix(validation, "uniqueItems"):
					result.Value.UniqueItems = true
				case strings.HasPrefix(validation, "omitnil"):
					result.Value.Description += " omitnil "
				case strings.HasPrefix(validation, "oneof="):
					options := []any{}
					stringOptions := strings.Split(strings.TrimPrefix(validation, "oneof="), " ")
					for _, option := range stringOptions {
						options = append(options, option)
					}
					handleEnumValues(result, options, true, fieldType)
				}
			}
			result.Value.Title = fieldName
			result.Value.Description = strings.ReplaceAll(result.Value.Description, "  ", "")

			schema.Properties[fieldName] = result
		}

		setToAcquiredSchemas(ref, &SchemaRef{
			Value: schema,
		})
		return &SchemaRef{
			Ref:   ref_path,
			Value: schema,
		}
	}

	return &SchemaRef{
		Value: schema,
	}
}

func getDefaultSchema(t reflect.Type) *Schema {
	schema := Schema{
		Properties: make(Schemas),
		Required:   []string{},
	}
	switch t.Kind() {
	case reflect.Bool:
		schema.Type = &Types{"boolean"}
		schema.Default = false

	case reflect.Int:
		schema.Type = &Types{"integer"}
		schema.Min = &minInt
		schema.Max = &maxInt
		schema.ExclusiveMin = false
		schema.ExclusiveMax = false
		schema.Default = 0
	case reflect.Int8:
		schema.Type = &Types{"integer"}
		schema.Min = &minInt8
		schema.Max = &maxInt8
		schema.ExclusiveMin = false
		schema.ExclusiveMax = false
		schema.Default = 0
	case reflect.Int16:
		schema.Type = &Types{"integer"}
		schema.Min = &minInt16
		schema.Max = &maxInt16
		schema.ExclusiveMin = false
		schema.ExclusiveMax = false
		schema.Default = 0
	case reflect.Int32:
		schema.Type = &Types{"integer"}
		schema.Format = "int32"
		schema.Min = &minInt32
		schema.Max = &maxInt32
		schema.ExclusiveMin = false
		schema.ExclusiveMax = false
		schema.Default = 0
	case reflect.Int64:
		schema.Type = &Types{"integer"}
		schema.Format = "int64"
		schema.Min = &minInt64
		schema.Max = &maxInt64
		schema.ExclusiveMin = false
		schema.ExclusiveMax = false
		schema.Default = 0
	case reflect.Uint:
		schema.Type = &Types{"integer"}
		schema.Min = &zeroInt
		schema.Max = &maxUint
		schema.ExclusiveMin = false
		schema.ExclusiveMax = false
		schema.Default = 0
	case reflect.Uint8:
		schema.Type = &Types{"integer"}
		schema.Min = &zeroInt
		schema.Max = &maxUint8
		schema.ExclusiveMin = false
		schema.ExclusiveMax = false
		schema.Default = 0
	case reflect.Uint16:
		schema.Type = &Types{"integer"}
		schema.Min = &zeroInt
		schema.Max = &maxUint16
		schema.ExclusiveMin = false
		schema.ExclusiveMax = false
		schema.Default = 0
	case reflect.Uint32:
		schema.Type = &Types{"integer"}
		schema.Min = &zeroInt
		schema.Max = &maxUint32
		schema.ExclusiveMin = false
		schema.ExclusiveMax = false
		schema.Default = 0
	case reflect.Uint64:
		schema.Type = &Types{"integer"}
		schema.Min = &zeroInt
		schema.Max = &maxUint64
		schema.ExclusiveMin = false
		schema.ExclusiveMax = false
		schema.Default = 0

	case reflect.Float32:
		schema.Type = &Types{"number"}
		schema.Format = "float"
		schema.Min = &minFloat32
		schema.Max = &maxFloat32
		schema.ExclusiveMin = false
		schema.ExclusiveMax = false
		schema.Default = 0
	case reflect.Float64:
		schema.Type = &Types{"number"}
		schema.Format = "double"
		schema.Min = &minFloat64
		schema.Max = &maxFloat64
		schema.ExclusiveMin = false
		schema.ExclusiveMax = false
		schema.Default = 0

	case reflect.String:
		schema.Type = &Types{"string"}
		schema.Default = ""

	case reflect.Array:
		if t.Name() == "UUID" && t.Elem().Kind() == reflect.Uint8 {
			schema.Type = &Types{"string"}
			schema.Format = "uuid"
		}
	}

	return &schema
}

// matches cases:
//
//	type SomeStruct struct{
//		SomeValue sql.Null* <---- this part
//	}
func isNullType(fieldType reflect.Type, nullFieldName string, uniqueFieldName string) bool {
	if fieldType.Kind() != reflect.Struct || fieldType.Name() != nullFieldName {
		return false
	}
	_, has_valid := fieldType.FieldByName("Valid")
	if !has_valid {
		return false
	}
	_, ok_unique := fieldType.FieldByName(uniqueFieldName)
	return ok_unique
}

// matches cases:
//
//	type SQLNull* struct {
//		sql.Null*
//	}
//
//	type SomeStruct struct {
//		SomeValue SQLNull* <---- this part
//	}
func isNullTypeWrapper(fieldType reflect.Type, nullFieldName string, uniqueFieldName string) bool {
	if fieldType.Kind() != reflect.Struct || fieldType.NumField() != 1 {
		return false
	}
	possible_null_type_field := fieldType.Field(0)
	if possible_null_type_field.Name != nullFieldName {
		return false
	}

	return isNullType(possible_null_type_field.Type, nullFieldName, uniqueFieldName)
}

// modifies the `result *SchemaRef`
func handleEnumValues(result *SchemaRef, options []any, overwrite bool, fieldType reflect.Type) {
	if result.Value.OneOf == nil || overwrite {
		result.Value.OneOf = []*SchemaRef{}
	}
	if result.Value.Enum == nil || overwrite {
		result.Value.Enum = []any{}
	}
	for _, option := range options {
		option_schema := generateSchema(fieldType)
		option_schema.Value.Default = option
		result.Value.OneOf = append(result.Value.OneOf, option_schema)
		result.Value.Enum = append(result.Value.Enum, option)
	}
	result.Value.Default = nil
}
