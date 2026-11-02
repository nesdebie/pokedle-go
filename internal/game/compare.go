package game

import (
	"strconv"

	"pokedle/internal/dataset"
	"pokedle/internal/pokeapi"
)

// Hints is the comparison between a guess and the target, sent to the front.
type Hints struct {
	Type1                string `json:"type1"`
	Type2                string `json:"type2"`
	Type1Match           bool   `json:"type1Match"`
	Type2Match           bool   `json:"type2Match"`
	Type1MatchWrongPlace bool   `json:"type1MatchWrongPlace"`
	Type2MatchWrongPlace bool   `json:"type2MatchWrongPlace"`
	GuessedGen           int    `json:"guessedGen"`
	CorrectGen           int    `json:"correctGen"`
	WeightHint           string `json:"weightHint"`
	HeightHint           string `json:"heightHint"`
	GuessPosition        int    `json:"guessPosition"`
	TargetPosition       int    `json:"targetPosition"`
	GuessFullyEvolved    int    `json:"guessFullyEvolved"`
	TargetFullyEvolved   int    `json:"targetFullyEvolved"`
	Distance             int    `json:"distance"`
}

// Data holds the static per-Pokémon information loaded from the CSV files.
type Data struct {
	Evolutions  map[int]dataset.Evolution
	Generations map[int]int
}

// Compare builds the hints for guess against target.
func (d Data) Compare(guess, target *pokeapi.Pokemon) Hints {
	gT1, gT2 := guess.TypePair()
	tT1, tT2 := target.TypePair()
	gEvo, tEvo := d.Evolutions[guess.ID], d.Evolutions[target.ID]

	return Hints{
		Type1:                gT1,
		Type2:                gT2,
		Type1Match:           gT1 == tT1,
		Type2Match:           gT2 == tT2,
		Type1MatchWrongPlace: gT1 == tT2,
		Type2MatchWrongPlace: gT2 == tT1,
		GuessedGen:           d.Generations[guess.ID],
		CorrectGen:           d.Generations[target.ID],
		WeightHint:           compareHint(guess.Weight, target.Weight, formatWeight(guess.Weight)),
		HeightHint:           compareHint(guess.Height, target.Height, formatHeight(guess.Height)),
		GuessPosition:        gEvo.Position,
		TargetPosition:       tEvo.Position,
		GuessFullyEvolved:    gEvo.IsFullyEvolved,
		TargetFullyEvolved:   tEvo.IsFullyEvolved,
		Distance:             abs(target.ID - guess.ID),
	}
}

// compareHint prefixes label with ">" if the target is bigger,
// "<" if it is smaller, nothing if equal.
func compareHint(guess, target int, label string) string {
	switch {
	case guess < target:
		return ">" + label
	case guess > target:
		return "<" + label
	default:
		return label
	}
}

// formatWeight converts hectograms to "x.ykg".
func formatWeight(hg int) string {
	return strconv.FormatFloat(float64(hg)/10, 'f', 1, 64) + "kg"
}

// formatHeight converts decimetres to "xcm".
func formatHeight(dm int) string {
	return strconv.Itoa(dm*10) + "cm"
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
