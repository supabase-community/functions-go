package functions

import (
	"net/http"
	"net/url"
)

var (
	version = "1.0.0-dev.1"
)

type Client struct {
	clientError     error
	session         http.Client
	clientTransport transport
}

type transport struct {
	header  http.Header
	baseUrl url.URL
}

func (t transport) RoundTrip(request *http.Request) (*http.Response, error) {
	for headerName, values := range t.header {
		for _, val := range values {
			request.Header.Add(headerName, val)
		}
	}
	request.URL = t.baseUrl.ResolveReference(request.URL)
	return http.DefaultTransport.RoundTrip(request)
}

// TokenAuth sets authorization headers for subsequent requests.
func (c *Client) TokenAuth(token string) *Client {
	c.clientTransport.header.Set("Authorization", "Bearer "+token)
	//c.clientTransport.header.Set("apikey", token)
	return c
}

/*
NewClient constructs a new client given a URL to a functions-go instance

Usage:
	client := functions.NewClient("https://abc.functions.supabase.co", "<service-token>", nil)

Inspired By Postgrest and storage-go.
*/
func NewClient(rawUrl string, token string, headers map[string]string) *Client {
	baseURL, err := url.Parse(rawUrl)
	if err != nil {
		return &Client{
			clientError: err,
		}
	}

	t := transport{
		header:  http.Header{},
		baseUrl: *baseURL,
	}

	c := Client{
		session:         http.Client{Transport: t},
		clientTransport: t,
	}

	// Set required headers
	c.clientTransport.header.Set("Accept", "application/json")
	c.clientTransport.header.Set("Content-Type", "application/json")
	c.clientTransport.header.Set("X-Client-Info", "functions-go/"+version)
	c.clientTransport.header.Set("Authorization", "Bearer "+token)

	// Optional headers [if exists]
	for key, value := range headers {
		c.clientTransport.header.Set(key, value)
	}

	return &c
}
