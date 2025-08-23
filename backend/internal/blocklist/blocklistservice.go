package blocklist

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

func ReadFromFile(fileName string) ([]string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("blocklist: opening blocklist file: %w", err)
	}
	defer file.Close()

	var lines []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("blocklist: reading blocklist file: %w", err)
	}

	var domains []string
	dones := 0

	// TODO parallelize
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		log.Printf("INFO Loading blocklist entry: %s\n", line)

		urls, err := download(line)
		if err != nil {
			log.Printf("ERROR error downloading blocklist %s: %v\n", line, err)
		} else {
			dones++
		}

		domains = append(domains, urls...)
	}

	log.Printf("INFO Loaded %d out of %d blocklists (%d domains)\n", dones, len(lines), len(domains))

	return domains, nil
}

func download(url string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("blocklist: downloading blocklist: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("blocklist: downloading blocklist: status code %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("blocklist: reading blocklist: %w", err)
	}

	split := strings.Split(string(body), "\n")
	var result []string

	for _, line := range split {
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, " ")
		for _, part := range parts {
			// Match ip addreses with regex
			r := regexp.MustCompile(`[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+`)
			if r.MatchString(part) {
				continue
			}

			result = append(result, part)
		}
	}

	return result, nil
}
