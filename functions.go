package functions

import (
	jsonParser "encoding/json"
	"io"
)

/*
Invokes a function

functionName: the name of the function to invoke

Usage:
	invoke("Hello-Function", &FunctionInvokeOptions{
		Body			io.Reader
		ResponseType	string
	})
*/
func (c *Client) invoke(functionName string, options FunctionInvokeOptions) FunctionResponse {
	var responseType string
	if len(options.ResponseType) > 0 {
		responseType = options.ResponseType
	} else {
		responseType = json
	}
	response, _ := c.session.Post(c.clientTransport.baseUrl.String()+"/"+functionName, responseType, options.Body)

	isRelayError := response.Header.Get("x-relay-error")
	if len(isRelayError) > 0 && isRelayError == "true" {
		return FunctionResponse{
			Error:  response,
			Status: response.StatusCode,
		}
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return FunctionResponse{
			Error: err,
		}
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(response.Body)

	var data interface{}
	if responseType == json {
		err = jsonParser.Unmarshal(body, &data)
		if err != nil {
			return FunctionResponse{
				Error: err,
			}
		}
	} else if responseType == arrayBuffer || responseType == blob {
		data = body
	} else if responseType == text {
		data = string(body)
	}

	return FunctionResponse{
		Data:   data,
		Status: 200,
	}
}
