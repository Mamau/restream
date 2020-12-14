package contraints

import (
	"github.com/mamau/restream/stream"
	"github.com/thedevsaddam/govalidator"
	"net/http"
	"net/url"
)

type StreamStart struct {
	Stream *stream.Stream
}

func (s *StreamStart) Validate(r *http.Request) url.Values {
	rules := govalidator.MapData{
		"filename": []string{"required"},
		"name":     []string{"required"},
	}

	opts := govalidator.Options{
		Request: r,
		Rules:   rules,
		Data:    s.Stream,
	}
	v := govalidator.New(opts)
	return v.ValidateJSON()
}
