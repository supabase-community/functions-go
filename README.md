# Go Supabase Functions Client

A Go client library for interacting with Supabase Edge Functions. This library provides a convenient way to invoke your serverless functions, supporting both POST and GET requests, JSON payloads, and query parameters.

## Features

*   Initialize client with Supabase function URL and service token.
*   Invoke functions using POST with JSON payloads.
*   Invoke functions using GET with optional query parameters.
*   Handles JSON request and response bodies.
*   Returns raw `[]byte` for response body, allowing flexible unmarshaling.

## Installation

To use this library in your Go project, you can get it using:

```bash
go get github.com/supabase-community/functions-go
```

## Usage

### 1. Initialize the Client

First, create a new client instance using your Supabase Function's base URL and your service (or anon) key. Refer to your `client.go` for the exact `NewClient` signature and usage.

```go
import (
	"log"
	"fmt"
	"github.com/supabase-community/functions-go"
)

func main() {
	// Replace with your actual Supabase Functions URL and token
	supabaseURL := "https://<your-project-ref>.functions.supabase.co"
	supabaseToken := "<your-supabase-service-role-key-or-anon-key>"

	// Example:
    client := functions.NewClient(supabaseURL, supabaseToken, nil)
	// We'll assume 'client' is an instance of *functions.Client for the examples below.
	// Please adapt this initialization according to your NewClient function in client.go
}
```
*(Note: The examples below assume `client` is a properly initialized `*functions.Client` instance based on your `client.go`.)*

### 2. Invoking Functions

The `Invoke` method is used to call your Supabase Edge Functions.

`Invoke(functionName string, method string, payload interface{}) ([]byte, error)`

*   `functionName`: The name of the Supabase function to call (e.g., "hello-world").
*   `method`: The HTTP method, either "POST" or "GET" (case-insensitive).
*   `payload`:
    *   For "POST" requests: The data to be marshaled into a JSON body (e.g., a `map[string]interface{}` or a struct).
    *   For "GET" requests:
        *   If `nil`, no query parameters are sent.
        *   If a `map[string]string` or `map[string]interface{}`, these are converted to URL query parameters.
*   Returns:
    *   `[]byte`: The raw response body from the function.
    *   `error`: An error if the invocation failed.

#### Example: POST Request

