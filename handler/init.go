package handler

import (
	"chatapp/entity"
	"chatapp/util/logger"
	"os"
	"sync"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
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

func Run() {
	// Init app handler
	app := &App{
		Logger: logger.Get(),
		User:   entity.User{},
	}
	// Init ws handler
	hub := getHub()

	gin.SetMode(os.Getenv("GIN_MODE"))
	server := gin.Default()
	// Setup session storage
	store, _ := redis.NewStore(10, "tcp", os.Getenv("redis_addr"), os.Getenv("redis_password"), []byte(os.Getenv("session_secret")))
	server.Use(sessions.Sessions("mysession", store))

	server.GET("/", func(c *gin.Context) {
		content, _ := os.ReadFile("index.html")
		c.Header("content-type", "text/html")
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0")
		c.String(200, string(content))
	})
	server.GET("/ws", func(c *gin.Context) {
		hub.serveWs(c.Writer, c.Request)
	})

	server.GET("/login", app.Login)
	server.GET("/login/callback", app.LoginCallback)
	server.GET("/users/bySession/", app.GetBySession)
	server.GET("/users/byId/:id", app.GetUserByID)
	if err := server.Run(); err != nil {
		panic(err)
	}
}
