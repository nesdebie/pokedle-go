package dataset

// Evolution describes where a Pokémon sits in its evolution line.
type Evolution struct {
	Position       int // 0 = basic, 1 = first evolution, 2 = second evolution
	IsFullyEvolved int // 1 if the Pokémon cannot evolve further
}

// LoadEvolutions reads data/pokemon_evolution_data.csv (id,position,is_fully_evolved).
func LoadEvolutions(path string) (map[int]Evolution, error) {
	records, err := ReadCSV(path)
	if err != nil {
		return nil, err
	}

	evolutions := make(map[int]Evolution, len(records))
	for _, rec := range records {
		if len(rec) < 3 {
			continue
		}
		id, err1 := atoi(rec[0])
		pos, err2 := atoi(rec[1])
		full, err3 := atoi(rec[2])
		if err1 != nil || err2 != nil || err3 != nil {
			continue
		}
		evolutions[id] = Evolution{Position: pos, IsFullyEvolved: full}
	}
	return evolutions, nil
}

// LoadGenerations reads data/pokemon_id_gen.csv (id,gen).
func LoadGenerations(path string) (map[int]int, error) {
	records, err := ReadCSV(path)
	if err != nil {
		return nil, err
	}

	generations := make(map[int]int, len(records))
	for _, rec := range records {
		if len(rec) < 2 {
			continue
		}
		id, err1 := atoi(rec[0])
		gen, err2 := atoi(rec[1])
		if err1 != nil || err2 != nil {
			continue
		}
		generations[id] = gen
	}
	return generations, nil
}
