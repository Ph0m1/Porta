package main

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

func main() {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	
	urls := []string{
		"http://backend-service-1:80/",
		"http://backend-service-2:80/",
		"http://172.21.0.3:80/",
		"http://172.21.0.2:80/",
	}
	
	for _, url := range urls {
		fmt.Printf("Testing: %s\n", url)
		resp, err := client.Get(url)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			fmt.Printf("Status: %d, Body: %s\n", resp.StatusCode, string(body))
		}
		fmt.Println("---")
	}
} 