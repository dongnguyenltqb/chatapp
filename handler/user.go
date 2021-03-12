package handler

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type googleUserInfoResponse struct {
	Sub           string `json:"sub"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

func (app *App) GetBySession(c *gin.Context) {
	session := sessions.Default(c)
	email := session.Get("email")
	if email != nil {
		picture := session.Get("picture")
		sub := session.Get("sub")
		email_verified := session.Get("email_verified")
		m := googleUserInfoResponse{
			Email:         email.(string),
			Picture:       picture.(string),
			Sub:           sub.(string),
			EmailVerified: email_verified.(bool),
		}
		c.JSON(200, m)
	} else {
		c.AbortWithStatus(400)
	}
}

// GetUserByID : get user by user id
func (app *App) GetUserByID(c *gin.Context) {
	id, _ := c.Params.Get("id")
	app.Logger.Info(id)
	app.User.GetUserByID(123)
	c.String(200, "hello world")
}

// Login by google
func (app *App) Login(c *gin.Context) {
	app.Logger.Info(cfg)
	url := getConf().AuthCodeURL("login")
	c.Redirect(301, url)
}

func (app *App) LoginCallback(c *gin.Context) {
	tok, err := getConf().Exchange(oauth2.NoContext, c.Query("code"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	client := getConf().Client(oauth2.NoContext, tok)
	email, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	defer email.Body.Close()
	data, _ := ioutil.ReadAll(email.Body)

	u := googleUserInfoResponse{}
	if err := json.Unmarshal(data, &u); err != nil {
		c.AbortWithStatus(500)
		return
	}

	session := sessions.Default(c)
	session.Set("email", u.Email)
	session.Set("picture", u.Picture)
	session.Set("sub", u.Sub)
	session.Set("email_verified", u.EmailVerified)
	session.Save()
	c.Redirect(301, "/")
}
