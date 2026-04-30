package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	endpointRegex = regexp.MustCompile(`https?://[^\s"'<>]+|/[a-zA-Z0-9_\-/.]+`)
	paramRegex    = regexp.MustCompile(`[?&]([a-zA-Z0-9_]+)=`)
	secretRegex   = regexp.MustCompile(`(?i)(api[_-]?key|token|Bearer)[^"' ]+`)
	jsRegex       = regexp.MustCompile(`\.js(\?|$)`)
	jsFinder      = regexp.MustCompile(`https?://[^\s"'<>]+\.js`)
)

// HTTP client with timeout
var client = &http.Client{
	Timeout: 10 * time.Second,
}

// Fetch content safely
func fetch(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "JSRecon/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Only process reasonable responses
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("bad status")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Skip extremely small responses (not useful JS)
	if len(body) < 50 {
		return "", fmt.Errorf("too small")
	}

	return string(body), nil
}

// Process JS content
func processJS(content string, endpoints, params, secrets map[string]bool, mu *sync.Mutex) {
	for _, match := range endpointRegex.FindAllString(content, -1) {
		mu.Lock()
		endpoints[match] = true
		mu.Unlock()
	}

	for _, match := range paramRegex.FindAllStringSubmatch(content, -1) {
		mu.Lock()
		params[match[1]] = true
		mu.Unlock()
	}

	for _, match := range secretRegex.FindAllString(content, -1) {
		mu.Lock()
		secrets[match] = true
		mu.Unlock()
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: jsrecon input.txt")
		return
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	endpoints := make(map[string]bool)
	params := make(map[string]bool)
	secrets := make(map[string]bool)

	var wg sync.WaitGroup
	var mu sync.Mutex

	// Concurrency limiter (important for bug bounty targets)
	sem := make(chan struct{}, 20)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			continue
		}

		wg.Add(1)
		go func(input string) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			// Case 1: Direct JS file
			if jsRegex.MatchString(input) {
				content, err := fetch(input)
				if err == nil {
					processJS(content, endpoints, params, secrets, &mu)
				}
				return
			}

			// Case 2: Normal URL (page)
			if strings.HasPrefix(input, "http") {
				body, err := fetch(input)
				if err != nil {
					return
				}

				jsLinks := jsFinder.FindAllString(body, -1)

				for _, js := range jsLinks {
					jsContent, err := fetch(js)
					if err == nil {
						processJS(jsContent, endpoints, params, secrets, &mu)
					}
				}
			}

			// Case 3: Domain (no scheme)
			if !strings.HasPrefix(input, "http") && strings.Contains(input, ".") {
				url := "http://" + input

				body, err := fetch(url)
				if err != nil {
					return
				}

				jsLinks := jsFinder.FindAllString(body, -1)

				for _, js := range jsLinks {
					jsContent, err := fetch(js)
					if err == nil {
						processJS(jsContent, endpoints, params, secrets, &mu)
					}
				}
			}

		}(line)
	}

	wg.Wait()

	writeToFile("endpoints.txt", endpoints)
	writeToFile("params.txt", params)
	writeToFile("secrets.txt", secrets)

	fmt.Println("[+] Done.")
}

// Write results to file
func writeToFile(filename string, data map[string]bool) {
	file, err := os.Create(filename)
	if err != nil {
		return
	}
	defer file.Close()

	for key := range data {
		file.WriteString(key + "\n")
	}
}
