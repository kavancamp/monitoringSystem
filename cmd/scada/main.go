package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/kavancamp/monitoringSystem/internal/api"
	"github.com/kavancamp/monitoringSystem/internal/database"
	db "github.com/kavancamp/monitoringSystem/internal/database/db"
)

func main() {
	ctx := context.Background()

	pool, err := database.NewPool(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	q := db.New(pool)
	srv := api.NewServer(q)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port

	log.Printf("scada-mini listening on %s", addr)
	if err := http.ListenAndServe(addr, srv.Routes()); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
