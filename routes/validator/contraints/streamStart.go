package contraints

import (
	"fmt"
	"github.com/mamau/restream/stream"
	"github.com/thedevsaddam/govalidator"
	"net/http"
	"net/url"
)

type StreamStart struct {
	Stream *stream.Stream
}

func init() {
	govalidator.AddCustomRule("file_manifest_available", isFileManifestAvailable)
}

func (s *StreamStart) Validate(r *http.Request) url.Values {
	rules := govalidator.MapData{
		"filename": []string{"required", "file_manifest_available"},
		"name":     []string{"required"},
	}

	opts := govalidator.Options{
		Request: r,
		Rules:   rules,
		Data:    s.Stream,
	}
	v := govalidator.New(opts)
	errBag := v.ValidateJSON()
	return errBag
}

func isFileManifestAvailable(field string, rule string, message string, value interface{}) error {
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
