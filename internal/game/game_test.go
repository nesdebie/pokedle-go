package game

import (
	"crypto/hmac"
	"crypto/sha256"
	"testing"
	"time"

	"pokedle/internal/dataset"
	"pokedle/internal/pokeapi"
)

// legacyDailyIndex is the algorithm of the original main.go, kept to make
// sure the refactor does not change the Pokémon of the day.
func legacyDailyIndex(n int, secret string, t time.Time) int {
	msg := []byte(t.UTC().Format("2006-01-02"))
	var sum []byte
	if secret != "" {
		m := hmac.New(sha256.New, []byte(secret))
		m.Write(msg)
		sum = m.Sum(nil)
	} else {
		h := sha256.Sum256(msg)
		sum = h[:]
	}
	var v uint64
	for i := 0; i < 8; i++ {
		v = (v << 8) | uint64(sum[i])
	}
	return int(v % uint64(n))
}

func TestDailyIndexMatchesLegacy(t *testing.T) {
	day := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 60; i++ {
		d := day.AddDate(0, 0, i)
		for _, secret := range []string{"", "s3cr3t"} {
			if got, want := DailyIndex(1082, secret, d), legacyDailyIndex(1082, secret, d); got != want {
				t.Fatalf("%s secret=%q: got %d want %d", DayKey(d), secret, got, want)
			}
		}
	}
	if DailyIndex(0, "", day) != 0 {
		t.Error("DailyIndex with n=0 should be 0")
	}
}

func pokemon(id, height, weight int, types ...string) *pokeapi.Pokemon {
	p := &pokeapi.Pokemon{ID: id, Height: height, Weight: weight}
	for i, t := range types {
		p.Types = append(p.Types, pokeapi.TypeEntry{Slot: i + 1, Type: pokeapi.NamedResource{Name: t}})
	}
	return p
}

func TestCompare(t *testing.T) {
	d := Data{
		Evolutions:  map[int]dataset.Evolution{1: {Position: 0}, 6: {Position: 2, IsFullyEvolved: 1}},
		Generations: map[int]int{1: 1, 6: 1},
	}
	bulbasaur := pokemon(1, 7, 69, "grass", "poison")
	charizard := pokemon(6, 17, 905, "fire", "flying")

	h := d.Compare(bulbasaur, charizard)
	if h.Type1 != "GRASS" || h.Type2 != "POISON" || h.Type1Match || h.Type2Match {
		t.Errorf("types: %+v", h)
	}
	if h.WeightHint != ">6.9kg" || h.HeightHint != ">70cm" {
		t.Errorf("measures: %q %q", h.WeightHint, h.HeightHint)
	}
	if h.Distance != 5 || h.GuessPosition != 0 || h.TargetPosition != 2 || h.TargetFullyEvolved != 1 {
		t.Errorf("evolution/distance: %+v", h)
	}

	same := d.Compare(charizard, charizard)
	if same.WeightHint != "90.5kg" || !same.Type1Match || !same.Type2Match {
		t.Errorf("same pokemon: %+v", same)
	}

	mono := pokemon(4, 6, 85, "fire")
	if h := d.Compare(mono, charizard); h.Type2 != pokeapi.NoType || !h.Type1Match {
		t.Errorf("mono type: %+v", h)
	}
}
