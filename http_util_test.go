package bpm

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type basicAuthTransportTest struct {
	username string
	password string
	t        *testing.T
}

func (transport *basicAuthTransportTest) RoundTrip(req *http.Request) (*http.Response, error) {
	username, password, ok := req.BasicAuth()
	if assert.True(transport.t, ok, "request should contain basic auth") {
		assert.Equal(transport.t, transport.username, username)
		assert.Equal(transport.t, transport.password, password)
	}
	return &http.Response{}, nil
}

func TestBasicAuthHTTPRoundtripper(t *testing.T) {
	expectedUser := "testUser"
	expectedPassword := "testPassword"

	roundTripper := newBasicAuthTransport(expectedUser, expectedPassword, nil)
	assert.Equal(t, roundTripper.parent, http.DefaultTransport, "default roundtripper should be set with nil parent")

	dummyTransport := &basicAuthTransportTest{
		username: expectedUser,
		password: expectedPassword,
		t:        t,
	}
	roundTripper = newBasicAuthTransport(expectedUser, expectedPassword, dummyTransport)
	request, err := http.NewRequest(http.MethodGet, "http://localhost", nil)
	if assert.NoError(t, err) {
		_, err := roundTripper.RoundTrip(request)
		assert.NoError(t, err)
	}
}
