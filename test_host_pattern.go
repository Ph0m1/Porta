package main

import (
	"fmt"
	"regexp"
)

func main() {
	hostPattern := regexp.MustCompile(`(https?://)?([a-zA-Z0-9\._\-]+)(:[0-9]{2,6})?/?`)

	hosts := []string{
		"http://backend-service-1:80",
		"http://backend-service-2:80",
		"backend-service-1:80",
		"backend-service-2:80",
		"http://127.0.0.1:8081",
		"127.0.0.1:8081",
	}

	for _, host := range hosts {
		matches := hostPattern.FindAllStringSubmatch(host, -1)
		fmt.Printf("Host: %s\n", host)
		fmt.Printf("Matches: %d\n", len(matches))
		if len(matches) > 0 {
			fmt.Printf("Groups: %v\n", matches[0])
		}
		fmt.Println("---")
	}
}
