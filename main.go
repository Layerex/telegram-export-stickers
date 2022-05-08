package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/3bl3gamer/tgclient"
	"github.com/3bl3gamer/tgclient/mtproto"
	"github.com/adrg/xdg"
	"github.com/akamensky/argparse"
)

const programName = "telegram-export-stickers"
const sessionFilePath = "telegram-export-stickers/tg.session"

func FormatDate(date int32) string {
	return time.Unix(int64(date), 0).UTC().Format(time.RFC3339)
}

type Telegram struct {
	AppID   int32
	AppHash string
	tg      tgclient.TGClient
	user    mtproto.TL_user
}

func (t *Telegram) Request(input mtproto.TLReq) mtproto.TL {
	return t.tg.SendSyncRetry(input, time.Second, 0, time.Second*30)
}

type DummyProgressHandler struct{}

func (h *DummyProgressHandler) OnProgress(fileLocation mtproto.TL, offset, size int64) {}

type DummyLogHandler struct{}

func (h *DummyLogHandler) Log(level mtproto.LogLevel, err error, msg string, args ...interface{}) {}

func (h *DummyLogHandler) Message(isIncoming bool, msg mtproto.TL, id int64) {}

func (t *Telegram) DownloadDocument(filepath string, document mtproto.TL_document) error {
	_, err := t.tg.DownloadFileToPath(filepath, mtproto.TL_inputDocumentFileLocation{ID: document.ID, AccessHash: document.AccessHash, FileReference: document.FileReference}, document.DcID, int64(document.Size), &DummyProgressHandler{})
	return err
}

func (t *Telegram) SignIn() error {
	appConfig := &mtproto.AppConfig{
		AppID:          t.AppID,
		AppHash:        t.AppHash,
		AppVersion:     "0.0.1",
		DeviceModel:    "Unknown",
		SystemVersion:  runtime.GOOS + "/" + runtime.GOARCH,
		SystemLangCode: "en",
		LangPack:       "",
		LangCode:       "en",
	}
	sessionFile, err := xdg.DataFile(sessionFilePath)
	if err != nil {
		return err
	}
	session := &mtproto.SessFileStore{FPath: sessionFile}
	t.tg = *tgclient.NewTGClientExt(appConfig, session, &DummyLogHandler{}, nil)

	err = t.tg.InitAndConnect()
	if err != nil {
		return err
	}

	authDataProvider := mtproto.ScanfAuthDataProvider{}
	payload := mtproto.TL_users_getUsers{ID: []mtproto.TL{mtproto.TL_inputUserSelf{}}}
	res, err := t.tg.AuthExt(authDataProvider, payload)
	if err != nil {
		return err
	}
	t.user = res.(mtproto.VectorObject)[0].(mtproto.TL_user)
	return nil
}

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
	stickerSet := stickerSetRes.Set.(mtproto.TL_stickerSet)
	if !ok {
		return errors.New("TL_messages_getStickerSet failed")
	}

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
		ExportedDate:  time.Now().UTC().Format(time.RFC3339),
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
				fmt.Printf("Failed to export sticker %s: %s", filename, err.Error())
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
	parser := argparse.NewParser(programName, "Export sticker sets from telegram")
	stickerSetNames := parser.StringList("s", "stickerpacks", &argparse.Options{Required: false, Help: "Specify names or urls of stickerpacks to export (by default all stickerpacks of account are exported)."})
	directory := parser.String("d", "directory", &argparse.Options{Required: false, Default: "stickers", Help: "Directory to export stickers to"})
	appID := parser.Int("", "app-id", &argparse.Options{Required: false, Default: 17349, Help: "Test credentials are used by default"})
	appHash := parser.String("", "app-hash", &argparse.Options{Required: false, Default: "344583e45741c457fe1862106095a5eb", Help: "Test credentials are used by default"})

	err := parser.Parse(os.Args)
	if err != nil {
		fmt.Print(parser.Usage(err))
		os.Exit(2)
	}

	err = os.MkdirAll(*directory, 0755)
	if err != nil {
		panic(err)
	}
	err = os.Chdir(*directory)
	if err != nil {
		panic(err)
	}

	var t Telegram
	t.AppID = int32(*appID)
	t.AppHash = *appHash
	t.SignIn()

	fmt.Printf("Exporting stickerpacks to %s\n", *directory)
	if len(*stickerSetNames) == 0 {
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
		stickerPackUrlRegex := regexp.MustCompile("^(?:http://|https://)?[^/]+/addstickers/([^/]+)/*$")
		for i, stickerSetName := range *stickerSetNames {
			// Handle stickerpack urls
			parts := stickerPackUrlRegex.FindStringSubmatch(stickerSetName)
			if len(parts) > 0 {
				stickerSetName = parts[1]
			}
			fmt.Printf("(%d/%d) Exporting stickerpack %s\n", i+1, len(*stickerSetNames), stickerSetName)
			err = t.ExportStickerSet(mtproto.TL_inputStickerSetShortName{ShortName: stickerSetName})
			if err != nil {
				panic(err)
			}
		}
	}
}
