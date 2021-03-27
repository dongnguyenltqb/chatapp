package handler

import (
	"chatapp/entity"
	reqschema "chatapp/handler/validation/schema"
	reqvalidator "chatapp/handler/validation/validator"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (app *App) GetPreSignedUploadUrl(c *gin.Context) {
	// bind request payload
	reqPayload := reqschema.GetPreSignedUploadUrlRequestSchema{}
	if err := c.BindJSON(&reqPayload); err != nil {
		app.Logger.Error(err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	// validate request payload
	if err := reqvalidator.StructValidate(reqschema.GetPreSignedUploadUrlRequestSchemaLoader, reqPayload); err != nil {
		app.Logger.Error(err)
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	i := entity.Image{
		FileType: *reqPayload.FileType,
		FileName: uuid.NewString(),
	}
	i.S3ObjectKey = i.FileName + "." + i.FileType
	url, err := i.GetPreSignedUploadUrl()
	if err != nil {
		app.Logger.Error(err)
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.String(http.StatusOK, url)
}
