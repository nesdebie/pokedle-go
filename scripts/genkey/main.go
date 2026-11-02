// Command genkey writes a random POKEDLE_SECRET into .env.
// The secret makes the Pokémon of the day unpredictable.
//
//	go run ./scripts/genkey
package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"pokedle/internal/game"
)

const envFile = ".env"

func main() {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		log.Fatal(err)
	}

	line := fmt.Sprintf("%s=%s\n", game.SecretKey, hex.EncodeToString(key))
	if err := os.WriteFile(envFile, []byte(line), 0o600); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s file created with %s\n", envFile, game.SecretKey)
}
