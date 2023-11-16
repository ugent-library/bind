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

	queryEncoder  = form.NewEncoder()
	formEncoder   = form.NewEncoder()
	headerEncoder = form.NewEncoder()

	PathValueFunc func(*http.Request, string) string
)

func init() {
	queryDecoder.SetTagName("query")
	queryDecoder.SetMode(form.ModeExplicit)
	formDecoder.SetTagName("form")
	formDecoder.SetMode(form.ModeExplicit)
	headerDecoder.SetTagName("header")
	headerDecoder.SetMode(form.ModeExplicit)

	queryEncoder.SetTagName("query")
	queryEncoder.SetMode(form.ModeExplicit)
	formEncoder.SetTagName("form")
	formEncoder.SetMode(form.ModeExplicit)
	headerEncoder.SetTagName("header")
	headerEncoder.SetMode(form.ModeExplicit)
}

func EncodeQuery(v any) (url.Values, error) {
	return queryEncoder.Encode(v)
}

func EncodeForm(v any) (url.Values, error) {
	return formEncoder.Encode(v)
}

func EncodeHeader(v any) (http.Header, error) {
	vals, err := formEncoder.Encode(v)
	return http.Header(vals), err
}

func DecodeQuery(vals url.Values, v any, flags ...Flag) error {
	if hasFlag(flags, Vacuum) {
		vals = vacuum(vals)
	}
	return queryDecoder.Decode(v, vals)
}

func DecodeForm(vals url.Values, v any, flags ...Flag) error {
	if hasFlag(flags, Vacuum) {
		vals = vacuum(vals)
	}
	return formDecoder.Decode(v, vals)
}

func DecodeHeader(header http.Header, v any, flags ...Flag) error {
	vals := url.Values(header)
	if hasFlag(flags, Vacuum) {
		vals = vacuum(vals)
	}
	return headerDecoder.Decode(v, vals)
}

func PathValue(r *http.Request, k string) string {
	if PathValueFunc != nil {
		return PathValueFunc(r, k)
	}
	return ""
}

func Request(r *http.Request, v any, flags ...Flag) error {
	if PathValueFunc != nil {
		if err := Path(r, v, flags...); err != nil {
			return err
		}
	}
	if err := Header(r, v, flags...); err != nil {
		return err
	}
	if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodDelete {
		return Query(r, v, flags...)
	}
	return Body(r, v, flags...)
}

func Query(r *http.Request, v any, flags ...Flag) error {
	return DecodeQuery(r.URL.Query(), v, flags...)
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
		return DecodeForm(r.Form, v, flags...)
	}
	return nil
}

func Header(r *http.Request, v any, flags ...Flag) error {
	return DecodeHeader(r.Header, v, flags...)
}

func Path(r *http.Request, v any, flags ...Flag) error {
	if PathValueFunc == nil {
		return errors.New("PathValueFunc not set")
	}

	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return &form.InvalidDecoderError{Type: reflect.TypeOf(v)}
	}

	return setPath(r, val)
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

func setPath(r *http.Request, val reflect.Value) error {
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil
	}

	t := val.Type()

	// TODO cache this
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Anonymous {
			setPath(r, val.Field(i))
			continue
		}
		pathParam := field.Tag.Get("path")
		if pathParam != "" && pathParam != "-" {
			if err := setField(field.Type.Kind(), PathValueFunc(r, pathParam), val.Field(i)); err != nil {
				return err
			}
		}
	}

	return nil
}

// code below is mostly taken from Echo's bind implementation
func setField(kind reflect.Kind, strVal string, field reflect.Value) error {
	switch kind {
	case reflect.Ptr:
		if field.IsNil() {
			newVal := reflect.New(field.Type().Elem())
			err := setField(newVal.Elem().Kind(), strVal, newVal.Elem())
			if err == nil {
				field.Set(newVal)
			}
			return err
		}
		return setField(field.Elem().Kind(), strVal, field.Elem())
	case reflect.Int:
		return setIntField(strVal, 0, field)
	case reflect.Int8:
		return setIntField(strVal, 8, field)
	case reflect.Int16:
		return setIntField(strVal, 16, field)
	case reflect.Int32:
		return setIntField(strVal, 32, field)
	case reflect.Int64:
		return setIntField(strVal, 64, field)
	case reflect.Uint:
		return setUintField(strVal, 0, field)
	case reflect.Uint8:
		return setUintField(strVal, 8, field)
	case reflect.Uint16:
		return setUintField(strVal, 16, field)
	case reflect.Uint32:
		return setUintField(strVal, 32, field)
	case reflect.Uint64:
		return setUintField(strVal, 64, field)
	case reflect.Bool:
		return setBoolField(strVal, field)
	case reflect.Float32:
		return setFloatField(strVal, 32, field)
	case reflect.Float64:
		return setFloatField(strVal, 64, field)
	case reflect.String:
		field.SetString(strVal)
	default:
		return errors.New("unknown type")
	}
	return nil
}

func setIntField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0"
	}
	intVal, err := strconv.ParseInt(val, 10, bitSize)
	if err == nil {
		field.SetInt(intVal)
	}
	return err
}

func setUintField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0"
	}
	uintVal, err := strconv.ParseUint(val, 10, bitSize)
	if err == nil {
		field.SetUint(uintVal)
	}
	return err
}

func setBoolField(val string, field reflect.Value) error {
	if val == "" {
		val = "false"
	}
	boolVal, err := strconv.ParseBool(val)
	if err == nil {
		field.SetBool(boolVal)
	}
	return err
}

func setFloatField(val string, bitSize int, field reflect.Value) error {
	if val == "" {
		val = "0.0"
	}
	floatVal, err := strconv.ParseFloat(val, bitSize)
	if err == nil {
		field.SetFloat(floatVal)
	}
	return err
}
