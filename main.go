package main

import (
	"chatapp/entity"
	"chatapp/handler"
	"chatapp/infra"
	"chatapp/util/logger"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/redis"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}
}

func main() {
	// Setup infra
	infra.Setup()

	// Init app handler
	app := &handler.App{
		Logger: logger.Get(),
		User:   entity.User{},
	}
	// Init ws handler
	hub := handler.GetHub()

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
		handler.ServeWs(hub, c.Writer, c.Request)
	})

	server.GET("/login", app.Login)
	server.GET("/login/callback", app.LoginCallback)
	// server.GET("/users/bySession/", app.GetBySession)
	// server.GET("/users/byId/:id", app.GetUserByID)
	if err := server.Run(); err != nil {
		panic(err)
	}
}
