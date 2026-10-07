// Fortify backend — Go port of the Python FastAPI service.
//
// Same API contract on :8500 so the React dashboard works unchanged.
// Faster because: one shared HTTP connection pool, concurrent probing
// (paths / params / scan families), and a single static binary.
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"fortify-go/internal/api"
	"fortify-go/internal/db"
)

func main() {
	if err := db.Init(); err != nil {
		log.Fatalf("db init: %v", err)
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8500"
	}
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           api.NewRouter(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	log.Printf("fortify-go listening on :%s", port)
	stopSched := make(chan struct{})
	defer close(stopSched)
	go api.StartScheduler(stopSched)
	log.Fatal(srv.ListenAndServe())
}
