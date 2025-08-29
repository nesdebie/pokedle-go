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

csv: names gen evolutions regions

names:
	@if [ -f data/pokemon_names_multilang.csv ]; then \
		rm data/pokemon_names_multilang.csv; \
	fi
	go run scripts/update_names_csv.go

gen:
	@if [ -f data/pokemon_id_gen.csv ]; then \
		rm data/pokemon_id_gen.csv; \
	fi
	go run scripts/update_gen_csv.go

evolutions:
	@if [ -f data/pokemon_evolution_data.csv ]; then \
		rm data/pokemon_evolution_data.csv; \
	fi
	go run scripts/update_evolutions_lines_csv.go

regions:
	@if [ -f data/pokemon_forms.csv ]; then \
		rm data/pokemon_forms.csv; \
	fi
	go run scripts/update_regional_names_csv.go

clean:
	@rm -f $(NAME) go.mod go.sum static/hint*.ogg .env

re: clean all

.PHONY: all clean names gen re evolutions csv
