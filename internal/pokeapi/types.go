package pokeapi

import (
	"fmt"
	"strings"
)

// NamedResource is the generic {name, url} reference used across PokéAPI.
type NamedResource struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// LocalizedName is a name in a given language.
type LocalizedName struct {
	Name     string        `json:"name"`
	Language NamedResource `json:"language"`
}

// NameIn returns the name in the requested language, or fallback if absent.
func NameIn(names []LocalizedName, lang, fallback string) string {
	for _, n := range names {
		if n.Language.Name == lang {
			return n.Name
		}
	}
	return fallback
}

// ---------- /pokemon/{id} ----------

type Pokemon struct {
	ID      int           `json:"id"`
	Name    string        `json:"name"`
	Height  int           `json:"height"` // decimetres
	Weight  int           `json:"weight"` // hectograms
	Types   []TypeEntry   `json:"types"`
	Sprites Sprites       `json:"sprites"`
	Cries   Cries         `json:"cries"`
	Species NamedResource `json:"species"`
}

type TypeEntry struct {
	Slot int           `json:"slot"`
	Type NamedResource `json:"type"`
}

type Sprites struct {
	FrontDefault string `json:"front_default"`
	Other        map[string]struct {
		FrontDefault string `json:"front_default"`
		FrontShiny   string `json:"front_shiny"`
	} `json:"other"`
}

type Cries struct {
	Latest string `json:"latest"`
}

// NoType is used when a Pokémon has no secondary type.
const NoType = "(none)"

// TypePair returns the upper-cased primary and secondary types.
func (p *Pokemon) TypePair() (string, string) {
	t1, t2 := NoType, NoType
	for _, te := range p.Types {
		switch te.Slot {
		case 1:
			t1 = strings.ToUpper(te.Type.Name)
		case 2:
			t2 = strings.ToUpper(te.Type.Name)
		}
	}
	return t1, t2
}

// Artwork returns the official artwork URL, falling back to the default sprite.
func (p *Pokemon) Artwork() string {
	if oa, ok := p.Sprites.Other["official-artwork"]; ok && oa.FrontDefault != "" {
		return oa.FrontDefault
	}
	return p.Sprites.FrontDefault
}

// ---------- /pokemon-species/{id} ----------

type PokemonSpecies struct {
	Names             []LocalizedName   `json:"names"`
	Generation        NamedResource     `json:"generation"`
	FlavorTextEntries []FlavorTextEntry `json:"flavor_text_entries"`
}

type FlavorTextEntry struct {
	FlavorText string        `json:"flavor_text"`
	Language   NamedResource `json:"language"`
}

// ---------- /evolution-chain/{id} ----------

type EvolutionChain struct {
	ID    int           `json:"id"`
	Chain EvolutionNode `json:"chain"`
}

type EvolutionNode struct {
	Species   NamedResource   `json:"species"`
	EvolvesTo []EvolutionNode `json:"evolves_to"`
}

// ---------- /pokemon-form/{id} ----------

type PokemonForm struct {
	ID      int             `json:"id"`
	Name    string          `json:"name"`
	Pokemon NamedResource   `json:"pokemon"`
	Names   []LocalizedName `json:"names"`
}

// ---------- typed helpers ----------

func (c *Client) Pokemon(id int) (*Pokemon, error) {
	var p Pokemon
	if err := c.Get(fmt.Sprintf("pokemon/%d", id), &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *Client) Species(id int) (*PokemonSpecies, error) {
	var s PokemonSpecies
	if err := c.Get(fmt.Sprintf("pokemon-species/%d", id), &s); err != nil {
		return nil, err
	}
	return &s, nil
}
