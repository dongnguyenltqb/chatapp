package handler

import (
	"chatapp/entity"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (app *App) validateGetPreSignedUploadUrl(i entity.Image) bool {
	if i.FileType == "" {
		return false
	}
	return true
}
func (app *App) GetPreSignedUploadUrl(c *gin.Context) {
	i := entity.Image{}
	if err := c.BindJSON(&i); err != nil {
		app.Logger.Error(err)
		c.AbortWithError(400, err)
		return
	}
	if app.validateGetPreSignedUploadUrl(i) == false {
		c.AbortWithError(400, errors.New("bad payload"))
		return
	}
	i.FileName = uuid.NewString()
	i.S3ObjectKey = i.FileName + "." + i.FileType
	url, err := i.GetPreSignedUploadUrl()
	if err != nil {
		c.String(400, err.Error())
		return
	}
	c.String(200, url)
}
