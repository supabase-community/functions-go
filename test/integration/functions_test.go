package functions_test

import (
	"fmt"
	"testing"

	"github.com/supabase-community/functions-go"
)

const (
	rawUrl = "https://your-supabase-url.co/functions/v1"
	token  = "supbase-service-key"
)

// TestHello tests the normal function
func TestHello(t *testing.T) {
	client := functions.NewClient(rawUrl, token, nil)
	type Body struct {
		Name string `json:"name"`
	}
	b := Body{Name: "world"}
	resp, err := client.Invoke("hello", b)
	if err != nil {
		t.Fatalf("Invoke failed: %s", err)
	}
	fmt.Println(resp)
}

// TestErrorHandling tests the error handling of the functions client
func TestErrorHandling(t *testing.T) {
	client := functions.NewClient(rawUrl, token, map[string]string{"custom-header": "custom-header"})
	type Body struct {
		Name string `json:"name"`
	}
	b := Body{Name: "error"}
	resp, err := client.Invoke("hello", b)
	if err != nil {
		t.Fatalf("Invoke failed: %s", err)
	}
	fmt.Println(resp)
}
