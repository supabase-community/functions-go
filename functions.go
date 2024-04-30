package functions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (c *Client) Invoke(functionName string, payload interface{}) (string, error) {
	// Marshal the payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	// Build the URL and create the request
	url := c.clientTransport.baseUrl.String() + "/" + functionName
	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return "", err
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Execute the request using the client's session
	resp, err := c.session.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read the response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Check HTTP response status code
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("server responded with error: %s", resp.Status)
	}

	return string(responseBody), nil
}
