package dataset

import (
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// NamesRow is one Pokémon with its name in every supported language.
type NamesRow struct {
	ID    int
	Names map[string]string // lang -> name
}

// NameIndex allows looking up a Pokémon by any of its localized names.
type NameIndex struct {
	rows    []NamesRow
	idByKey map[string]int // normalized name -> id
	enByID  map[int]string
}

// LoadNames reads data/pokemon_names_multilang.csv (id,en,fr,de,es,it).
func LoadNames(path string) (*NameIndex, error) {
	records, err := ReadCSV(path)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("%s: no rows", path)
	}

	idx := &NameIndex{
		idByKey: make(map[string]int),
		enByID:  make(map[int]string),
	}
	for _, rec := range records {
		if len(rec) < 1+len(Languages) {
			continue
		}
		id, err := atoi(rec[0])
		if err != nil {
			continue
		}

		row := NamesRow{ID: id, Names: make(map[string]string, len(Languages))}
		for i, lang := range Languages {
			name := rec[i+1]
			row.Names[lang] = name
			if k := NormalizeKey(name); k != "" {
				idx.idByKey[k] = id
			}
		}
		idx.rows = append(idx.rows, row)
		idx.enByID[id] = row.Names["en"]
	}
	return idx, nil
}

// Len returns the number of Pokémon in the index.
func (n *NameIndex) Len() int { return len(n.rows) }

// IDAt returns the Pokémon ID stored at row i (0 if out of range).
func (n *NameIndex) IDAt(i int) int {
	if i < 0 || i >= len(n.rows) {
		return 0
	}
	return n.rows[i].ID
}

// Lookup finds a Pokémon ID from a name typed in any supported language.
func (n *NameIndex) Lookup(name string) (int, bool) {
	id, ok := n.idByKey[NormalizeKey(name)]
	return id, ok
}

// EnglishName returns the English name of a Pokémon.
func (n *NameIndex) EnglishName(id int) string { return n.enByID[id] }

// SuggestionGroup lists the names starting with the query for one language.
type SuggestionGroup struct {
	Lang  string   `json:"lang"`
	Names []string `json:"names"`
}

// Suggest returns, per language, every name starting with query
// (case and accent insensitive). Languages without matches are omitted.
func (n *NameIndex) Suggest(query string) []SuggestionGroup {
	groups := []SuggestionGroup{}
	q := NormalizeKey(query)
	if q == "" {
		return groups
	}

	for _, lang := range Languages {
		var names []string
		for _, row := range n.rows {
			if name := row.Names[lang]; strings.HasPrefix(NormalizeKey(name), q) {
				names = append(names, name)
			}
		}
		if len(names) > 0 {
			groups = append(groups, SuggestionGroup{Lang: lang, Names: names})
		}
	}
	return groups
}

// NormalizeKey lowercases, trims and strips accents so that
// "Évoli", "evoli" and " EVOLI " all match.
func NormalizeKey(s string) string {
	return strings.ToLower(removeAccents(strings.TrimSpace(s)))
}

func removeAccents(s string) string {
	var b strings.Builder
	for _, r := range norm.NFD.String(s) {
		if !unicode.Is(unicode.Mn, r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}
