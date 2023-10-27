// Package bind contains convenience functions to decode HTTP request data.
package bind

import (
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
	if err := Header(r, v, flags...); err != nil {
		return err
	}
	if r.Method == http.MethodGet || r.Method == http.MethodDelete || r.Method == http.MethodHead {
		return Query(r, v, flags...)
	}
	return Form(r, v, flags...)
}

func Query(r *http.Request, v any, flags ...Flag) error {
	vals := r.URL.Query()
	if hasFlag(flags, Vacuum) {
		vals = vacuum(vals)
	}
	return queryDecoder.Decode(v, vals)
}

func Form(r *http.Request, v any, flags ...Flag) error {
	r.ParseForm()
	vals := r.Form
	if hasFlag(flags, Vacuum) {
		vals = vacuum(vals)
	}
	return formDecoder.Decode(v, vals)
}

func Header(r *http.Request, v any, flags ...Flag) error {
	vals := url.Values(r.Header)
	if hasFlag(flags, Vacuum) {
		vals = vacuum(vals)
	}
	return headerDecoder.Decode(v, vals)
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
