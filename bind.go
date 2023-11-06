// Package bind contains convenience functions to decode HTTP request data.
package bind

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/http"
	"net/url"
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
	pathDecoder   = form.NewDecoder()

	queryEncoder  = form.NewEncoder()
	formEncoder   = form.NewEncoder()
	headerEncoder = form.NewEncoder()

	PathFunc func(*http.Request) url.Values
)

func init() {
	queryDecoder.SetTagName("query")
	queryDecoder.SetMode(form.ModeExplicit)
	formDecoder.SetTagName("form")
	formDecoder.SetMode(form.ModeExplicit)
	headerDecoder.SetTagName("header")
	headerDecoder.SetMode(form.ModeExplicit)
	pathDecoder.SetTagName("path")
	pathDecoder.SetMode(form.ModeExplicit)

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

func Request(r *http.Request, v any, flags ...Flag) error {
	if PathFunc != nil {
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
	if PathFunc == nil {
		return errors.New("PathFunc not set")
	}

	vals := PathFunc(r)
	if hasFlag(flags, Vacuum) {
		vals = vacuum(vals)
	}
	return pathDecoder.Decode(v, vals)
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
