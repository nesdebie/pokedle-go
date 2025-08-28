package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const baseURL = "https://pokeapi.co/api/v2"

var formsToKeep = regexp.MustCompile(`(galar|hisui|alola|paldea)`)

type PokemonForm struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Pokemon struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"pokemon"`
}

type PokemonResponse struct {
	Count int `json:"count"`
}

func fetchJSON(url string, target interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("non-200 response: %d for %s", resp.StatusCode, url)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}

func getPokemonIDFromURL(url string) int {
	parts := strings.Split(strings.Trim(url, "/"), "/")
	idStr := parts[len(parts)-1]
	id, _ := strconv.Atoi(idStr)
	return id
}

func getMaxFormID() int {
	resp, err := http.Get(baseURL + "/pokemon-form")
	if err != nil {
		fmt.Println("Error fetching Pokémon-form count:", err)
		return 0
	}
	defer resp.Body.Close()

	var apiResp PokemonResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		fmt.Println("Error decoding Pokémon-form count:", err)
		return 0
	}

	low := 10001
	high := apiResp.Count + 10000
	maxValid := 0

	for low <= high {
		mid := (low + high) / 2
		url := fmt.Sprintf("%s/pokemon-form/%d", baseURL, mid)
		resp, err := http.Get(url)

		if err != nil {
			fmt.Printf("Error checking ID %d: %v\n", mid, err)
			return maxValid
		}
		resp.Body.Close()

		if resp.StatusCode == 200 {
			maxValid = mid
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return maxValid
}

func getGenerationFromName(name string) string {
	switch {
	case strings.Contains(name, "alola"):
		return "7"
	case strings.Contains(name, "galar"):
		return "8"
	case strings.Contains(name, "hisui"):
		return "9"
	case strings.Contains(name, "paldea"):
		return "9"
	default:
		return ""
	}
}

func main() {
	maxID := getMaxFormID()
	fmt.Printf("Max form ID: %d\n", maxID)

	file, err := os.Create("data/pokemon_forms.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{"pokemon_id", "name", "generation_id"})

	for id := 10001; id <= maxID; id++ {
		url := fmt.Sprintf("%s/pokemon-form/%d", baseURL, id)

		var pf PokemonForm
		if err := fetchJSON(url, &pf); err != nil {
			continue
		}

		if !formsToKeep.MatchString(pf.Name) {
			continue
		}

		pokemonID := getPokemonIDFromURL(pf.Pokemon.URL)
		genID := getGenerationFromName(pf.Name)

		writer.Write([]string{
			strconv.Itoa(pokemonID),
			pf.Name,
			genID,
		})
	}

	fmt.Println("CSV created : data/pokemon_forms.csv")
}
