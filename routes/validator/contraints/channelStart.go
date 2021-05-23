package contraints

import (
	"github.com/mamau/restream/stream"
	"github.com/thedevsaddam/govalidator"
	"net/http"
	"net/url"
)

type ChannelStart struct {
	Stream *stream.Stream
}

func (c *ChannelStart) Validate(r *http.Request) url.Values {
	rules := govalidator.MapData{
		"name": []string{"required"},
	}

	opts := govalidator.Options{
		Request: r,
		Rules:   rules,
		Data:    c.Stream,
	}
	v := govalidator.New(opts)
	return v.ValidateJSON()
}
