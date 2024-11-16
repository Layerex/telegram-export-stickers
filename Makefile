NAME=telegram-export-stickers

$(NAME): main.go client.go cli.go util.go
	go build

install: $(NAME)
	sudo install $(NAME) /usr/local/bin

lint:
	revive -config revive.toml

.PHONY: install lint
