package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"calendar/internal/calendar"
	"calendar/internal/httpserver"
)

func main() {
	var port string
	flag.StringVar(&port, "port", getenv("PORT", "8080"), "HTTP server port")
	flag.Parse()

	svc := calendar.NewService()
	srv := httpserver.New(svc)
	h := httpserver.LoggingMiddleware(srv.Router())

	addr := ":" + port
	log.Printf("Starting server on %s", addr)
	if err := http.ListenAndServe(addr, h); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
