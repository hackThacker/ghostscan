package main

import (
	"net/url"
	"strings"
)

// urlParseRaw is a thin wrapper to keep the OR/XSS scanner code decoupled from
// the net/url package. Returns *url.URL or an error.
func urlParseRaw(s string) (*url.URL, error) {
	if !strings.Contains(s, "://") {
		s = "http://" + s
	}
	return url.Parse(s)
}

// urlBuildFromParts constructs a URL string from the parts.
func urlBuildFromParts(scheme, host, path, rawQuery, fragment string) string {
	u := &url.URL{
		Scheme:   scheme,
		Host:     host,
		Path:     path,
		RawQuery: rawQuery,
		Fragment: fragment,
	}
	return u.String()
}
