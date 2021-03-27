package reqschema

import "github.com/xeipuuv/gojsonschema"

var GetPreSignedUploadUrlRequestSchemaLoader = gojsonschema.NewStringLoader(`
	{
		"type":"object",
		"properties":{
			"fileType":{
				"type":"string",
				"enum":["png","jpg","jpeg"]
			},
			"someField":{
				"type":"string",
				"default":"default value"
			}
		},
		"required":["fileType"],
		"additionalProperties":false
	}`)
