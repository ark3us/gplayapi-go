package gplayapi

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

// stubTransport answers every request with a canned response, or fails the
// round trip outright when err is set.
type stubTransport struct {
	status int
	body   string
	err    error
}

func (s stubTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &http.Response{
		StatusCode: s.status,
		Body:       http.NoBody,
		Request:    r,
	}, nil
}

func withTransport(t *testing.T, rt http.RoundTripper) {
	t.Helper()
	original := httpClient
	httpClient = &http.Client{Transport: rt}
	t.Cleanup(func() { httpClient = original })
}

func testClient() *GooglePlayClient {
	return &GooglePlayClient{
		AuthData:   &AuthData{},
		DeviceInfo: Pixel8,
		Locale:     "en_GB",
		LocaleDash: "en-GB",
		Country:    "uk",
	}
}

// An empty 200 unmarshals into a ResponseWrapper with no payload, so
// _doAuthedReq used to hand back (nil, nil). Callers that read the payload
// then dereferenced nil. Play answering with nothing already has a name here
// -- ErrNilPayload -- and every caller must get it.
func TestDoAuthedReqReportsAnEmptyPayloadAsAnError(t *testing.T) {
	withTransport(t, stubTransport{status: 200})
	r, _ := http.NewRequest("GET", UrlToc, nil)

	payload, err := testClient().doAuthedReq(r)
	if !errors.Is(err, ErrNilPayload) {
		t.Fatalf("err = %v, want ErrNilPayload", err)
	}
	if payload != nil {
		t.Fatalf("payload = %v, want nil", payload)
	}
}

// The panic seen in production on 2026-09-10: toc() read payload.TocResponse
// without checking payload, so an empty answer during client construction
// took down the caller instead of returning an error it could handle.
func TestTocReturnsAnErrorInsteadOfPanicking(t *testing.T) {
	withTransport(t, stubTransport{status: 200})

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("toc panicked instead of returning an error: %v", r)
		}
	}()
	if _, err := testClient().toc(); err == nil {
		t.Fatal("toc must report an empty answer as an error")
	}
}

// GenerateGPToken discarded its own transport error and returned an empty
// token with err == nil. RegenerateGPToken stored that empty token, the next
// request went out unauthenticated, and the failure resurfaced far away as a
// nil payload.
func TestGenerateGPTokenPropagatesTransportErrors(t *testing.T) {
	wanted := errors.New("connection reset")
	withTransport(t, stubTransport{err: wanted})

	token, err := testClient().GenerateGPToken()
	if err == nil {
		t.Fatal("a failed request must not be reported as a successful token")
	}
	if !strings.Contains(err.Error(), wanted.Error()) {
		t.Errorf("err = %v, want it to carry %v", err, wanted)
	}
	if token != "" {
		t.Errorf("token = %q, want empty", token)
	}
}
