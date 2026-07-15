package main

import (
	"net/url"
	"strings"
)

// parseURL separates a URL into scheme, netloc, path, query, and fragment,
// mirroring urlsplit from the Python source.
func parseURL(rawURL string) (scheme, netloc, path, query, fragment string) {
	if !strings.Contains(rawURL, "://") {
		rawURL = "http://" + rawURL
	}
	u, err := url.Parse(rawURL)
	if err != nil || u == nil {
		return "", "", "", "", ""
	}
	scheme = u.Scheme
	netloc = u.Host
	path = u.Path
	if path == "" {
		path = "/"
	}
	query = u.RawQuery
	fragment = u.Fragment
	return
}

// joinURL reconstructs a URL from its components, mirroring urlunsplit.
func joinURL(scheme, netloc, path, query, fragment string) string {
	u := &url.URL{
		Scheme:   scheme,
		Host:     netloc,
		Path:     path,
		RawQuery: query,
		Fragment: fragment,
	}
	return u.String()
}

// getDomain returns the netloc of a URL, mirroring get_domain.
func getDomain(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u == nil {
		return ""
	}
	return u.Host
}

// parseQueryString parses a query string into a map of key→[]value,
// keeping blank values, mirroring parse_qs(keep_blank_values=True).
func parseQueryString(query string) map[string][]string {
	values, err := url.ParseQuery(query)
	if err != nil {
		return map[string][]string{}
	}
	if values == nil {
		return map[string][]string{}
	}
	return values
}

// encodeQuery encodes a map of key→[]value into a query string,
// mirroring urlencode(..., doseq=True).
func encodeQuery(params map[string][]string) string {
	values := url.Values(params)
	return values.Encode()
}

// generatePayloadUrls generates URL combinations for XSS testing, mirroring
// the Python generate_payload_urls function exactly: substitute each query
// param value, then fragment params, then append ?test=payload + fragment
// payload if neither query nor fragment exist.
func generatePayloadUrls(rawURL, payload string) []string {
	var combinations []string
	scheme, netloc, path, queryString, fragment := splitURL(rawURL)
	if scheme == "" {
		scheme = "http"
	}

	queryParams := parseQueryString(queryString)
	for key := range queryParams {
		modified := map[string][]string{}
		for k, v := range queryParams {
			modified[k] = v
		}
		modified[key] = []string{payload}
		modifiedQuery := encodeQuery(modified)
		modifiedURL := joinURL(scheme, netloc, path, modifiedQuery, fragment)
		combinations = append(combinations, modifiedURL)
	}

	if fragment != "" {
		if strings.Contains(fragment, "=") {
			fragParams := parseQueryString(fragment)
			for key := range fragParams {
				modified := map[string][]string{}
				for k, v := range fragParams {
					modified[k] = v
				}
				modified[key] = []string{payload}
				modifiedFragment := encodeQuery(modified)
				modifiedURL := joinURL(scheme, netloc, path, queryString, modifiedFragment)
				combinations = append(combinations, modifiedURL)
			}
		} else {
			modifiedURL := joinURL(scheme, netloc, path, queryString, payload)
			combinations = append(combinations, modifiedURL)
		}
	}

	if len(queryParams) == 0 && fragment == "" {
		newQuery := encodeQuery(map[string][]string{"test": {payload}})
		modifiedURL := joinURL(scheme, netloc, path, newQuery, fragment)
		combinations = append(combinations, modifiedURL)

		modifiedURLFragment := joinURL(scheme, netloc, path, queryString, payload)
		combinations = append(combinations, modifiedURLFragment)
	}

	return combinations
}

// splitURL splits a URL string into (scheme, netloc, path, query, fragment),
// without adding a default scheme. Mirrors urlsplit.
func splitURL(rawURL string) (scheme, netloc, path, query, fragment string) {
	u, err := url.Parse(rawURL)
	if err != nil || u == nil {
		return "", "", "", "", ""
	}
	scheme = u.Scheme
	netloc = u.Host
	path = u.Path
	query = u.RawQuery
	fragment = u.Fragment
	return
}

// urlNetloc returns just the host:port portion of a URL, lower-cased.
func urlNetloc(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u == nil {
		return ""
	}
	return strings.ToLower(u.Host)
}

// addScheme prefixes http:// if the input URL has no scheme.
func addScheme(rawURL string) string {
	if !strings.Contains(rawURL, "://") {
		return "http://" + rawURL
	}
	return rawURL
}
