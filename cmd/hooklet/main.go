package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"hooklet/internal/event"
	"hooklet/internal/replay"
	"hooklet/internal/server"
	"hooklet/internal/store"
)

func main() {
 	port := flag.Int("port", 8080, "Port to listen on")
	target := flag.String("target", "http://localhost:8000", "Default target server for webhook replays")
	dbPath := flag.String("db", "hooklet.db", "Path to SQLite database file")
	flag.Parse()

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