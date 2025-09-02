NAME = pokedle
SRC = main.go

GREEN = \033[0;32m
RED = \033[0;31m
NC = \033[0m

all: $(NAME)

$(NAME):
	go mod init $(NAME)
	go get golang.org/x/text/unicode/norm
	go build -o $(NAME) $(SRC)
	@echo "$(RED)Usage: $(GREEN)./$(NAME)$(NC)"
	@echo "$(RED)Dev mode: $(GREEN)./$(NAME) dev$(NC)"
	go run scripts/genkey.go

csv: names gen evolutions regionals

names:
	@if [ -f data/pokemon_names_multilang.csv ]; then \
		rm data/pokemon_names_multilang.csv; \
	fi
	go run scripts/get_sts_names_multilang.go

gen:
	@if [ -f data/pokemon_id_gen.csv ]; then \
		rm data/pokemon_id_gen.csv; \
	fi
	go run scripts/get_std_generations_infos.go

evolutions:
	@if [ -f data/pokemon_evolution_data.csv ]; then \
		rm data/pokemon_evolution_data.csv; \
	fi
	go run scripts/get_std_evolution_lines_infos.go

regionals:
	@if [ -f data/pokemon_forms.csv ]; then \
		rm data/pokemon_forms.csv; \
	fi
	go run scripts/get_regionals_infos.go

clean:
	@rm -f $(NAME) go.mod go.sum static/hint*.ogg .env

re: clean all

.PHONY: all clean names gen re evolutions csv regionals
