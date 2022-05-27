package main

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
)

const helpMessage =
`usage: %s [-h] [-d DIRECTORY] [--app-id APP_ID] [--app-hash APP_HASH] [STICKER_SETS ...]

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
`

type Args struct {
	StickerSetNames []string
	Directory       string
	AppID           int32
	AppHash         string
}

func ParseArgs() Args {
	stickerSetUrlRegex := regexp.MustCompile("^(?:http://|https://)?[^/]+/addstickers/([a-zA-Z0-9_]{5,32})/*")
	stickerSetNameRegex := regexp.MustCompile("[a-zA-Z0-9_]{5,32}")

	var args Args
	end := len(os.Args) - 1
	for i := 1; i < len(os.Args); i++ {
		inc := func() {
			if i == end {
				panic(fmt.Sprintf("Option %s requires a value", os.Args[i]))
			}
			i++
		}
		switch os.Args[i] {
		case "-s", "--stickerpacks":
			// Options ignored for compatibility
			continue
		case "-d", "--directory":
			inc()
			args.Directory = os.Args[i]
		case "--app-id":
			inc()
			argument, err := strconv.Atoi(os.Args[i])
			if err != nil {
				panic("--app-id value has to be a 32-bit integer")
			}
			args.AppID = int32(argument)
		case "--app-hash":
			inc()
			if len(os.Args[i]) != 32 || !IsHex(os.Args[i]) {
				panic("--app-hash value has to be a hex string of 32 characters")
			}
			args.AppHash = os.Args[i]
		case "-h", "--help":
			fmt.Printf(helpMessage, os.Args[0])
			os.Exit(0)
		default:
			var stickerSetName string
			// Handle sticker set urls
			match := stickerSetUrlRegex.FindStringSubmatch(os.Args[i])
			if len(match) > 0 {
				stickerSetName = match[1]
			} else {
				// Handle sticker set names
				if (stickerSetNameRegex.MatchString(os.Args[i])) {
					stickerSetName = os.Args[i]
				} else {
					panic(fmt.Sprintf("\"%s\" is not a sticker set name or an url", os.Args[i]));
				}
			}

			args.StickerSetNames = append(args.StickerSetNames, stickerSetName)
		}
	}

	if args.Directory == "" {
		args.Directory = "stickers/"
	}
	if args.AppID == 0 {
		if args.AppHash != "" {
			panic("--app-hash is provided, but --app-id isn't")
		}
		args.AppID = 17349;
	}
	if args.AppHash == "" {
		if args.AppID == 0 {
			panic("--app-id is provided, but --app-hash isn't")
		}
		args.AppHash = "344583e45741c457fe1862106095a5eb";
	}
	return args
}