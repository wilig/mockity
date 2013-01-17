package mockity

import (
	"bytes"
	//"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

type StringResponseWriter struct {
	header http.Header
	output *bytes.Buffer
}

func (s StringResponseWriter) Write(buf []byte) (int, error) {
	return s.output.Write(buf)
}

func (s StringResponseWriter) Header() http.Header {
	return s.header
}

func (s StringResponseWriter) WriteHeader(status int) {
	return
}

func TestContains(t *testing.T) {
	if !contains([]string{"one", "two", "three"}, "one") {
		t.Fail()
	}
	if contains([]string{"one", "two", "three"}, "four") {
		t.Fail()
	}
}

func TestWriteResponse(t *testing.T) {
	v := "Testing"
	w := StringResponseWriter{output: &bytes.Buffer{}}
	//	r := http.Request{URL: &url.URL{Path: "/"}}
	r := Response{Body: v}
	writeResponse(w, r)
	if w.output.String() != v {
		t.Fail()
	}
}

func TestBaseRedirection(t *testing.T) {
	loc := "/_mockity_infinite_redirector/1"
	w := StringResponseWriter{output: &bytes.Buffer{}, header: http.Header{}}
	r := http.Request{URL: &url.URL{Path: "/"}, Method: "GET"}
	redirectEndlessly(w, &r)
	if !strings.Contains(w.output.String(), loc) {
		t.Error("Redirector not writing HTML link text")
	}
	if w.header.Get("Location") != loc {
		t.Error("Location header not set correctly")
	}
}

func TestIncrementalRedirection(t *testing.T) {
	loc := "/_mockity_infinite_redirector/5"
	w := StringResponseWriter{output: &bytes.Buffer{}, header: http.Header{}}
	r := http.Request{URL: &url.URL{Path: "/_mockity_infinite_redirector/4"}, Method: "GET"}
	redirectEndlessly(w, &r)
	if !strings.Contains(w.output.String(), loc) {
		t.Error("Redirector not writing HTML link text")
	}
	if w.header.Get("Location") != loc {
		t.Error("Redirector not incrementing location")
	}
}

func TestGetBody(t *testing.T) {
	v := "I shall now mock you"
	r := Response{Body: v}
	body := getBody(r)
	if string(body) != v {
		t.Error("getBody garbled the body")
	}
}

func TestGetBodyHandlesFiles(t *testing.T) {
	r := Response{Body: "!file:README.md"}
	body := getBody(r)
	if body[0] == '!' {
		t.Error("Response file was not read")
	}
}

func TestRouteMatching(t *testing.T) {
	route := Route{URL: "/test", Method: "GET"}
	r := http.Request{URL: &url.URL{Path: "/test"}, Method: "GET"}
	if !route.matches(&r) {
		t.Error("Failed to match simple URL")
	}
	// Should not match for missing parameters
	route = Route{URL: "/test", Method: "GET", Params: map[string]string{"name": "mockity"}}
	r = http.Request{URL: &url.URL{Path: "/test"}, Method: "GET"}
	if route.matches(&r) {
		t.Error("Matched when Request parameters were not present")
	}
	// Should match with parameters
	formValues := url.Values{}
	formValues.Add("name", "mockity")
	route = Route{URL: "/test", Method: "GET", Params: map[string]string{"name": "mockity"}}
	r = http.Request{URL: &url.URL{Path: "/test"}, Method: "GET", Form: formValues}
	if !route.matches(&r) {
		t.Error("Failed to match on Request parameters")
	}
	// Should match with parameters
	formValues = url.Values{}
	formValues.Add("name", "mockitie")
	route = Route{URL: "/test", Method: "GET", Params: map[string]string{"name": "mockity"}}
	r = http.Request{URL: &url.URL{Path: "/test"}, Method: "GET", Form: formValues}
	if route.matches(&r) {
		t.Error("Matched on differing parameter values")
	}
	// Should not match for missing headers
	route = Route{URL: "/test", Method: "GET",
		Headers: map[string][]string{"Accept": []string{"application/json"}}}
	r = http.Request{URL: &url.URL{Path: "/test"}, Method: "GET"}
	if route.matches(&r) {
		t.Error("Matched on absent headers")
	}
	// Should not match for differing values
	headers := http.Header{}
	headers.Add("Accept", "text/html")
	route = Route{URL: "/test", Method: "GET",
		Headers: map[string][]string{"Accept": []string{"application/json"}}}
	r = http.Request{URL: &url.URL{Path: "/test"}, Method: "GET", Header: headers}
	if route.matches(&r) {
		t.Error("Matched on differing header values")
	}
	// Should match with headers
	headers = http.Header{}
	headers.Add("Accept", "application/json")
	route = Route{URL: "/test", Method: "GET",
		Headers: map[string][]string{"Accept": []string{"application/json"}}}
	r = http.Request{URL: &url.URL{Path: "/test"}, Method: "GET", Header: headers}
	if !route.matches(&r) {
		t.Error("Failed to match on Request headers")
	}
}
