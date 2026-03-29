package gofiberswagger

import (
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
)

var (
	acquiredSchemas map[string]*SchemaRef
	schemasMutex    sync.RWMutex
)

func setToAcquiredSchemas(ref string, schema *SchemaRef) {
	schemasMutex.Lock()
	defer schemasMutex.Unlock()
	if acquiredSchemas == nil {
		acquiredSchemas = make(map[string]*SchemaRef)
	}
	if schema != nil {
		acquiredSchemas[ref] = schema
	}
}

func getFromAcquiredSchemas(ref string) *SchemaRef {
	schemasMutex.RLock()
	defer schemasMutex.RUnlock()
	if acquiredSchemas == nil {
		return nil
	}
	return acquiredSchemas[ref]
}

func CreateSchema[T any]() *SchemaRef {
	var t T
	reflectType := reflect.TypeOf(t)
	if reflectType == nil {
		return &SchemaRef{Value: &Schema{}}
	}
	return generateSchema(reflectType, false)
}

func getSpecialTypeSchema(t reflect.Type) (schema *Schema, isNullable bool, handled bool) {
	kind := t.Kind()
	name := t.Name()
	pkg := t.PkgPath()

	if t == timeType {
		return &Schema{Type: &Types{"string"}, Format: "date-time"}, false, true
	}

	if kind == reflect.Struct && name == "FileHeader" && pkg == "mime/multipart" {
		return &Schema{Type: &Types{"string"}, Format: "binary"}, false, true
	}

	isUUIDArray := kind == reflect.Array && name == "UUID" && t.Elem().Kind() == reflect.Uint8
	if t == uuidType || isUUIDArray {
		return &Schema{Type: &Types{"string"}, Format: "uuid"}, false, true
	}

	if t == rawMessageType {
		return &Schema{}, false, true
	}

	if kind == reflect.Struct {
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
				s := &Schema{
					Type:   &Types{nt.openType},
					Format: nt.format,
					Min:    nt.min,
					Max:    nt.max,
				}
				if nt.openType == "integer" || nt.openType == "number" {
					s.ExclusiveMin, s.ExclusiveMax = false, false
				}
				return s, true, true
			}
		}
	}

	return nil, false, false
}

