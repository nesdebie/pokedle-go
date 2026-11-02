NAME = pokedle

GREEN = \033[0;32m
RED   = \033[0;31m
NC    = \033[0m

# ---------- Server ----------

all: $(NAME) .env

$(NAME):
	go build -o $(NAME) ./cmd/pokedle
	@echo "$(RED)Usage: $(GREEN)./$(NAME)$(NC)"
	@echo "$(RED)Dev mode: $(GREEN)./$(NAME) dev$(NC)"

# Secret used to pick the Pokémon of the day (generated once).
.env:
	go run ./scripts/genkey

key:
	go run ./scripts/genkey

run: all
	./$(NAME)

dev: all
	./$(NAME) dev

test:
	go vet ./...
	go test ./...

# ---------- CSV generation (PokéAPI) ----------
# regionals reads data/pokemon_evolution_data.csv: run `make evolutions` first if it changed.

csv: names gen evolutions regionals

names:
	go run ./scripts/names

gen:
	go run ./scripts/generations

evolutions:
	go run ./scripts/evolutions

regionals:
	go run ./scripts/regionals

# ---------- Cleanup ----------

clean:
	@rm -f $(NAME) static/hint_*.ogg

fclean: clean
	@rm -f .env

re: fclean all

.PHONY: all $(NAME) key run dev test csv names gen evolutions regionals clean fclean re