```go
package main

import (
	"encoding/json"
	"log"
	"fmt"
    "github.com/supabase-community/functions-go"
)

// Assume 'client' is an initialized *functions.Client from your NewClient function
// var client *functions.Client 

type PostPayload struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type PostResponse struct {
	Message string `json:"message"`
	User    string `json:"user"`
}

func main() {
	// Initialize client here based on your client.go ...
	supabaseURL := "https://<your-project-ref>.functions.supabase.co"
	supabaseToken := "<your-supabase-service-role-key-or-anon-key>"
	client = functions.NewClient(supabaseURL, supabaseToken, nil) 
	if client == nil || client.clientError != nil { // Adjust error checking as per NewClient
	    log.Fatalf("Failed to create client.")
	}

	functionName := "user-profile"
	payloadData := PostPayload{Name: "Jane Doe", Age: 30}

	// Ensure 'client' is initialized before calling Invoke
	responseBody, err := client.Invoke(functionName, "POST", payloadData)
	if err != nil {
		log.Fatalf("Failed to invoke function '%s' with POST: %v", functionName, err)
	}

	log.Printf("Raw POST response for '%s': %s", functionName, string(responseBody))

	var parsedResponse PostResponse
	err = json.Unmarshal(responseBody, &parsedResponse)
	if err != nil {
	 	log.Fatalf("Failed to unmarshal POST response for '%s': %v", functionName, err)
	}

	fmt.Printf("Successfully called '%s'. Message: %s, User: %s", functionName, parsedResponse.Message, parsedResponse.User)
}
```
*(Note: The actual client initialization and Invoke call in the example above are commented out. You'll need to ensure `client` is properly initialized using your `NewClient` function before these snippets can run.)*

#### Example: GET Request without Query Parameters

```go
package main

import (
	"encoding/json"
	"log"
	"fmt"
	// "your/module/path/functions"
)

// Assume 'client' is an initialized *functions.Client
// var client *functions.Client

type GetItemResponse struct {
	ItemID   string `json:"itemId"`
	ItemName string `json:"itemName"`
}

func main() {
	// Initialize client here based on your client.go ...

	functionName := "get-item"
	
	responseBody, err := client.Invoke(functionName, "GET", nil) 
	if err != nil {
	 	log.Fatalf("Failed to invoke function '%s' with GET: %v", functionName, err)
	}

	log.Printf("Raw GET response for '%s': %s", functionName, string(responseBody))

	var items []GetItemResponse 
	err = json.Unmarshal(responseBody, &items)
	if err != nil {
		log.Fatalf("Failed to unmarshal GET response for '%s': %v", functionName, err)
	}

	fmt.Printf("Successfully called '%s'. Received %d items.", functionName, len(items))
	if len(items) > 0 {
		fmt.Printf("First item: ID=%s, Name=%s", items[0].ItemID, items[0].ItemName)
	}
}
```

#### Example: GET Request with Query Parameters

```go
package main

import (
	"encoding/json"
	"log"
	"fmt"
	// "your/module/path/functions"
)

// Assume 'client' is an initialized *functions.Client
// var client *functions.Client 

type Product struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

func main() {
	// Initialize client here based on your client.go ...

	functionName := "search-products"
	queryParams := map[string]interface{}{
		"category": "electronics",
		"limit":    10,
		"inStock":  true,
	}
	
	responseBody, err := client.Invoke(functionName, "GET", queryParams)
	if err != nil {
		log.Fatalf("Failed to invoke function '%s' with GET and query params: %v", functionName, err)
	}

	log.Printf("Raw GET response for '%s' with params: %s", functionName, string(responseBody))

	var products []Product
	err = json.Unmarshal(responseBody, &products)
	if err != nil {
		log.Fatalf("Failed to unmarshal GET response for '%s': %v", functionName, err)
	}

	fmt.Printf("Successfully called '%s'. Found %d products.", functionName, len(products))
	// Process products...
}
```

### 3. Handling the Response

The `Invoke` function returns the raw response body as `[]byte`. If your function returns JSON (which is common), you can unmarshal it using the standard `encoding/json` package:

```go
var result MyExpectedStruct // or map[string]interface{}
err = json.Unmarshal(responseBody, &result)
if err != nil {
// Handle error
}
// Use 'result'
```

If your function returns plain text or other non-JSON data, you can convert the `[]byte` to a string:
```go
responseText := string(responseBody)
// Use 'responseText'
```

### 4. Error Handling

The `Invoke` method returns an error if:
*   The HTTP method is unsupported (only "GET" and "POST" are allowed).
*   The payload for a POST request cannot be marshaled to JSON.
*   The payload for a GET request (if provided for query parameters) is not a `map[string]string` or `map[string]interface{}`.
*   The HTTP request cannot be created.
*   The request execution fails (e.g., network issues).
*   Reading the response body fails.
*   The server responds with an HTTP status code >= 400. The error message will include the status code and often the response body from the server for easier debugging.

Always check the returned error.

## Running Tests

To run the unit tests for this library (if `functions_test.go` is provided):
```bash
go test -v ./...
```

## Contributing

Details on how to contribute can be added here. For example:

1.  Fork the repository.
2.  Create your feature branch (`git checkout -b feature/AmazingFeature`).
3.  Commit your changes (`git commit -m 'Add some AmazingFeature'`).
4.  Push to the branch (`git push origin feature/AmazingFeature`).
5.  Open a Pull Request.

## License

This repository is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Credits

This library is inspired by the Supabase ecosystem and community projects.
For further inspiration and a JavaScript-based client, visit:

*   [functions-js](https://github.com/supabase/functions-js)
*   Official Supabase documentation: [https://supabase.com/docs/guides/functions](https://supabase.com/docs/guides/functions)
