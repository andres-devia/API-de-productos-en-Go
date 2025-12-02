package main

import (
	"flag"
	"log"
	"os"
	api "project/internal/server"
)

func main() {
	// Analizar los indicadores de la línea de comandos
	port := flag.String("port", "8080", "Server port")
	dbPath := flag.String("db", "items.db", "SQLite database file path")
	flag.Parse()

	// Crear y iniciar el servidor
	cfg := api.Config{
		Port:   *port,
		DBPath: *dbPath,
	}
	server, err := api.NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
		os.Exit(1)
	}

	// Iniciar el servidor (bloquea hasta la interrupción)
	if err := server.Start(); err != nil {
		log.Fatalf("Server error: %v", err)
		os.Exit(1)
	}
}
