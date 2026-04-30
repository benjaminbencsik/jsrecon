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
	endpointRegex = regexp.MustCompile(`https?://[^\s"']+|/[a-zA-Z0-9_\-/]+`)
	paramRegex    = regexp.MustCompile(`[?&]([a-zA-Z0-9_]+)=`)
	secretRegex   = regexp.MustCompile(`(?i)(api[_-]?key|token|Bearer)[^"' ]+`)
	jsRegex       = regexp.MustCompile(`\.js(\?|$)`)
)

var client = &http.Client{
	Timeout: 10 * time.Second,
}

func fetch(url string) (string, error) {
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	return string(body), err
}

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

	// limit concurrency (important)
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

			// JS file directly
			if jsRegex.MatchString(input) {
				content, err := fetch(input)
				if err == nil {
					processJS(content, endpoints, params, secrets, &mu)
				}
				return
			}

			// Normal URL → extract JS from page
			if strings.HasPrefix(input, "http") {
				body, err := fetch(input)
				if err != nil {
					return
				}

				jsLinks := regexp.MustCompile(`https?://[^\s"']+\.js`).FindAllString(body, -1)

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

func writeToFile(filename string, data map[string]bool) {
	file, _ := os.Create(filename)
	defer file.Close()

	for key := range data {
		file.WriteString(key + "\n")
	}
}
