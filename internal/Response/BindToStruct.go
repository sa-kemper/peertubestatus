package Response

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// BindToStruct gets the form tag from structs, retrieves the *http.Request values and sets the values from the struct according to the request
// This is used to parse http requests to a generic endpoint and its specification
// Example:
//
// var endpoint = struct{age int `form:user_age`}
// err := BindToStruct(request, &endpoint)
// // endpoint will now contain the provided age
func BindToStruct(r *http.Request, dest interface{}) error {
	v := reflect.ValueOf(dest)

	// Ensure dest is a pointer to a struct
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return errors.New("destination must be a pointer to a struct")
	}

	v = v.Elem()
	query := r.URL.Query()

	for i := 0; i < v.Type().NumField(); i++ {
		field := v.Type().Field(i)
		paramName := field.Tag.Get("form")

		if paramName == "" {
			paramName = strings.ToLower(field.Name)
		}

		paramValue := query.Get(paramName)
		fieldValue := v.Field(i)

		if paramValue == "" {
			if field.Type.Kind() == reflect.Struct {
				// One level of recursion for nested structs
				if fieldValue.Kind() == reflect.Struct && fieldValue.CanAddr() {
					// Create a new struct pointer for nested binding
					nestedPtr := fieldValue.Addr().Interface()
					if err := BindToStruct(r, nestedPtr); err != nil {
						return fmt.Errorf("error binding nested struct %s: %v", field.Name, err)
					}
				}
			}
			continue
		}

		if !fieldValue.CanSet() {
			continue
		}

		if err := setFieldValue(fieldValue, paramValue); err != nil {
			return errors.New("error setting field " + field.Name + ": " + err.Error())
		}
	}

	return nil
}

func setFieldValue(fieldValue reflect.Value, paramValue string) error {
	switch fieldValue.Kind() {
	case reflect.String:
		fieldValue.SetString(paramValue)
	case reflect.Int, reflect.Int64:
		intVal, err := strconv.ParseInt(paramValue, 10, 64)
		if err != nil {
			return err
		}
		fieldValue.SetInt(intVal)
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(paramValue)
		if err != nil {
			return err
		}
		fieldValue.SetBool(boolVal)
	case reflect.Struct:
		// Handle time.Time and other struct types
		if fieldValue.Type() == reflect.TypeOf(time.Time{}) {
			timeVal, err := time.Parse("2006-01-02", paramValue)
			if err != nil {
				return err
			}
			if timeVal.IsZero() {
				timeVal = time.Now()
			}
			fieldValue.Set(reflect.ValueOf(timeVal))
		} else {
			return errors.New("unsupported struct type: " + fieldValue.Type().String())
		}
	case reflect.Slice:
		// Basic slice support (comma-separated values)
		values := strings.Split(paramValue, ",")
		slice := reflect.MakeSlice(fieldValue.Type(), len(values), len(values))

		for i, val := range values {
			elemValue := slice.Index(i)
			if err := setFieldValue(elemValue, val); err != nil {
				return err
			}
		}
		fieldValue.Set(slice)
	default:
		return errors.New("unsupported field type: " + fieldValue.Kind().String())
	}
	return nil
}
