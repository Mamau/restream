package contraints

import (
	"fmt"
	"github.com/thedevsaddam/govalidator"
	"net/http"
	"net/url"
)

type Validatable interface {
	Validate(r *http.Request) url.Values
}

func init() {
	govalidator.AddCustomRule("file_manifest_available", isFileManifestAvailable)
}

func isFileManifestAvailable(_ string, _ string, _ string, value interface{}) error {
	fileManifest := value.(string)

	resp, err := http.Get(fileManifest)
	if err != nil {
		return fmt.Errorf("Error while check manifest %v with err: %v\n", fileManifest, err)
	}
	isOk := resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices
	if !isOk {
		return fmt.Errorf("File %v is not available %v\n", fileManifest, resp.StatusCode)
	}

	return nil
}
