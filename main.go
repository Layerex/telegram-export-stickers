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
const sessionFile = programName + "/tg.session"

func (t *Telegram) GetAllStickerSets() ([]mtproto.TL_stickerSet, error) {
	var archivedStickerSetsLimit int32 = 100

	tl := t.Request(mtproto.TL_messages_getAllStickers{Hash: 0})
	allStickersRes, ok := tl.(mtproto.TL_messages_allStickers)
	if !ok {
		return nil, errors.New("TL_messages_getAllStickers failed")
	}
	totalInstalledStickerSets := len(allStickersRes.Sets)
	fmt.Println(totalInstalledStickerSets, "stickerpacks installed ")

	fmt.Println("Getting page 1 of archived stickerpacks")
	tl = t.Request(mtproto.TL_messages_getArchivedStickers{OffsetID: -1, Limit: archivedStickerSetsLimit})
	archivedStickersRes, ok := tl.(mtproto.TL_messages_archivedStickers)
	if !ok {
		return nil, errors.New("TL_messages_getArchivedStickers failed")
	}
	// archivedStickersRes.Count includes deleted stickerpacks
	totalArchivedStickerSets := int(archivedStickersRes.Count-archivedStickerSetsLimit) + len(archivedStickersRes.Sets)
	totalStickerSets := totalInstalledStickerSets + totalArchivedStickerSets
	stickerSets := make([]mtproto.TL_stickerSet, 0, totalStickerSets)
	stickerSets = append(stickerSets, allStickersRes.Sets...)
	for page := 2; len(stickerSets) != totalStickerSets; page++ {
		fmt.Println("Getting page", page, "of archived stickerpacks")
		for _, set := range archivedStickersRes.Sets {
			stickerSets = append(stickerSets, set.(mtproto.TL_stickerSetCovered).Set)
		}
		tl = t.Request(mtproto.TL_messages_getArchivedStickers{OffsetID: stickerSets[len(stickerSets)-1].ID, Limit: archivedStickerSetsLimit})
		archivedStickersRes, ok = tl.(mtproto.TL_messages_archivedStickers)
		if !ok {
			return nil, errors.New("TL_messages_getArchivedStickers failed")
		}
	}
	fmt.Println(totalArchivedStickerSets, "stickerpacks archived ")
	archivedStickerSetsDeleted := int(archivedStickersRes.Count) - totalArchivedStickerSets
	if archivedStickerSetsDeleted == 1 {
		fmt.Println("1 stickerpack deleted while archived")
	} else {
		fmt.Println(archivedStickerSetsDeleted, "stickerpacks deleted while archived")
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

func (t *Telegram) ExportStickerSet(inputStickerSet mtproto.TL) error {
	tl := t.Request(mtproto.TL_messages_getStickerSet{Stickerset: inputStickerSet})
	stickerSetRes, ok := tl.(mtproto.TL_messages_stickerSet)
	if !ok {
		return errors.New("TL_messages_getStickerSet failed")
	}

	err := os.Mkdir(stickerSetRes.Set.ShortName, 0755)
	if err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}
	err = os.Chdir(stickerSetRes.Set.ShortName)
	if err != nil {
		return err
	}

	metadata := StickerSetMetadata{
		ID:            stickerSetRes.Set.ID,
		Title:         stickerSetRes.Set.Title,
		ShortName:     stickerSetRes.Set.ShortName,
		Count:         stickerSetRes.Set.Count,
		Archived:      stickerSetRes.Set.Archived,
		Masks:         stickerSetRes.Set.Masks,
		Official:      stickerSetRes.Set.Official,
		InstalledDate: FormatDate(*stickerSetRes.Set.InstalledDate),
		ExportedDate:  Now(),
		Stickers:      make(map[int64]*StickerMetadata),
	}

	var alreadyExported int
	printAlreadyExported := func(i int) {
		if alreadyExported > 1 {
			fmt.Printf("(%d-%d/%d) Stickers already exported\n", i-alreadyExported+1, i, len(stickerSetRes.Documents))
		} else if alreadyExported == 1 {
			fmt.Printf("(%d/%d) Sticker already exported\n", i, len(stickerSetRes.Documents))
		}
		alreadyExported = 0
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
			alreadyExported++
		} else {
			printAlreadyExported(i)
			fmt.Printf("(%d/%d) Exporting sticker\n", i+1, len(stickerSetRes.Documents))
			err := t.DownloadDocument(filename, document)
			if err != nil {
				fmt.Printf("Failed to export sticker: %s\n", err.Error())
			}
		}
	}

	if alreadyExported == len(stickerSetRes.Documents) {
		fmt.Println("All stickers already exported")
	} else {
		printAlreadyExported(len(stickerSetRes.Documents))
	}

	for _, pack := range stickerSetRes.Packs {
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
