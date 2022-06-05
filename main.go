package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/3bl3gamer/tgclient/mtproto"
	"github.com/adrg/xdg"
)

const programName = "telegram-export-stickers"
const sessionFile = "telegram-export-stickers/tg.session"

func (t *Telegram) GetAllStickerSets() ([]mtproto.TL_stickerSet, error) {
	tl := t.Request(mtproto.TL_messages_getAllStickers{Hash: 0})
	allStickersRes, ok := tl.(mtproto.TL_messages_allStickers)
	if !ok {
		return nil, errors.New("TL_messages_getAllStickers failed")
	}

	tl = t.Request(mtproto.TL_messages_getArchivedStickers{OffsetID: 0, Limit: (1 << 31) - 1})
	archivedStickersRes, ok := tl.(mtproto.TL_messages_archivedStickers)
	if !ok {
		return nil, errors.New("TL_messages_getArchivedStickers failed")
	}

	stickerSets := make([]mtproto.TL_stickerSet, len(allStickersRes.Sets)+len(archivedStickersRes.Sets))

	for i, set := range allStickersRes.Sets {
		stickerSets[i] = set.(mtproto.TL_stickerSet)
	}

	for i, set := range archivedStickersRes.Sets {
		stickerSets[len(allStickersRes.Sets)+i] = set.(mtproto.TL_stickerSetCovered).Set.(mtproto.TL_stickerSet)
	}

	return stickerSets, nil
}

type StickerMetadata struct {
	Emoticons string `json:"emoticons"`
	Date      string `json:"date"`
}

type StickerSetMetadata struct {
	ID            int64                      `json:"id"`
	Title         string                     `json:"title"`
	ShortName     string                     `json:"short_name"`
	Count         int32                      `json:"count"`
	Archived      bool                       `json:"archived"`
	Animated      bool                       `json:"animated"`
	Gifs          bool                       `json:"gifs"`
	Masks         bool                       `json:"masks"`
	Official      bool                       `json:"official"`
	InstalledDate string                     `json:"installed_date"`
	ExportedDate  string                     `json:"exported_date"`
	Stickers      map[int64]*StickerMetadata `json:"stickers"`
}

func (t *Telegram) ExportStickerSet(inputStickerSet mtproto.TLReq) error {
	tl := t.Request(mtproto.TL_messages_getStickerSet{Stickerset: inputStickerSet})
	stickerSetRes, ok := tl.(mtproto.TL_messages_stickerSet)
	if !ok {
		return errors.New("TL_messages_getStickerSet failed")
	}
	stickerSet := stickerSetRes.Set.(mtproto.TL_stickerSet)

	err := os.Mkdir(stickerSet.ShortName, 0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	err = os.Chdir(stickerSet.ShortName)
	if err != nil {
		return err
	}

	metadata := StickerSetMetadata{
		ID:            stickerSet.ID,
		Title:         stickerSet.Title,
		ShortName:     stickerSet.ShortName,
		Count:         stickerSet.Count,
		Archived:      stickerSet.Archived,
		Gifs:          stickerSet.Gifs,
		Masks:         stickerSet.Masks,
		Official:      stickerSet.Official,
		InstalledDate: FormatDate(stickerSet.InstalledDate),
		ExportedDate:  Now(),
		Stickers:      make(map[int64]*StickerMetadata),
	}

	for i, _document := range stickerSetRes.Documents {
		document := _document.(mtproto.TL_document)
		metadata.Stickers[document.ID] = &StickerMetadata{Emoticons: "", Date: FormatDate(document.Date)}

		var extension string
		for _, attribute := range document.Attributes {
			if filenameAttribute, ok := attribute.(mtproto.TL_documentAttributeFilename); ok {
				parts := strings.Split(filenameAttribute.FileName, ".")
				extension = parts[len(parts)-1]
			}
		}
		filename := strconv.FormatInt(document.ID, 10) + "." + extension
		fileInfo, err := os.Stat(filename)
		exists := !errors.Is(err, os.ErrNotExist)
		if exists && fileInfo.Size() == int64(document.Size) {
			fmt.Printf("(%d/%d) Sticker %s already exported\n", i+1, len(stickerSetRes.Documents), filename)
		} else {
			fmt.Printf("(%d/%d) Exporting sticker %s\n", i+1, len(stickerSetRes.Documents), filename)
			err := t.DownloadDocument(filename, document)
			if err != nil {
				fmt.Printf("Failed to export sticker %s: %s\n", filename, err.Error())
			}
		}
	}
	for _, _pack := range stickerSetRes.Packs {
		pack := _pack.(mtproto.TL_stickerPack)
		for _, document := range pack.Documents {
			metadata.Stickers[document].Emoticons += pack.Emoticon
		}
	}

	encodedMetadata, err := json.MarshalIndent(metadata, "", "\t")
	if err != nil {
		return err
	}

	err = os.WriteFile("metadata.json", encodedMetadata, 0644)
	if err != nil {
		return err
	}

	err = os.Chdir("..")
	if err != nil {
		return err
	}

	return nil
}

func main() {
	args := ParseArgs()

	var sessionFilePath string
	if !args.DontSaveSession {
		var err error
		sessionFilePath, err = xdg.DataFile(sessionFile)
		if err != nil {
			panic(err)
		}
	}

	var t Telegram
	err := t.SignIn(args.AppID, args.AppHash, sessionFilePath)
	if err != nil {
		panic(err)
	}

	err = os.MkdirAll(args.Directory, 0755)
	if err != nil {
		panic(err)
	}
	err = os.Chdir(args.Directory)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Exporting stickerpacks to %s\n", args.Directory)
	if len(args.StickerSetNames) == 0 {
		stickerSets, err := t.GetAllStickerSets()
		if err != nil {
			panic(err)
		}

		for i, stickerSet := range stickerSets {
			fmt.Printf("(%d/%d) Exporting stickerpack %s (%s)\n", i+1, len(stickerSets), stickerSet.Title, stickerSet.ShortName)
			err = t.ExportStickerSet(mtproto.TL_inputStickerSetID{ID: stickerSet.ID, AccessHash: stickerSet.AccessHash})
			if err != nil {
				panic(err)
			}
		}
	} else {
		for i, stickerSetName := range args.StickerSetNames {
			fmt.Printf("(%d/%d) Exporting stickerpack %s\n", i+1, len(args.StickerSetNames), stickerSetName)
			err = t.ExportStickerSet(mtproto.TL_inputStickerSetShortName{ShortName: stickerSetName})
			if err != nil {
				panic(err)
			}
		}
	}
}
