package unleash

import (
	"net/http"
	"time"
)

func CreateClient(baseURL string, authorizationToken string) (ClientWithResponsesInterface, error) {
	hc := http.Client{
		Timeout: 60 * time.Second,
	}
	hc.Transport = authHeaderTransport{
		roundTripper:       http.DefaultTransport,
		authorizationToken: authorizationToken,
	}

	c, err := NewClientWithResponses(baseURL, WithHTTPClient(&hc))
	if err != nil {
		return nil, err
	}

	return c, nil
}

type authHeaderTransport struct {
	roundTripper       http.RoundTripper
	authorizationToken string
}

func (t authHeaderTransport) RoundTrip(req *http.Request) (res *http.Response, e error) {
	req.Header.Set("Authorization", t.authorizationToken)

	return t.roundTripper.RoundTrip(req)
}
