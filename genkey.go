package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
)

func main() {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		panic(err)
	}

	secret := hex.EncodeToString(key)

	file, err := os.Create(".env")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("POKEDLE_SECRET=%s\n", secret))
	if err != nil {
		panic(err)
	}

	fmt.Println(".env file created with POKEDLE_SECRET")
}
