// Copyright 2013 - Will Groppe.  All rights reserved.
// 
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package mockity

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Route struct {
	URL      string            `json:"url"`
	Method   string            `json:"method"`
	Headers  Header            `json:"headers"`
	Params   map[string]string `json:"params"`
	Response Response          `json:"response"`
}

// Response specifies the what to return when a Route is matched.
// The ContentType, SetCookie entries are shortcuts to frequently
// used Headers.
type Response struct {
	Headers     Header    `json:"headers"`
	ContentType string    `json:"content-type"`
	StatusCode  int       `json:"status"`
	SetCookie   Cookie    `json:"cookies"`
	Body        string    `json:"body"`
	Directive   Directive `json:"!directive"`
}

// Directive denotes special handling of a Response.  
type Directive struct {
	Delay        int  `json:"delay"`
	Partial      bool `json:"partial"`
	Firehose     bool `json:"firehose"`
	Flaky        bool `json:"flaky"`
	RedirectLoop bool `json:"loop"`
}

type Header map[string][]string

type Cookie map[string]string

// MakeMockery returns a HTTP Handler that matches the supplied routes to
// http.Requests.  Unmatched requests are logged and returned as 
// 404.
func MakeMockery(routes []Route) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/_mockity_infinite_redirector") {
			redirectEndlessly(w, r)
			return
		}
		for _, route := range routes {
			if route.matches(r) {
				route.respond(w, r)
				return
			}
		}
		log.Printf("Unmatched: %s \"%s\"", r.Method, r.URL)
		http.NotFound(w, r)
	}
}

// matches determines if a HTTP request matches a defined route.  It
// checks the path, method, headers, and query/form parameters. 
// It returns a boolean value indicating whether the route machtes the
// request.
func (route Route) matches(r *http.Request) bool {
	if route.URL != r.URL.Path {
		return false
	}
	if route.Method != r.Method {
		return false
	}
	for header, values := range route.Headers {
		if r.Header[header] == nil {
			return false
		} else {
			req_values := r.Header[header]
			for _, value := range values {
				if !contains(req_values, value) {
					return false
				}
			}
		}
	}
	for param, value := range route.Params {
		if r.FormValue(param) != value {
			return false
		}
	}
	return true
}

func (route Route) respond(w http.ResponseWriter, r *http.Request) {
	resp := route.Response
	switch {
	case resp.Directive.Delay > 0:
		d := time.Duration(resp.Directive.Delay) * time.Millisecond
		writeDelayedResponse(w, d, resp)
		return
	case resp.Directive.Delay < 0:
		d := time.Duration(100000) * time.Hour
		writeDelayedResponse(w, d, resp)
		return
	case resp.Directive.Partial:
		writePartialResponse(w, resp)
		return
	case resp.Directive.RedirectLoop:
		redirectEndlessly(w, r)
		return
	case resp.Directive.Firehose:
		writeInfiniteStream(w)
		return
	case resp.Directive.Flaky:
		maybeWriteResponse(w, resp)
		return
	}
	writeResponse(w, resp)
	return
}

func writeResponse(w http.ResponseWriter, resp Response) {
	setHeaders(w.Header(), resp)
	// Must happen last as changes to headers are ignored after
	// WriteHeader is called.
	if resp.StatusCode != 0 {
		w.WriteHeader(resp.StatusCode)
	}
	w.Write(getBody(resp))
}

func writeDelayedResponse(w http.ResponseWriter, d time.Duration, resp Response) {
	time.Sleep(d)
	writeResponse(w, resp)
}

func writePartialResponse(w http.ResponseWriter, resp Response) {
	partialResponse := resp
	body := getBody(resp)
	length := len(body)
	if length > 2 {
		total := rand.Intn(len(resp.Body) - 2)
		partialResponse.Body = string(body[:total+1])
	}
	writeResponse(w, partialResponse)
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported, cannot do partial responses", http.StatusInternalServerError)
		return
	}
	conn, bufrw, err := hj.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	bufrw.Flush()
	conn.Close()
}

// redirectEndlessly calls http.Redirect with an incrementing counter
// so the redirects are never circular.  Calling it directly with
// something other then a number as the last part of the URL will 
// result in it resetting back to one.
func redirectEndlessly(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/_mockity_infinite_redirector") {
		http.Redirect(w, r, "/_mockity_infinite_redirector/1", 301)
	} else {
		strCount := strings.Trim(r.URL.Path, "/_mockity_infinite_redirector/")
		redirectCount, err := strconv.ParseInt(strCount, 0, 64)
		if err != nil {
			http.Redirect(w, r, "/_mockity_infinite_redirector/1", 301)
		}
		url := fmt.Sprintf("/_mockity_infinite_redirector/%d", redirectCount+1)
		http.Redirect(w, r, url, 301)
	}
}

func writeInfiniteStream(w http.ResponseWriter) {
	fmt.Fprint(w, "On, ")
	for {
		fmt.Fprint(w, "and on, and on, ")
		time.Sleep(time.Duration(5) * time.Millisecond)
	}
}

// maybeWriteResponse flips a coin (ok generates a random number)
// to determine whether to return the correct response or throw
// a 500 Server Error.
func maybeWriteResponse(w http.ResponseWriter, resp Response) {
	randSrc := rand.New(rand.NewSource(time.Now().Unix()))
	if randSrc.Float64() > 0.4 {
		writeResponse(w, resp)
	} else {
		http.Error(w, "Server Error: I'm being flaky!", http.StatusInternalServerError)
	}
}

func setHeaders(headers http.Header, resp Response) {
	for name, values := range resp.Headers {
		for _, value := range values {
			headers.Add(name, value)
		}
	}
	// Short cuts may overwrite previously set headers. 
	if resp.ContentType != "" {
		headers.Add("Content-Type", resp.ContentType)
	}
	if len(resp.SetCookie) > 0 {
		for name, value := range resp.SetCookie {
			c := &http.Cookie{Name: name, Value: value}
			headers.Add("Set-Cookie", c.String())
		}
	}
}

// getBody returns the Body of the response as a byte slice.  If the
// Body string starts with "!file:" then the response is read from 
// filename provided.
func getBody(r Response) []byte {
	if strings.HasPrefix(r.Body, "!file:") {
		filename := strings.Trim(r.Body, "!file:")
		body, err := ioutil.ReadFile(filename)
		if err != nil {
			msg := fmt.Sprintf("Error reading response file: %s [%s]", filename, err.Error())
			log.Printf(msg)
			return []byte(msg)
		}
		return body
	}
	return []byte(r.Body)
}

// contains checks for the presence of a string in a slice of strings.
// It returns a boolean indicating inclusion.
func contains(set []string, s string) bool {
	for _, i := range set {
		if i == s {
			return true
		}
	}
	return false
}
