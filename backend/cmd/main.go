package main

import (
	"bufio"
	"gohole/internal/controller/dns"
	"gohole/internal/controller/http"
	"gohole/internal/registry"
	"log"
	"os"
	"sync"
)

const (
	upstream    = "1.1.1.1:53"
	dnsAddress  = ":53"
	httpAddress = ":8080"
)

func mustParseBlockList(fileName string) []string {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var lines []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return lines
}

func main() {
	domains := mustParseBlockList("block.txt")

	reg := registry.NewRegistry(domains)

	wg := sync.WaitGroup{}

	go dns.Start(&wg, reg, dnsAddress, upstream)
	wg.Add(1)
	go http.Start(&wg, reg, httpAddress)
	wg.Add(1)

	wg.Wait()
}
