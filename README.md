# `functions-go`

Golang client library to interact with Supabase Functions.

## Quick start
Install
```shell
go get github.com/supabase-community/functions-go
```

Usage

```go
package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/supabase-community/functions-go"
)

func main() {
	client := functions.NewClient("https://abc.functions.supabase.co", "<service-token>", nil)

	newData, err := json.Marshal(post)
	if err != nil {
		log.Fatal(err)
	}

	response := client.invoke("functionName", &FunctionInvokeOptions{
		Body: bytes.NewBuffer(newData),
		ResponseType: "json",
	})
	
	fmt.Println(response)
}
```

## License

This repo is licensed under MIT.

## Credits

- https://github.com/supabase/functions-js
