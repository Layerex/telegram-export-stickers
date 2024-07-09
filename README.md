# telegram-export-stickers

Export sticker sets from telegram

## Installation

```sh
make install
```

## Usage

``` text
usage: telegram-export-stickers [-h] [-d DIRECTORY] [--dont-save-session] [--app-id APP_ID] [--app-hash APP_HASH] [STICKER_SETS ...]

Export sticker sets from telegram.

positional arguments:
  STICKER_SETS          Sticker set names or urls

options:
  -h, --help            Show this help message and exit
  -d DIRECTORY, --directory DIRECTORY
                        Directory to export stickers to
  --dont-save-session   Don't save session file (and don't use already saved one)
  --app-id APP_ID       Test credentials are used by default
  --app-hash APP_HASH   Test credentials are used by default
  -s, --stickerpacks    Ignored for compatibility

Session file is saved to /home/user/.local/share/telegram-export-stickers/tg.session
```
