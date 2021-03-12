package main

import (
	"chatapp/entity"
	"chatapp/handler"
	"chatapp/logger"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	app := &handler.App{
		Logger: logger.Get(),
		User:   entity.User{},
	}
	server := gin.Default()

	// Spin ws handler
	hub := handler.NewHub()
	go hub.Run()

	server.GET("/", func(c *gin.Context) {
		content, _ := os.ReadFile("index.html")
		c.Header("content-type", "text/html")
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate, max-age=0")
		c.String(200, string(content))
	})
	server.GET("/ws", func(c *gin.Context) {
		handler.ServeWs(hub, c.Writer, c.Request)
	})
	server.GET("/byId/:id", app.GetUserByID)
	server.Run()
}
