# telegram-export-stickers

Export sticker sets from telegram

## Installation

```sh
go build
sudo install telegram-export-stickers /usr/local/bin
```

## Usage

``` text
usage: telegram-export-stickers [-h|--help] [-s|--stickerpacks "<value>"
                                [-s|--stickerpacks "<value>" ...]]
                                [-d|--directory "<value>"] [--app-id <integer>]
                                [--app-hash "<value>"]

                                Export sticker sets from telegram

Arguments:

  -h  --help          Print help information
  -s  --stickerpacks  Specify names or urls of stickerpacks to export (by
                      default all stickerpacks of account are exported),
  -d  --directory     Directory to export stickers to. Default: stickers
      --app-id        Test credentials are used by default. Default: 17349
      --app-hash      Test credentials are used by default. Default:
                      344583e45741c457fe1862106095a5eb
```
