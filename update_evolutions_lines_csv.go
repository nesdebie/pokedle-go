package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"
)

type Species struct {
	URL  string `json:"url"`
}

type EvolutionDetail struct {
}

type EvolutionNode struct {
	Species   Species         `json:"species"`
	EvolvesTo []EvolutionNode `json:"evolves_to"`
}

type EvolutionChain struct {
	ID    int `json:"id"`
	Chain EvolutionNode `json:"chain"`
}

type PokemonResponse struct {
	Count int `json:"count"`
}

func getMaxId() int {
	resp, err := http.Get("https://pokeapi.co/api/v2/evolution-chain")
	if err != nil {
		fmt.Println("Error while getting count:", err)
		return 0
	}
	defer resp.Body.Close()

	var apiResp PokemonResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		fmt.Println("Error during decode:", err)
		return 0
	}

	low := 1
	high := apiResp.Count
	maxValid := 0

	for low <= high {
		mid := (low + high) / 2
		url := fmt.Sprintf("https://pokeapi.co/api/v2/evolution-chain/%d", mid)
		resp, err := http.Get(url)
		if err != nil {
			fmt.Printf("Error for ID %d: %v\n", mid, err)
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

func extractIDFromURL(url string) int {
	re := regexp.MustCompile(`/pokemon-species/(\d+)/`)
	matches := re.FindStringSubmatch(url)
	if len(matches) >= 2 {
		id, _ := strconv.Atoi(matches[1])
		return id
	}
	return 0
}

func traverseEvolutionChain(node EvolutionNode, evoChainID int, position int, output *[][]string) {
	pokemonID := extractIDFromURL(node.Species.URL)
	fullyEvolved := 0
	// 122 is hardcoded fix for mr mime error in PokeApi
	if len(node.EvolvesTo) == 0 || pokemonID == 122 {
		fullyEvolved = 1
	}

	row := []string{
		strconv.Itoa(pokemonID),
		strconv.Itoa(position),
		strconv.Itoa(fullyEvolved),
	}

	*output = append(*output, row)

	for _, next := range node.EvolvesTo {
		traverseEvolutionChain(next, evoChainID, position+1, output)
	}
}

func main() {
	const maxConsecutiveErrors = 10
	consecutiveErrors := 0
	currentID := 1

	client := &http.Client{Timeout: 5 * time.Second}
	output := [][]string{
		{"id", "position", "is_fully_evolved"},
	}

	visited := make(map[int]bool)

	for consecutiveErrors < maxConsecutiveErrors {
		url := fmt.Sprintf("https://pokeapi.co/api/v2/evolution-chain/%d", currentID)
		resp, err := client.Get(url)
		if err != nil {
			fmt.Printf("Error in evolution-chain %d: %v\n", currentID, err)
			consecutiveErrors++
			currentID++
			continue
		}

		if resp.StatusCode != 200 {
			resp.Body.Close()
			consecutiveErrors++
			currentID++
			continue
		}

		var chain EvolutionChain
		err = json.NewDecoder(resp.Body).Decode(&chain)
		resp.Body.Close()
		if err != nil {
			fmt.Printf("Error decoding JSON for evolution-chain %d: %v\n", currentID, err)
			consecutiveErrors++
			currentID++
			continue
		}

		consecutiveErrors = 0

		var tmpOutput [][]string
		traverseEvolutionChain(chain.Chain, chain.ID, 0, &tmpOutput)
		for _, row := range tmpOutput {
			id, _ := strconv.Atoi(row[0])
			if visited[id] {
				continue
			}
			visited[id] = true
			output = append(output, row)
			fmt.Println("[ADD]#", id)
		}

		currentID++
		time.Sleep(100 * time.Millisecond)
	}
	err := os.MkdirAll("data", os.ModePerm)
	if err != nil {
		panic(err)
	}

	file, err := os.Create("data/pokemon_evolution_data.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.WriteAll(output); err != nil {
		panic(err)
	}

	fmt.Println("data/pokemon_evolution_data.csv generated successfully !")

}