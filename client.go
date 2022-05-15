package main

import (
	"runtime"
	"time"

	"github.com/3bl3gamer/tgclient"
	"github.com/3bl3gamer/tgclient/mtproto"
)

type DummyProgressHandler struct{}

func (h *DummyProgressHandler) OnProgress(fileLocation mtproto.TL, offset, size int64) {}

type DummyLogHandler struct{}

func (h *DummyLogHandler) Log(level mtproto.LogLevel, err error, msg string, args ...interface{}) {}

func (h *DummyLogHandler) Message(isIncoming bool, msg mtproto.TL, id int64) {}

type Telegram struct {
	tgclient.TGClient
	user    mtproto.TL_user
	AppID   int32
	AppHash string
}

func (t *Telegram) Request(input mtproto.TLReq) mtproto.TL {
	return t.SendSyncRetry(input, time.Second, 0, time.Second*30)
}

func (t *Telegram) DownloadDocument(filepath string, document mtproto.TL_document) error {
	_, err := t.DownloadFileToPath(filepath, mtproto.TL_inputDocumentFileLocation{ID: document.ID, AccessHash: document.AccessHash, FileReference: document.FileReference}, document.DcID, int64(document.Size), &DummyProgressHandler{})
	return err
}

func (t *Telegram) SignIn(sessionFilePath string) error {
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
	session := &mtproto.SessFileStore{FPath: sessionFilePath}
	t.TGClient = *tgclient.NewTGClientExt(appConfig, session, &DummyLogHandler{}, nil)

	err := t.InitAndConnect()
	if err != nil {
		return err
	}

	authDataProvider := mtproto.ScanfAuthDataProvider{}
	payload := mtproto.TL_users_getUsers{ID: []mtproto.TL{mtproto.TL_inputUserSelf{}}}
	res, err := t.AuthExt(authDataProvider, payload)
	if err != nil {
		return err
	}
	t.user = res.(mtproto.VectorObject)[0].(mtproto.TL_user)
	return nil
}
