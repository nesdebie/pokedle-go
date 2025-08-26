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
	@echo "$(RED)Update csv: $(GREEN)make name && make gen$(NC)"
	go run genkey.go

names:
	@if [ -f data/pokemon_names_multilang.csv ]; then \
		rm data/pokemon_names_multilang.csv; \
	fi
	go run update_names_csv.go

gen:
	@if [ -f data/pokemon_id_gen.csv ]; then \
		rm data/pokemon_id_gen.csv; \
	fi
	go run update_gen_csv.go

clean:
	@rm -f $(NAME) go.mod go.sum .env

re: clean all

.PHONY: all clean names gen re
