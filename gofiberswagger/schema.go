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
	return generateSchema(reflect.TypeOf(t), false)
}

func generateSchema(t reflect.Type, stopRecursion bool) *SchemaRef {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
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
	if schema.Type != nil && (*schema.Type)[0] != "object" {
		return &SchemaRef{Value: schema}
	}

	// Handle empty struct{}
	if t.Kind() == reflect.Struct && t.NumField() == 0 {
		schema.Type = &Types{"object"}
		return &SchemaRef{
			Value: schema,
		}
	}

	// handle singular enums
	if !stopRecursion && implementsSwaggerEnum(t) {
		options := getSwaggerEnumValues(t)
		enumSchema := &SchemaRef{
			Value: getDefaultSchema(t),
		}
		handleEnumValues(enumSchema, options, false, t)
		return enumSchema
	}

	if t.Kind() == reflect.Struct {
		schema.Title = tName
		schema.Type = &Types{"object"}

		// set placeholder that will get overwritten to prevent recursion
		setToAcquiredSchemas(ref, &SchemaRef{Value: &Schema{}})

		for i := range t.NumField() {
			field := t.Field(i)

			if field.Anonymous {
				fieldType := field.Type
				for fieldType.Kind() == reflect.Pointer {
					fieldType = fieldType.Elem()
				}

				if fieldType.Kind() == reflect.Struct {
					subSchema := generateSchema(fieldType, false)
					if subSchema != nil && subSchema.Value != nil {
						for name, prop := range subSchema.Value.Properties {
							if _, ok := schema.Properties[name]; !ok {
								schema.Properties[name] = prop
							}
						}
					}
					continue
				}
			}

			jsonTag, jsonTagExists := field.Tag.Lookup("json")
			formTag, formTagExists := field.Tag.Lookup("form")
			queryTag, queryTagExists := field.Tag.Lookup("query")

			if jsonTag == "-" && !formTagExists && !queryTagExists {
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

			// create schema for the field. First handle special cases!
			var result *SchemaRef = nil
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
			case fieldKind == reflect.Struct && (isNullType(fieldType, "NullTime", "Time") || isNullTypeWrapper(fieldType, "NullTime", "Time")): // todo: we could also check whether the Time field is of time.Time type
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
				valueSchema := generateSchema(fieldType.Elem(), false)
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
				result = generateSchema(fieldType, false)

			// handle general slices / arrays
			case fieldKind == reflect.Slice, fieldKind == reflect.Array:
				result = &SchemaRef{Value: &Schema{
					Type:  &Types{"array"},
					Items: generateSchema(fieldType.Elem(), false),
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
			result.Value.Nullable = isNullable

			// handle property name
			fieldName := field.Name
			if jsonTagExists && strings.Split(jsonTag, ",")[0] != "" {
				fieldName = strings.Split(jsonTag, ",")[0]
			} else if formTagExists && strings.Split(formTag, ",")[0] != "" {
				fieldName = strings.Split(formTag, ",")[0]
			} else if queryTagExists && strings.Split(queryTag, ",")[0] != "" {
				fieldName = strings.Split(queryTag, ",")[0]
			}

			// handle json tag options
			jsonTagOptions := strings.Split(jsonTag, ",")
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
					alreadyRequired := false
					for _, r := range schema.Required {
						if r == fieldName {
							alreadyRequired = true
							break
						}
					}
					if !alreadyRequired {
						schema.Required = append(schema.Required, fieldName)
					}
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
					// oneof is more important than all other options since that's what the validator is using...
					// in that case, ignore and overwrite every other enum / OneOf options
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

	switch {
	case t == timeType:
		schema.Type = &Types{"string"}
		schema.Format = "date-time"
		return &schema

	case t == fileHeaderType:
		schema.Type = &Types{"string"}
		schema.Format = "binary"
		return &schema

	case t == uuidType:
		schema.Type = &Types{"string"}
		schema.Format = "uuid"
		return &schema

	case t == rawMessageType:
		return &schema
	}

	if t.Kind() == reflect.Struct {
		nullTypes := []struct {
			name, field, openType, format string
			min, max                      *float64
		}{
			{"NullUUID", "UUID", "string", "uuid", nil, nil},
			{"NullBool", "Bool", "boolean", "", nil, nil},
			{"NullByte", "Byte", "string", "byte", nil, nil},
			{"NullString", "String", "string", "", nil, nil},
			{"NullTime", "Time", "string", "date-time", nil, nil},
			{"NullInt16", "Int16", "integer", "", &minInt16, &maxInt16},
			{"NullInt32", "Int32", "integer", "int32", &minInt32, &maxInt32},
			{"NullInt64", "Int64", "integer", "int64", &minInt64, &maxInt64},
			{"NullFloat64", "Float64", "number", "double", &minInt64, &maxInt64},
		}

		for _, nt := range nullTypes {
			if isNullType(t, nt.name, nt.field) || isNullTypeWrapper(t, nt.name, nt.field) {
				schema.Type = &Types{nt.openType}
				schema.Format = nt.format
				schema.Nullable = true
				schema.Min = nt.min
				schema.Max = nt.max
				if nt.openType == "integer" || nt.openType == "number" {
					schema.ExclusiveMin = false
					schema.ExclusiveMax = false
				}
				return &schema
			}
		}
	}

	switch t.Kind() {
	case reflect.Bool:
		schema.Type = &Types{"boolean"}
		schema.Default = false

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		schema.Type = &Types{"integer"}
		switch t.Kind() {
		case reflect.Int:
			schema.Min, schema.Max = &minInt, &maxInt
		case reflect.Int8:
			schema.Min, schema.Max = &minInt8, &maxInt8
		case reflect.Int16:
			schema.Min, schema.Max = &minInt16, &maxInt16
		case reflect.Int32:
			schema.Format = "int32"
			schema.Min, schema.Max = &minInt32, &maxInt32
		case reflect.Int64:
			schema.Format = "int64"
			schema.Min, schema.Max = &minInt64, &maxInt64
		}
		schema.ExclusiveMin = false
		schema.ExclusiveMax = false
		schema.Default = 0

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		schema.Type = &Types{"integer"}
		schema.Min = &zeroUInt
		switch t.Kind() {
		case reflect.Uint:
			schema.Max = &maxUint
		case reflect.Uint8:
			schema.Max = &maxUint8
		case reflect.Uint16:
			schema.Max = &maxUint16
		case reflect.Uint32:
			schema.Max = &maxUint32
		case reflect.Uint64:
			schema.Max = &maxUint64
		}
		schema.ExclusiveMin = false
		schema.ExclusiveMax = false
		schema.Default = 0

	case reflect.Float32:
		schema.Type = &Types{"number"}
		schema.Format = "float"
		schema.Min, schema.Max = &minFloat32, &maxFloat32
		schema.ExclusiveMin = false
		schema.ExclusiveMax = false
		schema.Default = 0

	case reflect.Float64:
		schema.Type = &Types{"number"}
		schema.Format = "double"
		schema.Min, schema.Max = &minFloat64, &maxFloat64
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
		option_schema := generateSchema(fieldType, true)
		option_schema.Value.Default = option
		result.Value.OneOf = append(result.Value.OneOf, option_schema)
		result.Value.Enum = append(result.Value.Enum, option)
	}
	result.Value.Default = nil
}
