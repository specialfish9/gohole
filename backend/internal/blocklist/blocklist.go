package blocklist

import (
	"bufio"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"
)

// LoadRemote reads a file containing URLs of blocklists, downloads them, and
// returns a list of domains.
func LoadRemote(fileName string) ([]string, error) {
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

		slog.Info("Loading blocklist entry", "entry", line)

		blockList, err := download(line)
		if err != nil {
			slog.Error("Downloading blocklist", "blocklist", line, "error", err)
		} else {
			dones++
		}

		urls := parseBlockList(blockList)

		domains = append(domains, urls...)
	}

	slog.Info(fmt.Sprintf("Loaded %d out of %d blocklists (%d domains)\n", dones, len(lines), len(domains)))

	return domains, nil
}

func LoadLocalFile(fileName string) ([]string, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("blocklist: opening local blocklist file: %w", err)
	}

	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("blocklist: reading local blocklist file: %w", err)
	}

	domains := parseBlockList(string(content))
	return domains, nil
}

func download(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("blocklist: downloading blocklist: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("blocklist: downloading blocklist: status code %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("blocklist: reading blocklist: %w", err)
	}

	return string(body), nil
}

func parseBlockList(data string) []string {
	lines := strings.Split(data, "\n")
	var domains []string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip comments and empty lines
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

			domains = append(domains, part)
		}
	}

	return domains
}
