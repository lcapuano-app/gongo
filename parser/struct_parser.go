package parser

import (
	"errors"
	"reflect"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	bsonTag  = "bson"
	jsonTag  = "json"
	gongoTag = "gongo"
)

// Uses a interface "s" to parse the string value, usually from a query url string,
// to its struct primitive
//
// returns (parsedValue any, err error)
func parsePropValue(prop, strValue string, s interface{}) (parsedValue any, err error) {

	field, err := getFieldByProp(prop, s, nil)

	if err != nil {
		return parsedValue, err
	}

	if prop == "_id" {
		return primitive.ObjectIDFromHex(strValue)
	}

	return parseToPrimitive(field, strValue)
}

// Gets the struct field from a bson/json/gongo name
//
// eg: Name string `json:"name,omitempty,..."`
//
// returns Name reflect.StructField
func getFieldByProp(prop string, s interface{}, rt reflect.Type) (field reflect.StructField, err error) {

	props := strings.Split(prop, ".")

	if rt == nil {
		rt = reflect.TypeOf(s)
	}

	if rt.Kind() != reflect.Struct {
		return field, errors.New("bad type")
	}

	for i := 0; i < rt.NumField(); i++ {

		field := rt.Field(i)
		v := getPropValue(field)

		if v != props[0] {
			continue
		}

		if len(props) == 1 {
			return field, nil
		}

		kind := field.Type.Kind()
		if kind == reflect.Struct || kind == reflect.Interface {
			props = props[1:]

			if len(props) >= 1 {
				prop = strings.Join(props, ".")
				return getFieldByProp(prop, nil, field.Type)
			}

		}

		return field, errors.New("invalid prop query: " + prop)
	}

	return field, nil
}

// Just removes all "noise" from the struct field
//
// eg: Name string `json:"name,omitempty,..."`
//
// returns "name"
func getPropValue(field reflect.StructField) (fieldname string) {
	// use split to ignore tag "options" like omitempty, etc.

	var value string

	value = strings.Split(field.Tag.Get(gongoTag), ",")[0]
	if value != "" {
		return value
	}

	value = strings.Split(field.Tag.Get(bsonTag), ",")[0]
	if value != "" {
		return value
	}

	return strings.Split(field.Tag.Get(jsonTag), ",")[0]
}

// Parses the string value (val) to the correspondin
// field reflect from the given model
func parseToPrimitive(field reflect.StructField, val string) (parsedValue any, err error) {

	kind := field.Type.Kind()

	if kind == reflect.String {
		return val, nil
	}

	switch kind {

	case reflect.Array, reflect.Map, reflect.Slice:
		return "MAP", nil

	case reflect.Bool:
		res, err := strconv.ParseBool(val)
		if err != nil {
			return true, err
		}
		return res, err

		// COMPLEX
	case reflect.Complex64:
		return strconv.ParseComplex(val, 64)
	case reflect.Complex128:
		return strconv.ParseComplex(val, 128)

		// FLOAT
	case reflect.Float32:
		return strconv.ParseFloat(val, 32)
	case reflect.Float64:
		return strconv.ParseFloat(val, 64)

		// INT
	case reflect.Int, reflect.Int64:
		return strconv.ParseInt(val, 10, 64)
	case reflect.Int8:
		return strconv.ParseInt(val, 10, 8)
	case reflect.Int16:
		return strconv.ParseInt(val, 10, 16)
	case reflect.Int32:
		return strconv.ParseInt(val, 10, 32)

		// NESTED
	case reflect.Interface, reflect.Struct:
		return "STRUCT", nil

		// UINT
	case reflect.Uint, reflect.Uint64:
		return strconv.ParseUint(val, 10, 64)
	case reflect.Uint8:
		return strconv.ParseUint(val, 10, 8)
	case reflect.Uint16:
		return strconv.ParseUint(val, 10, 16)
	case reflect.Uint32:
		return strconv.ParseUint(val, 10, 32)

	default:
		return val, nil
	}
}
