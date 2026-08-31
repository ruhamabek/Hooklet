package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"hooklet/internal/event"
	"hooklet/internal/replay"
	"hooklet/internal/server"
	"hooklet/internal/store"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func main() {
	defaultPort := getEnvInt("PORT", 8080)
	defaultTarget := getEnv("HOOKLET_TARGET", getEnv("TARGET", "http://localhost:8000"))
	defaultDB := getEnv("HOOKLET_DB", getEnv("DB_PATH", "hooklet.db"))

	port := flag.Int("port", defaultPort, "Port to listen on (env: PORT)")
	target := flag.String("target", defaultTarget, "Default target server for webhook replays (env: HOOKLET_TARGET)")
	dbPath := flag.String("db", defaultDB, "Path to SQLite database file (env: HOOKLET_DB)")
	flag.Parse()

	if dir := filepath.Dir(*dbPath); dir != "." && dir != "" {
		_ = os.MkdirAll(dir, 0755)
	}

	db, err := store.NewSqliteStore(*dbPath)
	if err != nil {
		log.Fatalf("[ERROR] Failed to open database at %s: %v", *dbPath, err)
	}
	defer db.Close()

 	broker := event.NewBroker()
	dispatcher := replay.NewDispatcher(db, http.DefaultClient)

 	srv := server.New(db, broker, dispatcher, *target)

 	fmt.Println()
	fmt.Println("  Hooklet (v0.1) - Webhook Capture & Replay Engine")
	fmt.Println("  ---------------------------------------------------")
	fmt.Printf("  • Dashboard:        http://localhost:%d/\n", *port)
	fmt.Printf("  • Ingress Listener: http://localhost:%d/wh/*\n", *port)
	fmt.Printf("  • Forward Target:   %s\n", *target)
	fmt.Printf("  • SQLite Database:  %s\n", *dbPath)
	fmt.Println("  ---------------------------------------------------")
	fmt.Printf("  Ready. Listening for incoming webhooks on :%d\n\n", *port)

 	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	addr := fmt.Sprintf(":%d", *port)
	httpServer := &http.Server{
		Addr:    addr,
		Handler: srv.Handler(),
	}

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-stop
	fmt.Println("\n  Shutting down Hooklet...")
	_ = httpServer.Close()
	fmt.Println("  Done. Bye!")
}