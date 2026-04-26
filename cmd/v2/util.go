package cmd

import (
	"os"
	"reflect"
	"strings"
)

// UpdateConfig will update all config string fields (recursively), which the value has pattern ${ENV_NAME}
// and use the value from the environment variable.
// e.g: struct { Pass string } -> Pass: "${DB_PASS}" -> Pass: "password"
func UpdateConfig(ptr any) error {
	if ptr == nil {
		// do nothing
		return nil
	}

	// get the value of the pointer
	v := reflect.ValueOf(ptr).Elem()
	// iterate over the fields of the struct
	for i := 0; i < v.NumField(); i++ {
		fv := v.Field(i)

		// if the field is a struct, recursively update the fields
		if fv.Kind() == reflect.Struct {
			err := UpdateConfig(fv.Addr().Interface())
			if err != nil {
				return err
			}
		}

		// if the field is a pointer to a struct, recursively update the fields
		if fv.Kind() == reflect.Ptr && fv.Elem().Kind() == reflect.Struct {
			err := UpdateConfig(fv.Interface())
			if err != nil {
				return err
			}
		}

		// if the field is a string and the value has pattern ${ENV_NAME}
		if fv.Kind() == reflect.String {
			updateStringField(fv)
		}

		// if the field is a pointer to a string
		if fv.Kind() == reflect.Ptr && fv.Elem().Kind() == reflect.String {
			updateStringField(fv.Elem())
		}
	}

	return nil
}

// updateStringField updates a string field if it matches the ${ENV_NAME} pattern
func updateStringField(fv reflect.Value) {
	// get the value of the field
	value := fv.String()
	// get the environment variable name
	envName := getEnvName(value)

	// if the environment variable is not empty
	if envName != "" {
		// get the value of the environment variable
		envValue := os.Getenv(envName)
		// set the value of the field
		fv.SetString(envValue)
	}
}

// getEnvName will get the environment variable name from the value
// e.g: "${DB_PASS}" -> "DB_PASS"
func getEnvName(val string) string {
	val = strings.TrimSpace(val)
	if len(val) < 3 {
		return ""
	}

	if val[0] == '$' && val[1] == '{' && val[len(val)-1] == '}' {
		return strings.TrimSpace(val[2 : len(val)-1])
	}

	return ""
}
