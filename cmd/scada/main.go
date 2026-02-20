package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {

	// Port configuration try to read "port, if not exists, default to 8080"
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	//Router set up
	mux := http.NewServeMux()
	// Health endpoint
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	addr := ":" + port // server address
	log.Printf("scada-mini listening on %s", addr)
	// start the server
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
