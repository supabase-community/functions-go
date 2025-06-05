package functions

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// MockRoundTripper is a custom http.RoundTripper for mocking server responses.
type MockRoundTripper func(req *http.Request) (*http.Response, error)

func (mrt MockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return mrt(req)
}

// NOTE: The Client and transport structs are defined in client.go and are used here.
// Ensure client.go is part of the same 'functions' package.

// Helper function to create a new client with a mock server
func newTestClient(handler http.HandlerFunc) (*Client, *httptest.Server) {
	server := httptest.NewServer(handler)
	serverURL, _ := url.Parse(server.URL) // serverURL is *url.URL

	// Create a custom http.Client that uses the test server
	mockSessionTransport := MockRoundTripper(func(req *http.Request) (*http.Response, error) {
		// Ensure the request URL uses the test server's host and scheme
		// The Invoke function will build the full URL using clientTransport.baseUrl
		// So, the RoundTripper for the session just needs to execute it.
		// However, the httptest.Server expects requests to its specific URL.
		// We need to make sure the request passed to DefaultTransport has the test server's URL.
		finalReqURL := *serverURL // Dereference to get url.URL, then take parts
		finalReqURL.Path = req.URL.Path
		finalReqURL.RawQuery = req.URL.RawQuery
		req.URL = &finalReqURL
		return http.DefaultTransport.RoundTrip(req)
	})

	// The Client's session needs a transport. The mockSessionTransport ensures
	// that requests made by the session are correctly routed to the httptest.Server.

	return &Client{
		// clientError can be nil
		clientTransport: transport{ // transport.baseUrl is url.URL
			header:  http.Header{},
			baseUrl: *serverURL, // serverURL from httptest.NewServer is *url.URL
		},
		session: http.Client{Transport: mockSessionTransport}, // Client.session is http.Client
	}, server
}

func TestClient_Invoke_POST_Success(t *testing.T) {
	expectedResponseString := `{"message": "success"}`
	expectedPayload := map[string]string{"key": "value"}
	functionName := "testPostFunc"

	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
			http.Error(w, "Wrong method", http.StatusMethodNotAllowed)
			return
		}
		// The path check in the handler should be against the functionName only,
		// as the test server's base URL is handled by the RoundTripper or client setup.
		if r.URL.Path != "/"+functionName {
			t.Errorf("Expected path /%s, got %s", functionName, r.URL.Path)
			http.Error(w, "Wrong path", http.StatusBadRequest)
			return
		}
		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", contentType)
			http.Error(w, "Wrong content type", http.StatusBadRequest)
			return
		}

		var receivedPayload map[string]string
		err := json.NewDecoder(r.Body).Decode(&receivedPayload)
		if err != nil {
			t.Errorf("Failed to decode request body: %v", err)
			http.Error(w, "Bad request body", http.StatusBadRequest)
			return
		}
		if receivedPayload["key"] != expectedPayload["key"] {
			t.Errorf("Expected payload %v, got %v", expectedPayload, receivedPayload)
			http.Error(w, "Wrong payload", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, expectedResponseString)
	})
	defer server.Close()

	responseBytes, err := client.Invoke(functionName, "POST", expectedPayload)
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}
	if string(responseBytes) != expectedResponseString {
		t.Errorf("Expected response %s, got %s", expectedResponseString, string(responseBytes))
	}
}

func TestClient_Invoke_GET_Success_NoParams(t *testing.T) {
	expectedResponseString := `{"data": "some data"}`
	functionName := "testGetFunc"

	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
			http.Error(w, "Wrong method", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/"+functionName {
			t.Errorf("Expected path /%s, got %s", functionName, r.URL.Path)
			http.Error(w, "Wrong path", http.StatusBadRequest)
			return
		}
		if r.URL.RawQuery != "" {
			t.Errorf("Expected no query params, got %s", r.URL.RawQuery)
			http.Error(w, "Unexpected query params", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, expectedResponseString)
	})
	defer server.Close()

	responseBytes, err := client.Invoke(functionName, "GET", nil)
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}
	if string(responseBytes) != expectedResponseString {
		t.Errorf("Expected response %s, got %s", expectedResponseString, string(responseBytes))
	}
}

func TestClient_Invoke_GET_Success_WithParams_MapStringString(t *testing.T) {
	expectedResponseString := `{"data": "filtered data"}`
	functionName := "testGetFiltered"
	params := map[string]string{"filter": "active", "limit": "10"}

	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
			return
		}
		if r.URL.Path != "/"+functionName {
			t.Errorf("Expected path /%s, got %s", functionName, r.URL.Path)
			return
		}
		query := r.URL.Query()
		if query.Get("filter") != params["filter"] {
			t.Errorf("Expected query param filter=%s, got %s", params["filter"], query.Get("filter"))
		}
		if query.Get("limit") != params["limit"] {
			t.Errorf("Expected query param limit=%s, got %s", params["limit"], query.Get("limit"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, expectedResponseString)
	})
	defer server.Close()

	responseBytes, err := client.Invoke(functionName, "GET", params)
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}
	if string(responseBytes) != expectedResponseString {
		t.Errorf("Expected response %s, got %s", expectedResponseString, string(responseBytes))
	}
}

