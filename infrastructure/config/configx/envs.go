package configx

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
)

// Loader represents the strategy used to load the configuration field.
// It's configured via the field tag, in the key section.
type Loader string

const (
	// EnvLoader loads the configuration from the environment variable
	// provided in the field tag, in the value section.
	//
	// The environment variable is automatically expanded.
	//
	// Example:
	//  type Config struct {
	// 		Foo string `env:"FOO_VAR"`
	//  }
	EnvLoader Loader = "env"

	// DefaultLoader specifies a default value, in case the other loaders fail.
	//
	// Example:
	//  type Config struct {
	// 		Foo string `env:"FOO_VAR" default:"foo"`
	//  }
	DefaultLoader Loader = "default"
)

var ErrNoPointer = errors.New("provided config parameter is not a pointer")
var ErrNoStruct = errors.New("provided config parameter reference is not a struct")
var ErrUnhandledType = errors.New("field type is unhandled")
var ErrUnsafe = errors.New("field has no supported source tag; if this is intended, run [LoadUnsafe]")
var ErrNotFound = errors.New("field source not found by loader")
var ErrParsingFailed = errors.New("loader parsing failed")

// Load inspects the provided configuration and loads every field with
// the provided strategies (see [Loader]).
//
// The parameter must be a pointer to the configuration object to load.
// The configuration type must be public, as all its members.
// The function is recursive: inner structs will be inspected.
//
// This function returns an error if it finds a field without a [Loader] annotation;
// if this is not the desired behavior, call [LoadUnsafe]
//
// Example:
//
//	 type Config struct {
//		Foo string `env:"FOO_VAR" default:"foo"`
//	 }
//
//	 func main() {
//			var config Config
//			if err := configx.Load(&config); err != nil {
//				// handle error...
//			}
//
//			// configuration loaded ✨
//	 }
func Load(config any) error {
	return load(config, true)
}

// LoadUnsafe is the unsafe version of [Load]; refer to its documentation
func LoadUnsafe(config any) error {
	return load(config, false)
}

func load(config any, safe bool) error {
	val := reflect.ValueOf(config)
	if val.Kind() != reflect.Pointer {
		return fmt.Errorf("%w: value kind is %s", ErrNoPointer, val.Kind())
	}

	val = reflect.Indirect(val)
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("%w: referenced value kind is %s", ErrNoStruct, val.Kind())
	}

	for i := range val.NumField() {
		field := val.Type().Field(i)
		fieldVal := val.Field(i)

		if field.Type.Kind() == reflect.Struct {
			if err := load(fieldVal.Addr().Interface(), safe); err != nil {
				return fmt.Errorf("field %s, type %s: %w", field.Name, field.Type.Name(), err)
			}

			continue
		}

		// the only implementation for now. we could support multiple sources in the future
		envTag, ok := field.Tag.Lookup(string(EnvLoader))
		if !ok {
			if safe {
				return fmt.Errorf("field %s, type %s: %w", field.Name, field.Type.Name(), ErrUnsafe)
			}

			continue
		}

		envVal, ok := os.LookupEnv(envTag)
		if !ok {
			defVal, ok := field.Tag.Lookup(string(DefaultLoader))
			if !ok {
				return fmt.Errorf("field %s, type %s, loader %s: %w", field.Name, field.Type.Name(), EnvLoader, ErrNotFound)
			}

			envVal = defVal
		} else {
			envVal = os.ExpandEnv(envVal)
		}

		parseErr := fmt.Errorf("field %s, type %s: %w", field.Name, field.Type.Name(), ErrParsingFailed)
		switch field.Type.Kind() {
		case reflect.String:
			fieldVal.Set(reflect.ValueOf(envVal).Convert(field.Type))

		case reflect.Int:
			envInt, err := strconv.ParseInt(envVal, 10, strconv.IntSize)
			if err != nil {
				return parseErr
			}

			fieldVal.Set(reflect.ValueOf(envInt).Convert(field.Type))

		case reflect.Int64:
			envInt, err := strconv.ParseInt(envVal, 10, 64)
			if err != nil {
				return parseErr
			}

			fieldVal.Set(reflect.ValueOf(envInt).Convert(field.Type))

		case reflect.Int32:
			envInt, err := strconv.ParseInt(envVal, 10, 32)
			if err != nil {
				return parseErr
			}

			fieldVal.Set(reflect.ValueOf(envInt).Convert(field.Type))

		case reflect.Int16:
			envInt, err := strconv.ParseInt(envVal, 10, 16)
			if err != nil {
				return parseErr
			}

			fieldVal.Set(reflect.ValueOf(envInt).Convert(field.Type))

		case reflect.Int8:
			envInt, err := strconv.ParseInt(envVal, 10, 8)
			if err != nil {
				return parseErr
			}

			fieldVal.Set(reflect.ValueOf(envInt).Convert(field.Type))

		case reflect.Uint:
			envInt, err := strconv.ParseUint(envVal, 10, strconv.IntSize)
			if err != nil {
				return parseErr
			}

			fieldVal.Set(reflect.ValueOf(envInt).Convert(field.Type))

		case reflect.Uint64:
			envInt, err := strconv.ParseUint(envVal, 10, 64)
			if err != nil {
				return parseErr
			}

			fieldVal.Set(reflect.ValueOf(envInt).Convert(field.Type))

		case reflect.Uint32:
			envInt, err := strconv.ParseUint(envVal, 10, 32)
			if err != nil {
				return parseErr
			}

			fieldVal.Set(reflect.ValueOf(envInt).Convert(field.Type))

		case reflect.Uint16:
			envInt, err := strconv.ParseUint(envVal, 10, 16)
			if err != nil {
				return parseErr
			}

			fieldVal.Set(reflect.ValueOf(envInt).Convert(field.Type))

		case reflect.Uint8:
			envInt, err := strconv.ParseUint(envVal, 10, 8)
			if err != nil {
				return parseErr
			}

			fieldVal.Set(reflect.ValueOf(envInt).Convert(field.Type))

		case reflect.Float64:
			envFloat, err := strconv.ParseFloat(envVal, 64)
			if err != nil {
				return parseErr
			}

			fieldVal.Set(reflect.ValueOf(envFloat).Convert(field.Type))

		case reflect.Float32:
			envFloat, err := strconv.ParseFloat(envVal, 32)
			if err != nil {
				return parseErr
			}

			fieldVal.Set(reflect.ValueOf(envFloat).Convert(field.Type))

		case reflect.Bool:
			envBool, err := strconv.ParseBool(envVal)
			if err != nil {
				return parseErr
			}

			fieldVal.Set(reflect.ValueOf(envBool).Convert(field.Type))

		default:
			return fmt.Errorf("%w: %s kind is %s", ErrUnhandledType, field.Name, field.Type.Kind())
		}
	}

	return nil
}
