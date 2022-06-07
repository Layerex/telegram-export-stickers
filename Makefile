NAME=telegram-export-stickers

make:
	go build

install: make
	sudo install $(NAME) /usr/local/bin

lint:
	revive -config revive.toml