func generateSchema(t reflect.Type, stopRecursion bool) *SchemaRef {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	tName := t.Name()
	if tName == "" {
		genPart := uuid.NewString()
		tName = "generated-" + genPart
	}

	ref := strings.ReplaceAll(strings.ReplaceAll(t.PkgPath(), "/", "_"), ".", "_") + tName
	refPath := "#/components/schemas/" + ref

	if cached := getFromAcquiredSchemas(ref); cached != nil {
		if t.Kind() == reflect.Struct {
			return &SchemaRef{Ref: refPath, Extensions: cached.Extensions, Origin: cached.Origin, Value: cached.Value}
		}
		return cached
	}

	if special, isNullable, ok := getSpecialTypeSchema(t); ok {
		special.Nullable = isNullable
		return &SchemaRef{Value: special}
	}

	schema := getDefaultSchema(t)
	if schema.Type != nil && (*schema.Type)[0] != "object" && t.Kind() != reflect.Struct {
		return &SchemaRef{Value: schema}
	}

	if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {
		if t.Elem().Kind() == reflect.Uint8 {
			return &SchemaRef{Value: &Schema{Type: &Types{"string"}, Format: "byte"}}
		}
		return &SchemaRef{Value: &Schema{Type: &Types{"array"}, Items: generateSchema(t.Elem(), false)}}
	}

	if t.Kind() == reflect.Struct {
		if t.NumField() == 0 {
			schema.Type = &Types{"object"}
			return &SchemaRef{Value: schema}
		}

		if !stopRecursion && implementsSwaggerEnum(t) {
			enumSchema := &SchemaRef{Value: getDefaultSchema(t)}
			handleEnumValues(enumSchema, getSwaggerEnumValues(t), false, t)
			return enumSchema
		}

		schema.Title = tName
		schema.Type = &Types{"object"}
		setToAcquiredSchemas(ref, &SchemaRef{Value: &Schema{}})

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.Name == "XMLName" {
				continue
			}

			if field.Anonymous {
				fType := field.Type
				for fType.Kind() == reflect.Pointer {
					fType = fType.Elem()
				}
				if fType.Kind() == reflect.Struct {
					sub := generateSchema(fType, false)
					if sub != nil && sub.Value != nil {
						for name, prop := range sub.Value.Properties {
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
			xmlTag, xmlTagExists := field.Tag.Lookup("xml")
			if jsonTagExists && strings.Split(jsonTag, ",")[0] == "-" && !formTagExists && !queryTagExists {
				continue
			}
			if xmlTagExists && strings.Split(xmlTag, ",")[0] == "-" && !jsonTagExists && !formTagExists && !queryTagExists {
				continue
			}

			fieldType := field.Type
			isNullable := false
			for fieldType.Kind() == reflect.Pointer {
				fieldType = fieldType.Elem()
				isNullable = true
			}

			var result *SchemaRef
			if spec, specNull, ok := getSpecialTypeSchema(fieldType); ok {
				result = &SchemaRef{Value: spec}
				if specNull {
					isNullable = true
				}
			} else {
				switch fieldType.Kind() {
				case reflect.Func, reflect.Chan:
					continue
				case reflect.Map, reflect.Interface:
					result = &SchemaRef{Value: &Schema{Type: &Types{"object"}}}
					if fieldType.Kind() == reflect.Map && fieldType.Key().Kind() == reflect.String {
						has := true
						result.Value.AdditionalProperties = AdditionalProperties{Has: &has, Schema: generateSchema(fieldType.Elem(), false)}
					}
				case reflect.Slice, reflect.Array:
					if fieldType.Elem().Kind() == reflect.Uint8 {
						result = &SchemaRef{Value: &Schema{Type: &Types{"string"}, Format: "byte"}}
					} else {
						result = &SchemaRef{Value: &Schema{Type: &Types{"array"}, Items: generateSchema(fieldType.Elem(), false)}}
					}
				case reflect.Struct:
					result = generateSchema(fieldType, false)
				default:
					result = &SchemaRef{Value: getDefaultSchema(fieldType)}
				}
			}

			fieldSchema := *result.Value
			fieldResult := &SchemaRef{
				Ref:   result.Ref,
				Value: &fieldSchema,
			}
			fieldResult.Value.Nullable = isNullable
			fieldName := field.Name
			for _, tag := range []string{jsonTag, formTag, queryTag} {
				if parts := strings.Split(tag, ","); parts[0] != "" && parts[0] != "-" {
					fieldName = parts[0]
					break
				}
			}
			fieldResult.Value.Title = fieldName

			parseTags(field, fieldResult)
			if implementsSwaggerEnum(fieldType) {
				handleEnumValues(fieldResult, getSwaggerEnumValues(fieldType), false, fieldType)
			}
			applyValidationTags(field, fieldResult, schema, fieldName)

			schema.Properties[fieldName] = fieldResult
		}

		setToAcquiredSchemas(ref, &SchemaRef{Value: schema})
		return &SchemaRef{Ref: refPath, Value: schema}
	}

	return &SchemaRef{Value: schema}
}

func getDefaultSchema(t reflect.Type) *Schema {
	schema := &Schema{Properties: make(Schemas), Required: []string{}}

	if special, isNullable, ok := getSpecialTypeSchema(t); ok {
		special.Nullable = isNullable
		return special
	}

	switch t.Kind() {
	case reflect.Bool:
		schema.Type, schema.Default = &Types{"boolean"}, false
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		schema.Type, schema.Default = &Types{"integer"}, 0
		schema.ExclusiveMin, schema.ExclusiveMax = false, false
		switch t.Kind() {
		case reflect.Int:
			schema.Min, schema.Max = &minInt, &maxInt
		case reflect.Int8:
			schema.Min, schema.Max = &minInt8, &maxInt8
		case reflect.Int16:
			schema.Min, schema.Max = &minInt16, &maxInt16
		case reflect.Int32:
			schema.Format, schema.Min, schema.Max = "int32", &minInt32, &maxInt32
		case reflect.Int64:
			schema.Format, schema.Min, schema.Max = "int64", &minInt64, &maxInt64
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		schema.Type, schema.Default, schema.Min = &Types{"integer"}, 0, &zeroUInt
		schema.ExclusiveMin, schema.ExclusiveMax = false, false
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
	case reflect.Float32, reflect.Float64:
		schema.Type, schema.Default = &Types{"number"}, 0
		schema.ExclusiveMin, schema.ExclusiveMax = false, false
		if t.Kind() == reflect.Float32 {
			schema.Format, schema.Min, schema.Max = "float", &minFloat32, &maxFloat32
		} else {
			schema.Format, schema.Min, schema.Max = "double", &minFloat64, &maxFloat64
		}
	case reflect.String:
		schema.Type, schema.Default = &Types{"string"}, ""
	}
	return schema
}

func parseTags(field reflect.StructField, result *SchemaRef) {
	jsonTag := field.Tag.Get("json")
	for _, opt := range strings.Split(jsonTag, ",")[1:] {
		switch opt {
		case "string":
			result.Value.Type = &Types{"string"}
		case "omitempty", "omitzero":
			result.Value.Nullable = true
			result.Value.Description += " " + opt + " "
		}
	}

	xmlTag := field.Tag.Get("xml")
	if xmlTag != "" {
		parts := strings.Split(xmlTag, ",")
		if result.Value.XML == nil {
			result.Value.XML = &XML{}
		}
		if parts[0] != "" {
			result.Value.XML.Name = parts[0]
		}
		for _, opt := range parts[1:] {
			switch opt {
			case "attr":
				result.Value.XML.Attribute = true
			case "omitempty":
				result.Value.Nullable = true
			default:
				result.Value.Description += " " + opt + " "
			}
		}
	}
}

func applyValidationTags(field reflect.StructField, result *SchemaRef, parent *Schema, fieldName string) {
	validate := field.Tag.Get("validate")
	if validate == "" {
		return
	}

	kind := field.Type.Kind()
	for _, v := range strings.Split(validate, ",") {
		parts := strings.SplitN(v, "=", 2)
		key := parts[0]

		var value string
		if len(parts) > 1 {
			value = parts[1]
		}

		switch key {
		case "required":
			exists := false
			for _, r := range parent.Required {
				if r == fieldName {
					exists = true
					break
				}
			}
			if !exists {
				parent.Required = append(parent.Required, fieldName)
			}
			result.Value.Nullable = false
			result.Value.AllowEmptyValue = false
		case "min":
			val, err := strconv.ParseFloat(value, 64)
			if err != nil {
				continue
			}
			if kind == reflect.Slice || kind == reflect.Array {
				result.Value.MinItems = uint64(val)
			}
			if kind == reflect.String {
				result.Value.MinLength = uint64(val)
			}
			if kind >= reflect.Int && kind <= reflect.Float64 {
				result.Value.Min = &val
				result.Value.Default = val
			}
		case "max":
			val, err := strconv.ParseFloat(value, 64)
			if err != nil {
				continue
			}
			uVal := uint64(val)
			if kind == reflect.Slice || kind == reflect.Array {
				result.Value.MaxItems = &uVal
			}
			if kind == reflect.String {
				result.Value.MaxLength = &uVal
			}
			if kind >= reflect.Int && kind <= reflect.Float64 {
				result.Value.Max = &val
			}
		case "oneof":
			options := []any{}
			for _, opt := range strings.Split(value, " ") {
				options = append(options, opt)
			}
			handleEnumValues(result, options, true, field.Type)
		case "uniqueItems":
			result.Value.UniqueItems = true
		case "minLength":
			val, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				continue
			}
			result.Value.MinLength = val
		case "maxLength":
			val, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				continue
			}
			result.Value.MaxLength = &val
		case "omitnil":
			result.Value.Description += " omitnil "
		}
	}
}

func isNullType(fieldType reflect.Type, nullFieldName string, uniqueFieldName string) bool {
	if fieldType.Kind() != reflect.Struct || fieldType.Name() != nullFieldName {
		return false
	}
	_, hasValid := fieldType.FieldByName("Valid")
	_, hasUnique := fieldType.FieldByName(uniqueFieldName)
	return hasValid && hasUnique
}

func isNullTypeWrapper(fieldType reflect.Type, nullFieldName string, uniqueFieldName string) bool {
	if fieldType.Kind() != reflect.Struct || fieldType.NumField() != 1 {
		return false
	}
	return isNullType(fieldType.Field(0).Type, nullFieldName, uniqueFieldName)
}

func handleEnumValues(result *SchemaRef, options []any, overwrite bool, fieldType reflect.Type) {
	if result.Value.OneOf == nil || overwrite {
		result.Value.OneOf = []*SchemaRef{}
	}
	if result.Value.Enum == nil || overwrite {
		result.Value.Enum = []any{}
	}
	for _, opt := range options {
		optSchema := generateSchema(fieldType, true)
		optSchema.Value.Default = opt
		result.Value.OneOf = append(result.Value.OneOf, optSchema)
		result.Value.Enum = append(result.Value.Enum, opt)
	}
	result.Value.Default = nil
}
