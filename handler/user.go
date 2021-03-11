package handler

import "github.com/gin-gonic/gin"

// GetUserByID : get user by user id
func (app *App) GetUserByID(c *gin.Context) {
	id, _ := c.Params.Get("id")
	app.Logger.Info(id)
	app.User.GetUserByID(123)
	c.String(200, "hello world")
}
