package app

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

func SendGet(url string) (string, error) {
	return send(url, "GET")
}

func SendPost(url string) (string, error) {
	return send(url, "POST")
}

func send(url string, requestType string) (string, error) {
	// URL of the endpoint you want to send the command to
	//url := "http://example.com/api/command"
	if len(requestType) == 0 {
		requestType = "GET"
	}

	// JSON payload for the command
	jsonStr := []byte(`{"key": "value"}`) // Modify this according to your command's JSON structure

	// Create a new HTTP POST request with the JSON payload
	req, err := http.NewRequest(requestType, url, bytes.NewBuffer(jsonStr))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return "", err
	}

	// Set the request headers (optional)
	req.Header.Set("Content-Type", "application/json")

	// Create an HTTP client and send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return "", err
	}

	defer resp.Body.Close()

	// Print the response status code
	if false {
		fmt.Println("Response Status:", resp.Status)
	}

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response body: %v\n", err)
		return "", err
	}

	return string(responseBody), nil
}
