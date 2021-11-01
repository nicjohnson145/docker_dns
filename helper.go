package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

func getJsonResponse(verb string, address string, headers map[string]string, body io.Reader) (string, error) {
	client := &http.Client{}
	req, err := http.NewRequest(verb, address, body)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	for key, val := range headers {
		req.Header.Set(key, val)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("executing request: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("bad status code: %v", resp.StatusCode)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	return string(respBody), nil
}

func stringInSlice(s string, slice []string) bool {
	for _, str := range slice {
		if str == s {
			return true
		}
	}

	return false
}

func getCurrentIp() (string, error) {
	return getJsonResponse("GET", "https://ifconfig.me", map[string]string{}, nil)
}
