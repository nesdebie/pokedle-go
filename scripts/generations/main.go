// Command generations generates data/pokemon_id_gen.csv:
// the generation (1-9) each Pokémon species was introduced in.
//
//	go run ./scripts/generations
package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"pokedle/internal/dataset"
	"pokedle/internal/pokeapi"
)

var romanToGen = map[string]int{
	"i": 1, "ii": 2, "iii": 3, "iv": 4, "v": 5,
	"vi": 6, "vii": 7, "viii": 8, "ix": 9,
}

// genNumber converts "generation-iv" to 4 (0 if unknown).
func genNumber(name string) int {
	return romanToGen[strings.TrimPrefix(name, "generation-")]
}

func main() {
	api := pokeapi.NewClient()

	count, err := api.Count("pokemon")
	if err != nil {
		log.Fatal(err)
	}
	maxID := api.FindMaxID("pokemon", 1, count)
	fmt.Println("Max valid Pokémon ID:", maxID)

	var rows [][]string
	for id := 1; id <= maxID; id++ {
		species, err := api.Species(id)
		if err != nil {
			fmt.Printf("Skip #%d: %v\n", id, err)
			continue
		}

		if gen := genNumber(species.Generation.Name); gen > 0 {
			rows = append(rows, []string{strconv.Itoa(id), strconv.Itoa(gen)})
			fmt.Printf("[ADD] #%d - %dG\n", id, gen)
		}
		time.Sleep(100 * time.Millisecond)
	}

	if err := dataset.WriteCSV(dataset.GenerationsFile, []string{"id", "gen"}, rows); err != nil {
		log.Fatal(err)
	}
	fmt.Println(dataset.GenerationsFile, "generated successfully!")
}
