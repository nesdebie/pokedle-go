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

var excludedPokemonIDs = map[int]bool{
	10093: true, //Totem Galarian Ratticate
	10099: true, //Alolan Cap Pikachu
	10178: true, //Zen Galarian Darmanitan
}

type Name struct {
	Name     string `json:"name"`
	Language struct {
		Name string `json:"name"`
	} `json:"language"`
}

type PokemonForm struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Pokemon struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"pokemon"`
	Names []Name `json:"names"`
}

type PokemonResponse struct {
	Count int `json:"count"`
}

type PokemonAPI struct {
	Species struct {
		URL string `json:"url"`
	} `json:"species"`
}

type EvolutionData struct {
	SpeciesID      string
	Position       string
	IsFullyEvolved string
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

func getNameInLanguage(names []Name, lang string, fallback string) string {
	for _, n := range names {
		if n.Language.Name == lang {
			return n.Name
		}
	}
	return fallback
}

func loadEvolutionData(filepath string) (map[int]EvolutionData, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	data := make(map[int]EvolutionData)
	for i, row := range records {
		if i == 0 {
			continue
		}
		speciesID, _ := strconv.Atoi(row[0])
		data[speciesID] = EvolutionData{
			SpeciesID:      row[0],
			Position:       row[1],
			IsFullyEvolved: row[2],
		}
	}
	return data, nil
}

func getSpeciesID(pokemonID int) (int, error) {
	var poke PokemonAPI
	url := fmt.Sprintf("%s/pokemon/%d", baseURL, pokemonID)
	if err := fetchJSON(url, &poke); err != nil {
		return 0, err
	}
	return getPokemonIDFromURL(poke.Species.URL), nil
}

func main() {
	maxID := getMaxFormID()
	fmt.Printf("Max form ID: %d\n", maxID)

	evoData, err := loadEvolutionData("data/pokemon_evolution_data.csv")
	if err != nil {
		panic(err)
	}

	file, err := os.Create("data/pokemon_forms.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	writer.Write([]string{
		"id",
		"en",
		"fr",
		"de",
		"es",
		"it",
		"gen",
		"position",
		"is_fully_evolved",
	})

	for formID := 10001; formID <= maxID; formID++ {
		url := fmt.Sprintf("%s/pokemon-form/%d", baseURL, formID)

		var pf PokemonForm
		if err := fetchJSON(url, &pf); err != nil {
			continue
		}

		if !formsToKeep.MatchString(pf.Name) {
			continue
		}

		pokemonID := getPokemonIDFromURL(pf.Pokemon.URL)

		if excludedPokemonIDs[pokemonID] {
			continue
		}

		nameEn := getNameInLanguage(pf.Names, "en", "")
		nameFr := getNameInLanguage(pf.Names, "fr", nameEn)
		nameDe := getNameInLanguage(pf.Names, "de", nameEn)
		nameEs := getNameInLanguage(pf.Names, "es", nameEn)
		nameIt := getNameInLanguage(pf.Names, "it", nameEn)

		// Fix for pokemonID Galarian Darmanitan because of its Zen Mode (remove Standard/Normal)
		if pokemonID == 10177 {
			for _, name := range []*string{&nameEn, &nameFr, &nameDe, &nameEs, &nameIt} {
				*name = strings.ReplaceAll(*name, "Standard", "")
				*name = strings.ReplaceAll(*name, "Normal", "")
				*name = strings.TrimSpace(*name)
			}
		}

		genID := getGenerationFromName(pf.Name)

		speciesID, err := getSpeciesID(pokemonID)
		if err != nil {
			continue
		}

		position := ""
		isFullyEvolved := ""

		// Special case: Galarian Linoone and Fartfetchd
		if strings.Contains(strings.ToLower(nameEn), "linoone") && strings.Contains(pf.Name, "galar") {
			position = "1"
			isFullyEvolved = "0"
		} else if strings.Contains(strings.ToLower(nameFr), "canarticho") && strings.Contains(pf.Name, "galar") {
			position = "0"
			isFullyEvolved = "0"
		} else if evo, ok := evoData[speciesID]; ok {
			position = evo.Position
			isFullyEvolved = evo.IsFullyEvolved
		}

		writer.Write([]string{
			strconv.Itoa(pokemonID),
			nameEn,
			nameFr,
			nameDe,
			nameEs,
			nameIt,
			genID,
			position,
			isFullyEvolved,
		})

		fmt.Println("Added Pokémon:", nameEn)
	}

	fmt.Println("CSV created : data/pokemon_forms.csv")
}
