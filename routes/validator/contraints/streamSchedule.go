package contraints

import (
	"github.com/mamau/restream/stream/scheduler"
	"github.com/thedevsaddam/govalidator"
	"net/http"
	"net/url"
)

type ScheduleStart struct {
	Stream *scheduler.ScheduledStream
}

func (s *ScheduleStart) Validate(r *http.Request) url.Values {
	rules := govalidator.MapData{
		"startAt":  []string{"required", "numeric"},
		"stopAt":   []string{"required", "numeric"},
		"manifest": []string{"required", "file_manifest_available"},
		"name":     []string{"required"},
	}

	opts := govalidator.Options{
		Request: r,
		Rules:   rules,
		Data:    s.Stream,
	}

	v := govalidator.New(opts)
	errBag := v.ValidateJSON()
	return additionalCheckRules(s.Stream, errBag)
}

func additionalCheckRules(s *scheduler.ScheduledStream, errBag url.Values) url.Values {
	if s.StopAt <= s.StartAt {
		errBag.Add("startAt", "must be less than stop at")
		errBag.Add("stopAt", "must be greater than start at")
	}

	return errBag
}
