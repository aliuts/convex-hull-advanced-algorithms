package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/ali/convex-hull-advanced-algorithms/internal/algorithms"
	"github.com/ali/convex-hull-advanced-algorithms/internal/api"
)

func main() {
	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}
	server := &http.Server{Addr: addr, Handler: api.NewServer(algorithms.All()), ReadHeaderTimeout: 5 * time.Second}
	log.Printf("convex hull comparison server listening on %s", addr)
	log.Fatal(server.ListenAndServe())
}
