package main

import (
	"log"
	"net/http"
	"os"

	"task-api-go/internal/api"
	"task-api-go/internal/store"
)

func main() {
	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}
	log.Printf("task API listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, api.NewRouter(store.New())))
}
