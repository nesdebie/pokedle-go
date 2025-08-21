# pokedle-go

Pokédle is a Wordle-like game built in Go where players guess Pokémon.

## 🚀 Features
- Guess Pokémon names in a Wordle-style game
- Multilingual support (planned)
- Fetches Pokémon data and images from [PokéAPI](https://pokeapi.co/)

## ⚖️ License
This project’s **source code** is licensed under the [MIT License](LICENSE).

## FileTree
```
pokedle/
├── LICENSE
├── README.md
├── go.mod
├── go.sum
├── cmd/
│   └── pokedle/       # point d'entrée (main.go)
│       └── main.go
├── internal/          # code interne
│   ├── game/          # logique du jeu
│   │   └── game.go
│   ├── pokemon/       # gestion des données Pokémon (API, noms, images)
│   │   └── pokemon.go
│   └── utils/         # helpers divers
│       └── utils.go
└── assets/ (optionnel, vide ou ignoré)
    └── .gitkeep
```

## ❗ Disclaimer
Pokédle uses Pokémon names and images, which are trademarks and copyrighted material owned by **The Pokémon Company** and **Nintendo**.  
This project is an **unofficial fan work** and is provided for **educational and fan purposes only**.  
The **source code** is under MIT License, but this does **not** apply to Pokémon assets.
