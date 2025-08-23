package main

import (
	"context"
	"gohole/internal/blocklist"
	"gohole/internal/controller/dns"
	"gohole/internal/controller/http"
	"gohole/internal/database"
	"gohole/internal/registry"
	"log"
	"sync"
)

const (
	upstream      = "1.1.1.1:53"
	dnsAddress    = ":53"
	httpAddress   = ":8080"
	dbAddress     = "localhost:9000"
	dbUser        = "gohole"
	dbPassword    = "password"
	dbName        = "default"
	blocklistFile = "block.txt"
)

func main() {
	log.Println("========")
	log.Println(" GOHOLE ")
	log.Println("========")

	domains, err := blocklist.ReadFromFile(blocklistFile)
	if err != nil {
		log.Fatal(err)
	}

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
