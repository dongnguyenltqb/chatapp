package reqvalidator

import (
	"errors"

	"github.com/xeipuuv/gojsonschema"
)

// Validate a go struct payload
func StructValidate(schemaLoader gojsonschema.JSONLoader, payload interface{}) error {
	documentLoader := gojsonschema.NewGoLoader(payload)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return err
	}

	if result.Valid() {
		return nil
	} else {
		textError := ""
		for _, err := range result.Errors() {
			textError += err.Description()
		}
		return errors.New(textError)
	}
}