func TestClient_Invoke_GET_Success_WithParams_MapStringInterface(t *testing.T) {
	expectedResponseString := `{"data": "interface data"}`
	functionName := "testGetInterfaceParams"
	params := map[string]interface{}{"id": 123, "active": true}

	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
			return
		}
		query := r.URL.Query()
		if query.Get("id") != "123" {
			t.Errorf("Expected query param id=123, got %s", query.Get("id"))
		}
		if query.Get("active") != "true" {
			t.Errorf("Expected query param active=true, got %s", query.Get("active"))
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, expectedResponseString)
	})
	defer server.Close()

	responseBytes, err := client.Invoke(functionName, "GET", params)
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}
	if string(responseBytes) != expectedResponseString {
		t.Errorf("Expected response %s, got %s", expectedResponseString, string(responseBytes))
	}
}

func TestClient_Invoke_POST_MarshalError(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Server handler should not be called on marshal error")
		http.Error(w, "Should not be reached", http.StatusInternalServerError)
	})
	defer server.Close()

	invalidPayload := make(chan int)
	_, err := client.Invoke("testFunc", "POST", invalidPayload)

	if err == nil {
		t.Fatal("Expected an error from Invoke due to marshaling failure, but got nil")
	}
	if !strings.Contains(err.Error(), "failed to marshal payload for POST request") {
		t.Errorf("Expected marshal error message, got: %v", err)
	}
}

func TestClient_Invoke_GET_InvalidPayloadType(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Server handler should not be called on invalid payload type for GET")
		http.Error(w, "Should not be reached", http.StatusInternalServerError)
	})
	defer server.Close()

	invalidPayload := []string{"this", "is", "not", "a", "map"}
	_, err := client.Invoke("testGetInvalid", "GET", invalidPayload)

	if err == nil {
		t.Fatal("Expected an error due to invalid payload type for GET, but got nil")
	}
	expectedErrorMsg := "for GET requests, payload must be map[string]string or map[string]interface{} for query parameters, got []string"
	if !strings.Contains(err.Error(), expectedErrorMsg) {
		t.Errorf("Expected error message '%s', got: %v", expectedErrorMsg, err)
	}
}

func TestClient_Invoke_UnsupportedMethod(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Server handler should not be called for unsupported method")
		http.Error(w, "Should not be reached", http.StatusInternalServerError)
	})
	defer server.Close()

	_, err := client.Invoke("testFunc", "PUT", nil)
	if err == nil {
		t.Fatal("Expected an error for unsupported HTTP method, but got nil")
	}
	if !strings.Contains(err.Error(), "unsupported HTTP method: PUT") {
		t.Errorf("Expected unsupported method error message, got: %v", err)
	}
}

func TestClient_Invoke_ServerError(t *testing.T) {
	errorMessage := "Internal Server Error"
	errorStatus := http.StatusInternalServerError
	functionName := "testErrorFunc"

	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(errorStatus)
		fmt.Fprint(w, errorMessage)
	})
	defer server.Close()

	_, err := client.Invoke(functionName, "GET", nil)
	if err == nil {
		t.Fatal("Expected an error from Invoke due to server error, but got nil")
	}

	expectedErrorSubstring := fmt.Sprintf("server responded with error: %d %s", errorStatus, http.StatusText(errorStatus))
	if !strings.Contains(err.Error(), expectedErrorSubstring) {
		t.Errorf("Error message '%v' does not contain expected substring '%s'", err, expectedErrorSubstring)
	}
	if !strings.Contains(err.Error(), errorMessage) {
		t.Errorf("Error message '%v' does not contain the server error body '%s'", err, errorMessage)
	}
}

func TestClient_Invoke_NetworkError(t *testing.T) {
	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
	})
	server.Close() // Close server immediately to simulate network error

	_, err := client.Invoke("testNetError", "POST", map[string]string{"data": "test"})
	if err == nil {
		t.Fatal("Expected a network error, but got nil")
	}
	if !strings.Contains(err.Error(), "failed to execute request") {
		t.Errorf("Expected network error message to contain 'failed to execute request', got: %v", err)
	}
}

type errorReader struct{}

func (er *errorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("simulated read error")
}

func (er *errorReader) Close() error { return nil }

