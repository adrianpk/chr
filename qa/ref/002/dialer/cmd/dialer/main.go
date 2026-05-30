package main

import (
	"log"
	"net/http"

	"dialer/internal/dialer"
)

func main() {
	server := dialer.NewServer(dialer.NewStore(dialer.SeedLeads()))
	log.Fatal(http.ListenAndServe(":8080", server.Routes()))
}
