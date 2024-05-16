# `functions-go`

Golang client library to interact with Supabase Functions.

## Quick start

### Installation

Install the package using:

```shell
go get github.com/supabase-community/functions-go
```

### Usage

The following example demonstrates how to create a client, marshal data into JSON, and make a request to execute a function on the server.

```go
package main

import (
 "fmt"
 "log"

 "github.com/supabase-community/functions-go"
)

func main() {
 client := functions.NewClient("https://abc.supabase.co/functions/v1", "<service-token>", nil)

 // Define your data struct
 type Post struct {
  Title   string `json:"title"`
  Content string `json:"content"`
 }
 post := Post{Title: "Hello, world!", Content: "This is a new post."}

 // Invoke the function with the post data
 response, err := client.Invoke("createPost", post)
 if err != nil {
  log.Fatal(err)
 }

 fmt.Println("Response from server:", response)
}
```

This code will marshal the `Post` struct into JSON, send it to the `createPost` function, and print the response.

## License

This repository is licensed under the MIT License.

## Credits

For further inspiration and a JavaScript-based client, visit:

- [functions-js](https://github.com/supabase/functions-js)
