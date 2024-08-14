NAME=telegram-export-stickers

$(NAME):
	go build

install: $(NAME)
	sudo install $(NAME) /usr/local/bin

lint:
	revive -config revive.toml

.PHONY: install lint
