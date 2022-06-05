package main

import (
	"runtime"
	"time"

	"github.com/3bl3gamer/tgclient"
	"github.com/3bl3gamer/tgclient/mtproto"
)

type Telegram struct {
	tgclient.TGClient
	user mtproto.TL_user
}

func (t *Telegram) Request(input mtproto.TLReq) mtproto.TL {
	return t.SendSyncRetry(input, time.Second, 0, time.Second*30)
}

func (t *Telegram) DownloadDocument(filepath string, document mtproto.TL_document) error {
	_, err := t.DownloadFileToPath(filepath, mtproto.TL_inputDocumentFileLocation{ID: document.ID, AccessHash: document.AccessHash, FileReference: document.FileReference}, document.DcID, int64(document.Size), &tgclient.NoopFileProgressHandler{})
	return err
}

func (t *Telegram) SignIn(appID int32, appHash string, sessionFilePath string) error {
	appConfig := &mtproto.AppConfig{
		AppID:          appID,
		AppHash:        appHash,
		AppVersion:     "0.0.1",
		DeviceModel:    "Unknown",
		SystemVersion:  runtime.GOOS + "/" + runtime.GOARCH,
		SystemLangCode: "en",
		LangPack:       "",
		LangCode:       "en",
	}
	var sessionStore mtproto.SessionStore
	if sessionFilePath != "" {
		sessionStore = &mtproto.SessFileStore{FPath: sessionFilePath}
	} else {
		sessionStore = &mtproto.SessNoopStore{}
	}
	t.TGClient = *tgclient.NewTGClientExt(appConfig, sessionStore, &mtproto.NoopLogHandler{}, nil)

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
