package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type PokemonResponse struct {
	Count int `json:"count"`
}

type PokemonSpecies struct {
	Generation struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"generation"`
}

func getMaxId() int {
	resp, err := http.Get("https://pokeapi.co/api/v2/pokemon")
	if err != nil {
		fmt.Println("Error fetching Pokémon count:", err)
		return 0
	}
	defer resp.Body.Close()

	var apiResp PokemonResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		fmt.Println("Error decoding Pokémon count:", err)
		return 0
	}

	low := 1
	high := apiResp.Count
	maxValid := 0

	for low <= high {
		mid := (low + high) / 2
		url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%d", mid)
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

func main() {
	maxID := getMaxId()
	fmt.Printf("Max valid Pokémon ID: %d\n", maxID)

	output := [][]string{
		{"id", "gen"},
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	for i := 1; i <= maxID; i++ {
		url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon-species/%d", i)
		apiResponse, err := client.Get(url)
		if err != nil {
			fmt.Printf("Error on ID %d: %v\n", i, err)
			continue
		}

		if apiResponse.StatusCode != 200 {
			apiResponse.Body.Close()
			continue
		}

		var species PokemonSpecies
		err = json.NewDecoder(apiResponse.Body).Decode(&species)
		apiResponse.Body.Close()
		if err != nil {
			fmt.Printf("Error decoding JSON on ID %d: %v\n", i, err)
			continue
		}

		genName := species.Generation.Name
		genNum := 0
		if genName != "" {
			parts := strings.Split(genName, "-")
			if len(parts) == 2 {
				switch parts[1] {
				case "i":
					genNum = 1
				case "ii":
					genNum = 2
				case "iii":
					genNum = 3
				case "iv":
					genNum = 4
				case "v":
					genNum = 5
				case "vi":
					genNum = 6
				case "vii":
					genNum = 7
				case "viii":
					genNum = 8
				case "ix":
					genNum = 9
				default:
					genNum = 0
				}
			}
		}

		if genNum > 0 {
			row := []string{
				strconv.Itoa(i),
				strconv.Itoa(genNum),
			}
			output = append(output, row)
			fmt.Println("[ADD]#", i, " - ", genNum, "G")
		}

		time.Sleep(100 * time.Millisecond)
	}

	err := os.MkdirAll("data", os.ModePerm)
	if err != nil {
		panic(err)
	}

	file, err := os.Create("data/pokemon_id_gen.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	err = writer.WriteAll(output)
	if err != nil {
		panic(err)
	}

	fmt.Println("data/pokemon_id_gen.csv generated successfully!")
}
