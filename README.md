# telegram-export-stickers

Export sticker sets from telegram

## Installation

```sh
go build
sudo install telegram-export-stickers /usr/local/bin
```

## Usage

``` text
usage: telegram-export-stickers [-h] [-d DIRECTORY] [--app-id APP_ID] [--app-hash APP_HASH] [STICKER_SETS ...]

Export sticker sets from telegram.

positional arguments:
  STICKER_SETS          Sticker set names or urls

options:
  -h, --help            Show this help message and exit
  -d DIRECTORY, --directory DIRECTORY
                        Directory to export stickers to
  --app-id APP_ID       Test credentials are used by default
  --app-hash APP_HASH   Test credentials are used by default
  -s, --stickerpacks    Ignored for compatibility
```
