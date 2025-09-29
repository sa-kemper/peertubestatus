package peertubeApi

import (
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

// toQueryParams converts a struct to URL query parameters.
//
// This function uses reflection to transform a struct's exported fields into URL query parameters.
// It handles various field types including strings, integers, booleans, and slices.
//
// Key behaviors:
// - Converts struct field names to camelCase for query parameter names
// - Skips unexported (private) struct fields
// - Supports slices of strings, integers, and booleans
//
// Example:
//
//	type SearchParams struct {
//	  Query string
//	  Limit int
//	  Tags  []string
//	}
//	params := SearchParams{Query: "example", Limit: 10, Tags: []string{"go", "programming"}}
//	queryValues := toQueryParams(params)
//	// Result would be: "query=example&limit=10&tags=go&tags=programming"
//
// Parameters:
//   - paramObject: A struct (or pointer to a struct) to be converted to query parameters
//
// Returns:
//   - url.Values containing the query parameters derived from the struct fields
//
// See Also: Test_toQueryParams for a detailed behaviour demonstration.
func toQueryParams(paramObject any) url.Values {
	values := url.Values{}

	v := reflect.ValueOf(paramObject)

	// Handle pointer to struct
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Ensure we're working with a struct
	if v.Kind() != reflect.Struct {
		return values
	}

	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if fieldType.PkgPath != "" {
			continue
		}

		// Convert first char of field name to lowercase
		paramName := func(s string) string {
			if s == "" {
				return s
			}
			return strings.ToLower(s[:1]) + s[1:]
		}(fieldType.Name)

		// Handle different types of fields
		switch field.Kind() {
		case reflect.Slice:
			if field.Len() > 0 {
				// The slice contains values.
				sliceType := field.Index(0).Type()
				switch sliceType.Kind() {
				case reflect.String:
					for i := 0; i < field.Len(); i++ {
						values.Add(paramName, field.Index(i).String())
					}
				case reflect.Int64:
					fallthrough
				case reflect.Int32:
					fallthrough
				case reflect.Int16:
					fallthrough
				case reflect.Int8:
					fallthrough
				case reflect.Uint8:
					fallthrough
				case reflect.Uint16:
					fallthrough
				case reflect.Int:
					for i := 0; i < field.Len(); i++ {
						handleInt(field.Index(i), &values, paramName)
					}
				case reflect.Bool:
					for i := 0; i < field.Len(); i++ {
						values.Add(paramName, strconv.FormatBool(field.Index(i).Bool()))
					}
				}
			} else {

				// values.Add(paramName, "") // add the argument anyway... peertube is stupid sometimes.
			}
		case reflect.Int64:
			fallthrough
		case reflect.Int:
			handleInt(field, &values, paramName)
		case reflect.Bool:
			if field.Bool() {
				values.Set(paramName, "true")
			}
		case reflect.String:
			values.Set(paramName, field.String())
		}
	}

	return values
}

func handleInt(field reflect.Value, values *url.Values, paramName string) {
	values.Add(paramName, strconv.FormatInt(field.Int(), 10))
}
