# pokedle-go

Pokédle is a Wordle-like game built in Go where players guess Pokémon.

## 🚀 Features
- Guess Pokémon names in a Wordle-style game
- Multilingual support (planned)
- Fetches Pokémon data and images from [PokéAPI](https://pokeapi.co/) v2.

## ⚖️ License
This project’s **source code** is licensed under the [MIT License](LICENSE).

## 🛠️ Usage
Requires Go 1.24+.

```bash
make          # build ./pokedle and generate .env (POKEDLE_SECRET) if missing
./pokedle     # http://localhost:8080 (PORT env var to change it)
./pokedle dev # dev mode: cookies expire after 5 minutes
make test     # go vet + unit tests
```

The CSV files in `data/` are already committed. To regenerate them from PokéAPI:

```bash
make csv      # names, gen, evolutions, regionals (in this order)
```

## FileTree
```
├── cmd/pokedle/            # server entry point (main.go)
├── internal/
│   ├── dataset/            # CSV paths, loaders/writers, name index & suggestions
│   ├── game/               # daily Pokémon pick, guess vs target comparison
│   ├── pokeapi/            # PokéAPI client & types (shared by server and scripts)
│   └── server/             # HTTP handlers (/api/guess, /api/hints, /api/today, /api/suggest)
├── scripts/                # one-shot tools, run with `go run ./scripts/<name>`
│   ├── genkey/             # random POKEDLE_SECRET -> .env
│   ├── names/              # -> data/pokemon_names_multilang.csv
│   ├── generations/        # -> data/pokemon_id_gen.csv
│   ├── evolutions/         # -> data/pokemon_evolution_data.csv
│   └── regionals/          # -> data/pokemon_forms.csv (needs the evolutions CSV)
├── data/                   # every Pokémon up to Pecharunt (no Mega / Dynamax)
│   ├── pokemon_evolution_data.csv
│   ├── pokemon_forms.csv
│   ├── pokemon_id_gen.csv
│   └── pokemon_names_multilang.csv
├── static/
│   ├── js/
│   │   ├── main.js         # entry point, guess form
│   │   ├── api.js          # backend calls
│   │   ├── badges.js       # comparison badges under each guess
│   │   ├── hints.js        # hint box (cry / types / Pokédex description)
│   │   ├── cryPlayer.js    # audio player for the cry hint
│   │   └── suggestions.js  # autocomplete dropdown
│   ├── fonts/
│   ├── images/
│   ├── index.html
│   └── styles.css
├── Makefile
└── go.mod
```

## ❗ Disclaimer
Pokédle uses Pokémon names and images, which are trademarks and copyrighted material owned by **The Pokémon Company** and **Nintendo**.  
This project is an **unofficial fan work** and is provided for **educational and fan purposes only**.
