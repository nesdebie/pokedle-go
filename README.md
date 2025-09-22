# pokedle-go

Pokédle is a Wordle-like game built in Go where players guess Pokémon.

## 🚀 Features
- Guess Pokémon names in a Wordle-style game
- Multilingual support (planned)
- Fetches Pokémon data and images from [PokéAPI](https://pokeapi.co/) v2.

## ⚖️ License
This project’s **source code** is licensed under the [MIT License](LICENSE).

## FileTree
```
├── data/
│   ├── pokemon_evolution_data.csv
│   ├── pokemon_forms.csv
│   ├── pokemon_id_gen.csv
│   └── pokemon_names_multilang.csv
├── scripts/
│   ├── genkey.go
│   ├── get_regionals_infos.go
│   ├── get_std_evolution_lines_infos.go
│   ├── get_std_generations_infos.go
│   └── get_std_names_multilang.go
├── static/
│   ├── img/
│   │   └── language.svg
│   ├── app.js
│   ├── index.html
│   └── styles.css
├── Makefile
└── main.go
```

## ❗ Disclaimer
Pokédle uses Pokémon names and images, which are trademarks and copyrighted material owned by **The Pokémon Company** and **Nintendo**.  
This project is an **unofficial fan work** and is provided for **educational and fan purposes only**.  
The **source code** is under MIT License, but this does **not** apply to Pokémon assets.
