package bpm

import "net/http"

// basicAuthTransport is the struct that handles basic auth.
type basicAuthTransport struct {
	username string
	password string
	parent   http.RoundTripper
}

// newBasicAuthTransport creates a new http.RoundTripper which sets basic auth to all requests.
// The parent parameter is the roundtripper which becomes the request after basic auth is added.
// If nil the http.DefaultTransport will be used.
func newBasicAuthTransport(username, password string, parent http.RoundTripper) *basicAuthTransport {
	if parent == nil {
		parent = http.DefaultTransport
	}
	return &basicAuthTransport{
		username: username,
		password: password,
		parent:   parent,
	}
}

// RoundTrip implements the http.RoundTripper interface.
// This method adds basic auth to all requests.
func (transport *basicAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.SetBasicAuth(transport.username, transport.password)
	return transport.parent.RoundTrip(req)
}
