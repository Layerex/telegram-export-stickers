package main

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strconv"

	"github.com/adrg/xdg"
)

const helpMessage = `usage: %s [-h] [-d DIRECTORY] [--dont-save-session] [--app-id APP_ID] [--app-hash APP_HASH] [STICKER_SETS ...]

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

Session file is saved to %s
`

type Args struct {
	StickerSetNames []string
	Directory       string
	DontSaveSession bool
	AppID           int32
	AppHash         string
}

func ParseArgs() Args {
	stickerSetURLRegex := regexp.MustCompile("^(?:http://|https://)?[^/]+/addstickers/([a-zA-Z0-9_]{5,32})/*")
	stickerSetNameRegex := regexp.MustCompile("[a-zA-Z0-9_]{5,32}")

	var args Args
	end := len(os.Args) - 1
	for i := 1; i < len(os.Args); i++ {
		nextArg := func() {
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
			if args.Directory != "" {
				panic("--directory option is provided more than one time")
			}
			nextArg()
			args.Directory = os.Args[i]
		case "--dont-save-session":
			if args.DontSaveSession {
				panic("--dont-save-session option is provided more than one time")
			}
			args.DontSaveSession = true
		case "--app-id":
			if args.AppID != 0 {
				panic("--app-id option is provided more than one time")
			}
			nextArg()
			argument, err := strconv.Atoi(os.Args[i])
			if err != nil {
				panic("--app-id value has to be a 32-bit integer")
			}
			args.AppID = int32(argument)
		case "--app-hash":
			if args.AppHash != "" {
				panic("--app-hash option is provided more than one time")
			}
			nextArg()
			if len(os.Args[i]) != 32 || !IsHex(os.Args[i]) {
				panic("--app-hash value has to be a hex string of 32 characters")
			}
			args.AppHash = os.Args[i]
		case "-h", "--help":
			fmt.Printf(helpMessage, os.Args[0], path.Join(xdg.DataHome, sessionFile))
			os.Exit(0)
		default:
			var stickerSetName string
			// Handle sticker set urls
			match := stickerSetURLRegex.FindStringSubmatch(os.Args[i])
			if len(match) > 0 {
				stickerSetName = match[1]
			} else {
				// Handle sticker set names
				if !stickerSetNameRegex.MatchString(os.Args[i]) {
					panic(fmt.Sprintf("\"%s\" is not a sticker set name or an url", os.Args[i]))
				}
				stickerSetName = os.Args[i]
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
		args.AppID = 17349
	}
	if args.AppHash == "" {
		if args.AppID == 0 {
			panic("--app-id is provided, but --app-hash isn't")
		}
		args.AppHash = "344583e45741c457fe1862106095a5eb"
	}
	return args
}
