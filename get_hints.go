package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

var targetLangs = []string{"en", "fr", "de", "it", "es"}

type FlavorTextEntry struct {
	FlavorText string `json:"flavor_text"`
	Language   struct {
		Name string `json:"name"`
	} `json:"language"`
}

type SpeciesResponse struct {
	FlavorTextEntries []FlavorTextEntry `json:"flavor_text_entries"`
}

type TypeSlot struct {
	Slot int `json:"slot"`
	Type struct {
		Name string `json:"name"`
	} `json:"type"`
}

type Cries struct {
	Latest string `json:"latest"`
}

type PokemonResponse struct {
	Id    int        `json:"id"`
	Types []TypeSlot `json:"types"`
	Cries Cries      `json:"cries"`
}

func cleanFlavorText(text string) string {
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\f", " ")
	return strings.TrimSpace(text)
}

func fetchDescriptions(id int) map[string]string {
	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon-species/%d/", id)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var data SpeciesResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		panic(err)
	}

	desc := make(map[string]string)
	for _, entry := range data.FlavorTextEntries {
		lang := entry.Language.Name
		if contains(targetLangs, lang) && desc[lang] == "" {
			desc[lang] = cleanFlavorText(entry.FlavorText)
		}
	}
	return desc
}

func fetchPokemonData(id int) (type1, type2 string, cryURL string) {
	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%d/", id)
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	var data PokemonResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		panic(err)
	}

	type1, type2 = "(none)", "(none)"
	for _, t := range data.Types {
		if t.Slot == 1 {
			type1 = t.Type.Name
		} else if t.Slot == 2 {
			type2 = t.Type.Name
		}
	}

	return type1, type2, data.Cries.Latest
}

func downloadCry(id int, url string) {
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		fmt.Printf("Error by downloading  #%d cry.\n", id)
		return
	}
	defer resp.Body.Close()

	filename := "hint.ogg"
	out, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		panic(err)
	}
}

func writeCSV(id int, type1, type2 string, descriptions map[string]string) {
	fileExists := false
	filename := "pokedex_data.csv"

	if _, err := os.Stat(filename); err == nil {
		fileExists = true
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if !fileExists {
		header := append([]string{"id", "type1", "type2"}, targetLangs...)
		writer.Write(header)
	}

	row := []string{
		fmt.Sprintf("%d", id),
		type1,
		type2,
	}
	for _, lang := range targetLangs {
		row = append(row, descriptions[lang])
	}

	writer.Write(row)
}

func contains(list []string, item string) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}

func main() {
	pokemonID := 1 // Bulbizarre temp

	type1, type2, cryURL := fetchPokemonData(pokemonID)
	descriptions := fetchDescriptions(pokemonID)

	downloadCry(pokemonID, cryURL)

	writeCSV(pokemonID, type1, type2, descriptions)
}