func TestClient_Invoke_ReadResponseBodyError(t *testing.T) {
	functionName := "testReadError"

	// Custom RoundTripper for the session to inject errorReader
	mockSessionRT := MockRoundTripper(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       &errorReader{},
			Header:     make(http.Header),
		}, nil
	})

	serverURL, _ := url.Parse("http://localhost") // This base URL is for client.clientTransport
	client := &Client{
		clientTransport: transport{
			header:  http.Header{},
			baseUrl: *serverURL, // transport.baseUrl is url.URL
		},
		session: http.Client{Transport: mockSessionRT}, // Client.session is http.Client
	}

	_, err := client.Invoke(functionName, "GET", nil)
	if err == nil {
		t.Fatal("Expected an error from reading response body, but got nil")
	}
	if !strings.Contains(err.Error(), "failed to read response body: simulated read error") {
		t.Errorf("Expected read body error message, got: %v", err)
	}
}

func TestClient_Invoke_GET_NilPayload(t *testing.T) {
	expectedResponseString := `{"message": "get success no params"}`
	functionName := "testGetNilPayload"

	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
			http.Error(w, "Wrong method", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/"+functionName {
			t.Errorf("Expected path /%s, got %s", functionName, r.URL.Path)
			http.Error(w, "Wrong path", http.StatusBadRequest)
			return
		}
		if r.URL.RawQuery != "" {
			t.Errorf("Expected no query params for nil payload, got %s", r.URL.RawQuery)
			http.Error(w, "Unexpected query params", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, expectedResponseString)
	})
	defer server.Close()

	responseBytes, err := client.Invoke(functionName, "GET", nil)
	if err != nil {
		t.Fatalf("Invoke failed for GET with nil payload: %v", err)
	}
	if string(responseBytes) != expectedResponseString {
		t.Errorf("Expected response '%s', got '%s'", expectedResponseString, string(responseBytes))
	}
}

func TestClient_Invoke_POST_NilPayload(t *testing.T) {
	expectedResponseString := `{"message": "post success with null"}`
	functionName := "testPostNilPayload"

	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
			http.Error(w, "Wrong method", http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/"+functionName {
			t.Errorf("Expected path /%s, got %s", functionName, r.URL.Path)
			http.Error(w, "Wrong path", http.StatusBadRequest)
			return
		}
		bodyBytes, ioErr := io.ReadAll(r.Body)
		if ioErr != nil {
			t.Errorf("Failed to read request body: %v", ioErr)
			http.Error(w, "Cannot read body", http.StatusBadRequest)
			return
		}
		if string(bodyBytes) != "null" {
			t.Errorf("Expected body 'null' for nil payload, got '%s'", string(bodyBytes))
			http.Error(w, "Wrong body", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, expectedResponseString)
	})
	defer server.Close()

	responseBytes, err := client.Invoke(functionName, "POST", nil)
	if err != nil {
		t.Fatalf("Invoke failed for POST with nil payload: %v", err)
	}
	if string(responseBytes) != expectedResponseString {
		t.Errorf("Expected response '%s', got '%s'", expectedResponseString, string(responseBytes))
	}
}

func TestClient_Invoke_GET_EmptyMapPayload(t *testing.T) {
	expectedResponseString := `{"message": "get success empty params"}`
	functionName := "testGetEmptyMap"
	params := map[string]string{} // Empty map

	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
			return
		}
		if r.URL.RawQuery != "" {
			t.Errorf("Expected no query params for empty map payload, got %s", r.URL.RawQuery)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, expectedResponseString)
	})
	defer server.Close()

	responseBytes, err := client.Invoke(functionName, "GET", params)
	if err != nil {
		t.Fatalf("Invoke failed for GET with empty map payload: %v", err)
	}
	if string(responseBytes) != expectedResponseString {
		t.Errorf("Expected response '%s', got '%s'", expectedResponseString, string(responseBytes))
	}
}

func TestClient_Invoke_ResponseParsing_JSON(t *testing.T) {
	expectedData := map[string]interface{}{
		"id":      float64(123),
		"name":    "Test Item",
		"enabled": true,
	}
	serverResponseBytes, _ := json.Marshal(expectedData)
	// serverResponse string is not directly used for comparison, server just sends bytes
	functionName := "testJsonResponseFunc"

	client, server := newTestClient(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(serverResponseBytes) // Server sends bytes
	})
	defer server.Close()

	responseBytes, err := client.Invoke(functionName, "GET", nil)
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	var parsedResponse map[string]interface{}
	err = json.Unmarshal(responseBytes, &parsedResponse) // Unmarshal directly from []byte
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON response: %v. Response string: %s", err, string(responseBytes))
	}

	// Verify the parsed data
	if parsedResponse["name"] != expectedData["name"] {
		t.Errorf("Expected name '%s', got '%s'", expectedData["name"], parsedResponse["name"])
	}
	if parsedResponse["id"].(float64) != expectedData["id"].(float64) {
		t.Errorf("Expected id '%v', got '%v'", expectedData["id"], parsedResponse["id"])
	}
	if parsedResponse["enabled"] != expectedData["enabled"] {
		t.Errorf("Expected enabled '%v', got '%v'", expectedData["enabled"], parsedResponse["enabled"])
	}
}
