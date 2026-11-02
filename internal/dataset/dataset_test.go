package dataset

import (
	"path/filepath"
	"testing"
)

func dataFile(name string) string { return filepath.Join("..", "..", name) }

func TestNormalizeKey(t *testing.T) {
	cases := map[string]string{
		"  Évoli ":   "evoli",
		"FLABÉBÉ":    "flabebe",
		"Farfetch’d": "farfetch’d",
	}
	for in, want := range cases {
		if got := NormalizeKey(in); got != want {
			t.Errorf("NormalizeKey(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLoadNames(t *testing.T) {
	idx, err := LoadNames(dataFile(NamesFile))
	if err != nil {
		t.Fatal(err)
	}
	if idx.Len() < 1000 {
		t.Fatalf("expected at least 1000 Pokémon, got %d", idx.Len())
	}
	for _, name := range []string{"Bulbizarre", "bulbasaur", "BISASAM"} {
		if id, ok := idx.Lookup(name); !ok || id != 1 {
			t.Errorf("Lookup(%q) = %d, %v; want 1, true", name, id, ok)
		}
	}
	if got := idx.EnglishName(25); got != "Pikachu" {
		t.Errorf("EnglishName(25) = %q", got)
	}
	if idx.IDAt(0) != 1 || idx.IDAt(-1) != 0 || idx.IDAt(idx.Len()) != 0 {
		t.Error("IDAt bounds are wrong")
	}
}

func TestSuggest(t *testing.T) {
	idx, err := LoadNames(dataFile(NamesFile))
	if err != nil {
		t.Fatal(err)
	}
	if got := idx.Suggest("   "); len(got) != 0 {
		t.Errorf("empty query should return no group, got %v", got)
	}
	groups := idx.Suggest("pika")
	if len(groups) == 0 || groups[0].Lang != "en" || groups[0].Names[0] != "Pikachu" {
		t.Errorf("unexpected suggestions for 'pika': %v", groups)
	}
}

func TestLoadEvolutionsAndGenerations(t *testing.T) {
	evo, err := LoadEvolutions(dataFile(EvolutionsFile))
	if err != nil {
		t.Fatal(err)
	}
	if e := evo[3]; e.Position != 2 || e.IsFullyEvolved != 1 { // Venusaur
		t.Errorf("Venusaur evolution = %+v", e)
	}

	gen, err := LoadGenerations(dataFile(GenerationsFile))
	if err != nil {
		t.Fatal(err)
	}
	if gen[1] != 1 || gen[1025] != 9 {
		t.Errorf("gen[1]=%d gen[1025]=%d", gen[1], gen[1025])
	}
}
