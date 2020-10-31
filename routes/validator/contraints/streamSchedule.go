package contraints

import (
	"github.com/mamau/restream/stream"
	"github.com/thedevsaddam/govalidator"
	"net/http"
	"net/url"
)

type ScheduleStart struct {
	Stream *stream.ScheduledStream
}

func (s ScheduleStart) Validate(r *http.Request) url.Values {
	rules := govalidator.MapData{
		"startAt":  []string{"required"},
		"stopAt":   []string{"required"},
		"filename": []string{"required"},
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
