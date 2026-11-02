// Command pokedle starts the Pokédle web server.
//
//	./pokedle       # production mode
//	./pokedle dev   # cookies expire after 5 minutes
package main

import (
	"log"
	"net/http"
	"os"

	"pokedle/internal/server"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	srv, err := server.New(server.Config{
		StaticDir: "static",
		EnvFile:   ".env",
		DevMode:   len(os.Args) == 2 && os.Args[1] == "dev",
	})
	if err != nil {
		log.Fatal(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Pokedle running on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, srv.Routes()))
}
