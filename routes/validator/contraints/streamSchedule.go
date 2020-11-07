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

func (s *ScheduleStart) Validate(r *http.Request) url.Values {
	rules := govalidator.MapData{
		"startAt": []string{"required", "numeric"},
		"stopAt":  []string{"required", "numeric"},
		//"filename": []string{"required", "file_manifest_available"},
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
	return additionalCheckRules(s.Stream, errBag)
}

func additionalCheckRules(s *stream.ScheduledStream, errBag url.Values) url.Values {
	if s.StopAt <= s.StartAt {
		errBag.Add("startAt", "must be less than stop at")
		errBag.Add("stopAt", "must be greater than start at")
	}

	return errBag
}
