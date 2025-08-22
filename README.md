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
├── main.go                # Point d’entrée du serveur web
├── internal/
│   ├── game/
│   │   ├── game.go        # Logique du jeu (Pokémon du jour, vérif réponses)
│   ├── pokeapi/
│   │   ├── client.go      # Client HTTP pour PokéAPI
│   ├── data/
│       ├── names.go       # Gestion du CSV des noms multi-langues
│   └── utils/
│       ├── random.go      # Génération sécurisée aléatoire
├── static/
│   ├── js/                # Frontend JS (interaction)
│   ├── css/
│   └── index.html
└── data/
    └── pokemon_names.csv  # Ton CSV pré-généré

```

## ❗ Disclaimer
Pokédle uses Pokémon names and images, which are trademarks and copyrighted material owned by **The Pokémon Company** and **Nintendo**.  
This project is an **unofficial fan work** and is provided for **educational and fan purposes only**.  
The **source code** is under MIT License, but this does **not** apply to Pokémon assets.
