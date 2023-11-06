// Package bind contains convenience functions to decode HTTP request data.
package bind

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-playground/form/v4"
)

type Flag int

const (
	// When the Vacuum flag is set, url.Values is cleaned before trying to bind the values.
	// Strings are trimmed, empty strings and zero length slices are deleted.
	Vacuum Flag = iota
)

var (
	queryDecoder  = form.NewDecoder()
	formDecoder   = form.NewDecoder()
	headerDecoder = form.NewDecoder()

	PathValueFunc func(*http.Request, string) string
)

func init() {
	queryDecoder.SetTagName("query")
	queryDecoder.SetMode(form.ModeExplicit)
	formDecoder.SetTagName("form")
	formDecoder.SetMode(form.ModeExplicit)
	headerDecoder.SetTagName("header")
	headerDecoder.SetMode(form.ModeExplicit)
}

func Request(r *http.Request, v any, flags ...Flag) error {
	if err := Path(r, v, flags...); err != nil {
		return err
	}
	if err := Header(r, v, flags...); err != nil {
		return err
	}
	if r.Method == http.MethodGet || r.Method == http.MethodDelete || r.Method == http.MethodHead {
		return Query(r, v, flags...)
	}
	return Body(r, v, flags...)
}

func Query(r *http.Request, v any, flags ...Flag) error {
	vals := r.URL.Query()
	if hasFlag(flags, Vacuum) {
		vals = vacuum(vals)
	}
	return queryDecoder.Decode(v, vals)
}

func Body(r *http.Request, v any, flags ...Flag) error {
	if r.ContentLength == 0 {
		return nil
	}

	ct := r.Header.Get("Content-Type")

	switch {
	case strings.HasPrefix(ct, "application/json"):
		return json.NewDecoder(r.Body).Decode(v)
	case strings.HasPrefix(ct, "application/xml") || strings.HasPrefix(ct, "text/xml"):
		return xml.NewDecoder(r.Body).Decode(v)
	case strings.HasPrefix(ct, "application/x-www-form-urlencoded") || strings.HasPrefix(ct, "multipart/form-data"):
		r.ParseForm()
		vals := r.Form
		if hasFlag(flags, Vacuum) {
			vals = vacuum(vals)
		}
		return formDecoder.Decode(v, vals)
	}
	return nil
}

func Header(r *http.Request, v any, flags ...Flag) error {
	vals := url.Values(r.Header)
	if hasFlag(flags, Vacuum) {
		vals = vacuum(vals)
	}
	return headerDecoder.Decode(v, vals)
}

// TODO handle embedded structs
// TODO make vacuum aware?
func Path(r *http.Request, v any, flags ...Flag) error {
	if PathValueFunc == nil {
		return nil
	}

	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return &form.InvalidDecoderError{Type: reflect.TypeOf(v)}
	}
	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return nil
	}

	typ := val.Type()

	// TODO cache this
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		pathParam := field.Tag.Get("path")
		if pathParam != "" && pathParam != "-" {
			value := PathValueFunc(r, pathParam)
			if err := setWithProperType(field.Type.Kind(), value, val.Field(i)); err != nil {
				return err
			}
		}
	}

	return nil
}

func vacuum(values url.Values) url.Values {
	newValues := make(url.Values)
	for key, vals := range values {
		var newVals []string
		for _, val := range vals {
			val = strings.TrimSpace(val)
			if val != "" {
				newVals = append(newVals, val)
			}
		}
		if len(newVals) > 0 {
			newValues[key] = newVals
		}
	}
	return newValues
}

func hasFlag(flags []Flag, flag Flag) bool {
	for _, f := range flags {
		if f == flag {
			return true
		}
	}
	return false
}

// code below is taken from Echo's bind implementation
func setWithProperType(valueKind reflect.Kind, val string, structField reflect.Value) error {
	switch valueKind {
	case reflect.Ptr:
		return setWithProperType(structField.Elem().Kind(), val, structField.Elem())
	case reflect.Int:
		return setIntField(val, 0, structField)
	case reflect.Int8:
		return setIntField(val, 8, structField)
	case reflect.Int16:
		return setIntField(val, 16, structField)
	case reflect.Int32:
		return setIntField(val, 32, structField)
	case reflect.Int64:
		return setIntField(val, 64, structField)
	case reflect.Uint:
		return setUintField(val, 0, structField)
	case reflect.Uint8:
		return setUintField(val, 8, structField)
	case reflect.Uint16:
		return setUintField(val, 16, structField)
	case reflect.Uint32:
		return setUintField(val, 32, structField)
	case reflect.Uint64:
		return setUintField(val, 64, structField)
	case reflect.Bool:
		return setBoolField(val, structField)
	case reflect.Float32:
		return setFloatField(val, 32, structField)
	case reflect.Float64:
		return setFloatField(val, 64, structField)
	case reflect.String:
		structField.SetString(val)
	default:
		return errors.New("unknown type")
	}
	return nil
}

func setIntField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0"
	}
	intVal, err := strconv.ParseInt(value, 10, bitSize)
	if err == nil {
		field.SetInt(intVal)
	}
	return err
}

func setUintField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0"
	}
	uintVal, err := strconv.ParseUint(value, 10, bitSize)
	if err == nil {
		field.SetUint(uintVal)
	}
	return err
}

func setBoolField(value string, field reflect.Value) error {
	if value == "" {
		value = "false"
	}
	boolVal, err := strconv.ParseBool(value)
	if err == nil {
		field.SetBool(boolVal)
	}
	return err
}

func setFloatField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0.0"
	}
	floatVal, err := strconv.ParseFloat(value, bitSize)
	if err == nil {
		field.SetFloat(floatVal)
	}
	return err
}
