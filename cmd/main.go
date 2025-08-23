package main

import (
	"bufio"
	"gohole/internal/controller/dns"
	"gohole/internal/query"
	"log"
	"os"
)

const (
	upstream = "1.1.1.1:53"
	address  = ":53"
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

	dns.Start(query.Trie(domains), address, upstream)
}
