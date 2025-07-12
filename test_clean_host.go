package main

import (
	"fmt"
	"regexp"
	"strings"
)

func main() {
	hostPattern := regexp.MustCompile(`(https?://)?([a-zA-Z0-9\._\-]+)(:[0-9]{2,6})?/?`)

	hosts := []string{
		"http://backend-service-1:80",
		"http://backend-service-2:80",
		"backend-service-1:80",
		"backend-service-2:80",
	}

	for _, host := range hosts {
		fmt.Printf("Testing host: %s\n", host)

		matches := hostPattern.FindAllStringSubmatch(host, -1)
		fmt.Printf("Matches count: %d\n", len(matches))

		if len(matches) != 1 {
			fmt.Printf("ERROR: Expected 1 match, got %d\n", len(matches))
			continue
		}

		keys := matches[0][1:]
		fmt.Printf("Keys: %v\n", keys)

		if keys[0] == "" {
			keys[0] = "http://"
		}

		result := strings.Join(keys, "")
		fmt.Printf("Result: %s\n", result)
		fmt.Println("---")
	}
}
