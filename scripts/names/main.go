// Command names generates data/pokemon_names_multilang.csv:
// every Pokémon species with its name in each supported language.
// Species missing a name in one of the languages are skipped.
//
//	go run ./scripts/names
package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"pokedle/internal/dataset"
	"pokedle/internal/pokeapi"
)

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

		row := []string{strconv.Itoa(id)}
		complete := true
		for _, lang := range dataset.Languages {
			name := pokeapi.NameIn(species.Names, lang, "")
			if name == "" {
				complete = false
				break
			}
			row = append(row, name)
		}

		if complete {
			rows = append(rows, row)
			fmt.Printf("[ADD] #%d - %s\n", id, row[1])
		}
		time.Sleep(100 * time.Millisecond)
	}

	header := append([]string{"id"}, dataset.Languages...)
	if err := dataset.WriteCSV(dataset.NamesFile, header, rows); err != nil {
		log.Fatal(err)
	}
	fmt.Println(dataset.NamesFile, "generated successfully!")
}
