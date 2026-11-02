// Command evolutions generates data/pokemon_evolution_data.csv:
// for each species, its position in its evolution line (0 = basic)
// and whether it is fully evolved.
//
//	go run ./scripts/evolutions
package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"pokedle/internal/dataset"
	"pokedle/internal/pokeapi"
)

// Evolution chain IDs are not contiguous: stop after this many misses in a row.
const maxConsecutiveErrors = 10

// PokéAPI lists these species as able to evolve, but only through a
// regional form (Galarian Farfetch'd -> Sirfetch'd, Galarian Mr. Mime ->
// Mr. Rime). The standard form is therefore fully evolved.
var forceFullyEvolved = map[int]bool{
	83:  true, // Farfetch'd
	122: true, // Mr. Mime
}

type entry struct {
	id, position int
	fullyEvolved bool
}

// walk flattens an evolution tree depth-first.
func walk(node pokeapi.EvolutionNode, position int, out *[]entry) {
	id := pokeapi.IDFromURL(node.Species.URL)
	*out = append(*out, entry{
		id:           id,
		position:     position,
		fullyEvolved: len(node.EvolvesTo) == 0 || forceFullyEvolved[id],
	})
	for _, next := range node.EvolvesTo {
		walk(next, position+1, out)
	}
}

func main() {
	api := pokeapi.NewClient()

	var rows [][]string
	visited := make(map[int]bool)

	for chainID, misses := 1, 0; misses < maxConsecutiveErrors; chainID++ {
		var chain pokeapi.EvolutionChain
		if err := api.Get(fmt.Sprintf("evolution-chain/%d", chainID), &chain); err != nil {
			misses++
			continue
		}
		misses = 0

		var entries []entry
		walk(chain.Chain, 0, &entries)
		for _, e := range entries {
			if visited[e.id] {
				continue
			}
			visited[e.id] = true
			rows = append(rows, []string{
				strconv.Itoa(e.id),
				strconv.Itoa(e.position),
				boolToBit(e.fullyEvolved),
			})
			fmt.Printf("[ADD] #%d\n", e.id)
		}
		time.Sleep(100 * time.Millisecond)
	}

	header := []string{"id", "position", "is_fully_evolved"}
	if err := dataset.WriteCSV(dataset.EvolutionsFile, header, rows); err != nil {
		log.Fatal(err)
	}
	fmt.Println(dataset.EvolutionsFile, "generated successfully!")
}

func boolToBit(b bool) string {
	if b {
		return "1"
	}
	return "0"
}
