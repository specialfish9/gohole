package main

import (
	"bufio"
	"context"
	"gohole/internal/controller/dns"
	"gohole/internal/controller/http"
	"gohole/internal/database"
	"gohole/internal/registry"
	"log"
	"os"
	"sync"
)

const (
	upstream    = "1.1.1.1:53"
	dnsAddress  = ":53"
	httpAddress = ":8080"
	dbAddress   = "localhost:9000"
	dbUser      = "gohole"
	dbPassword  = "password"
	dbName      = "default"
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

	log.Println("INFO Starting GoHole...")
	dbConn, err := database.Connect(dbAddress, dbName, dbUser, dbPassword, false)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("INFO Connected to DB")

	if err := database.Init(context.Background(), dbConn); err != nil {
		log.Fatal(err)
	}

	log.Printf("INFO created tables")

	reg := registry.NewRegistry(domains, dbConn)

	wg := sync.WaitGroup{}

	go dns.Start(&wg, reg, dnsAddress, upstream)
	wg.Add(1)
	go http.Start(&wg, reg, httpAddress)
	wg.Add(1)

	wg.Wait()
}
