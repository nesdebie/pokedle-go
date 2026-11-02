// Package dataset reads and writes the CSV files stored in data/.
package dataset

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// Default locations of the CSV files, relative to the project root.
const (
	Dir             = "data"
	NamesFile       = Dir + "/pokemon_names_multilang.csv"
	GenerationsFile = Dir + "/pokemon_id_gen.csv"
	EvolutionsFile  = Dir + "/pokemon_evolution_data.csv"
	FormsFile       = Dir + "/pokemon_forms.csv"
)

// Languages supported by the game, in display order.
var Languages = []string{"en", "fr", "de", "es", "it"}

// ReadCSV returns every row of a CSV file, header excluded.
func ReadCSV(path string) ([][]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return rows[1:], nil
}

// WriteCSV (re)creates a CSV file with the given header and rows.
func WriteCSV(path string, header []string, rows [][]string) error {
	if err := os.MkdirAll(filepath.Dir(path), os.ModePerm); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := csv.NewWriter(f)
	if err := w.Write(header); err != nil {
		return err
	}
	if err := w.WriteAll(rows); err != nil { // WriteAll flushes
		return err
	}
	return f.Close()
}

// atoi parses a trimmed integer cell.
func atoi(s string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(s))
}
