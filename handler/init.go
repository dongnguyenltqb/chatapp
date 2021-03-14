package handler

import (
	"chatapp/entity"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type App struct {
	Logger *logrus.Logger
	User   entity.User
}

type HandlerResp struct {
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
	Error   bool        `json:"error"`
	Message string      `json:"message"`
}

var oauth2Cfg *oauth2.Config
var onceInitOauth2Cfg sync.Once

func getConf() *oauth2.Config {
	onceInitOauth2Cfg.Do(func() {
		oauth2Cfg = &oauth2.Config{
			ClientID:     os.Getenv("google_client_id"),
			ClientSecret: os.Getenv("google_client_secret"),
			RedirectURL:  os.Getenv("google_redirect_url"),
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
			},
			Endpoint: google.Endpoint,
		}
	})
	return oauth2Cfg
}
