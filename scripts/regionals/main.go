// Command regionals generates data/pokemon_forms.csv: the regional forms
// (Alola, Galar, Hisui, Paldea) with their names, generation and evolution data.
//
// It reads data/pokemon_evolution_data.csv, so run `make evolutions` first.
//
//	go run ./scripts/regionals
package main

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	"pokedle/internal/dataset"
	"pokedle/internal/pokeapi"
)

// Alternate forms all have an ID >= firstFormID in PokéAPI.
const firstFormID = 10001

var regionalForm = regexp.MustCompile(`(galar|hisui|alola|paldea)`)

// generationByRegion maps the region found in a form name to the
// generation that introduced it.
var generationByRegion = []struct{ region, gen string }{
	{"alola", "7"},
	{"galar", "8"},
	{"hisui", "9"}, // Legends: Arceus, counted as gen 9 for the game
	{"paldea", "9"},
}

// Forms matching the regex that must not appear in the game.
var excludedPokemonIDs = map[int]bool{
	10093: true, // Totem Alolan Raticate
	10099: true, // Alolan Cap Pikachu
	10178: true, // Zen Mode Galarian Darmanitan
}

const galarianDarmanitanID = 10177

// evolutionOverrides fixes Galarian forms whose evolution line differs from
// their standard species. Farfetch'd is matched on its French name to avoid
// PokéAPI's typographic apostrophe in "Farfetch’d".
var evolutionOverrides = []struct {
	lang, name             string
	position, fullyEvolved string
}{
	{"en", "linoone", "1", "0"},    // Galarian Linoone -> Obstagoon
	{"fr", "canarticho", "0", "0"}, // Galarian Farfetch'd -> Sirfetch'd
}

func main() {
	api := pokeapi.NewClient()

	evolutions, err := dataset.LoadEvolutions(dataset.EvolutionsFile)
	if err != nil {
		log.Fatalf("%v (run `make evolutions` first)", err)
	}

	count, err := api.Count("pokemon-form")
	if err != nil {
		log.Fatal(err)
	}
	maxID := api.FindMaxID("pokemon-form", firstFormID, count+firstFormID-1)
	fmt.Println("Max form ID:", maxID)

	var rows [][]string
	for formID := firstFormID; formID <= maxID; formID++ {
		var form pokeapi.PokemonForm
		if err := api.Get(fmt.Sprintf("pokemon-form/%d", formID), &form); err != nil {
			continue
		}
		if !regionalForm.MatchString(form.Name) {
			continue
		}

		pokemonID := pokeapi.IDFromURL(form.Pokemon.URL)
		if excludedPokemonIDs[pokemonID] {
			continue
		}

		names := localizedNames(form, pokemonID)

		var pokemon pokeapi.Pokemon
		if err := api.Get(fmt.Sprintf("pokemon/%d", pokemonID), &pokemon); err != nil {
			continue
		}
		speciesID := pokeapi.IDFromURL(pokemon.Species.URL)
		position, fullyEvolved := evolutionOf(form.Name, names, evolutions, speciesID)

		row := append([]string{strconv.Itoa(pokemonID)}, names...)
		row = append(row, generationOf(form.Name), position, fullyEvolved)
		rows = append(rows, row)
		fmt.Println("Added Pokémon:", names[0])
	}

	header := append([]string{"id"}, dataset.Languages...)
	header = append(header, "gen", "position", "is_fully_evolved")
	if err := dataset.WriteCSV(dataset.FormsFile, header, rows); err != nil {
		log.Fatal(err)
	}
	fmt.Println("CSV created:", dataset.FormsFile)
}

// localizedNames returns the form name in each supported language
// (in dataset.Languages order), falling back to English.
func localizedNames(form pokeapi.PokemonForm, pokemonID int) []string {
	en := pokeapi.NameIn(form.Names, "en", "")
	names := make([]string, len(dataset.Languages))
	for i, lang := range dataset.Languages {
		names[i] = pokeapi.NameIn(form.Names, lang, en)

		// Galarian Darmanitan: drop the "Standard"/"Normal" mode suffix.
		if pokemonID == galarianDarmanitanID {
			n := strings.ReplaceAll(names[i], "Standard", "")
			n = strings.ReplaceAll(n, "Normal", "")
			names[i] = strings.TrimSpace(n)
		}
	}
	return names
}

func generationOf(formName string) string {
	for _, g := range generationByRegion {
		if strings.Contains(formName, g.region) {
			return g.gen
		}
	}
	return ""
}

// evolutionOf returns the (position, is_fully_evolved) cells of a form:
// an override if one applies, else the data of its standard species
// (empty cells if unknown).
func evolutionOf(formName string, names []string, evolutions map[int]dataset.Evolution, speciesID int) (string, string) {
	if strings.Contains(formName, "galar") {
		for _, o := range evolutionOverrides {
			if strings.Contains(strings.ToLower(nameIn(names, o.lang)), o.name) {
				return o.position, o.fullyEvolved
			}
		}
	}
	if evo, ok := evolutions[speciesID]; ok {
		return strconv.Itoa(evo.Position), strconv.Itoa(evo.IsFullyEvolved)
	}
	return "", ""
}

// nameIn picks a name from a slice ordered like dataset.Languages.
func nameIn(names []string, lang string) string {
	for i, l := range dataset.Languages {
		if l == lang {
			return names[i]
		}
	}
	return ""
}
