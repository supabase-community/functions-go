package functions

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func (c *Client) Invoke(functionName string, method string, payload interface{}) ([]byte, error) {
	// Build the base URL
	targetURLString := c.clientTransport.baseUrl.String() + "/" + functionName

	var req *http.Request
	var err error

	// Normalize method to uppercase for reliable comparison
	httpMethod := strings.ToUpper(method)

	switch httpMethod {
	case "POST":
		jsonData, marshalErr := json.Marshal(payload)
		if marshalErr != nil {
			return nil, fmt.Errorf("failed to marshal payload for POST request: %w", marshalErr)
		}
		req, err = http.NewRequest(httpMethod, targetURLString, bytes.NewReader(jsonData))
		if err != nil {
			return nil, fmt.Errorf("failed to create POST request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

	case "GET":
		parsedURL, parseErr := url.Parse(targetURLString)
		if parseErr != nil {
			return nil, fmt.Errorf("failed to parse base URL for GET request: %w", parseErr)
		}

		if payload != nil {
			var queryValues url.Values
			switch p := payload.(type) {
			case map[string]string:
				queryValues = make(url.Values)
				for k, v := range p {
					queryValues.Set(k, v)
				}
			case map[string]interface{}:
				queryValues = make(url.Values)
				for k, v := range p {
					queryValues.Set(k, fmt.Sprint(v))
				}
			default:
				if payload != nil {
					return nil, fmt.Errorf("for GET requests, payload must be map[string]string or map[string]interface{} for query parameters, got %T", payload)
				}
			}

			if len(queryValues) > 0 {
				parsedURL.RawQuery = queryValues.Encode()
			}
		}

		req, err = http.NewRequest(httpMethod, parsedURL.String(), nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create GET request: %w", err)
		}

	default:
		return nil, fmt.Errorf("unsupported HTTP method: %s. Only GET and POST are supported", method)
	}

	// Set common headers (if any) applicable to all processed methods
	req.Header.Set("Accept", "application/json")

	// Execute the request using the client's session
	resp, execErr := c.session.Do(req)
	if execErr != nil {
		return nil, fmt.Errorf("failed to execute request: %w", execErr)
	}
	defer resp.Body.Close()

	// Read the response body
	responseBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("failed to read response body: %w", readErr)
	}

	// Check HTTP response status code
	if resp.StatusCode >= 400 {
		errorMsg := fmt.Sprintf("server responded with error: %s (status code %d)", resp.Status, resp.StatusCode)
		if len(responseBody) > 0 {
			errorMsg += fmt.Sprintf(", body: %s", string(responseBody))
		}
		return nil, fmt.Errorf(errorMsg)
	}

	return responseBody, nil
}
