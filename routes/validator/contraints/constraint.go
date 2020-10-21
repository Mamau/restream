package contraints

import (
	"net/http"
	"net/url"
)

type Validatable interface {
	Validate(r *http.Request) url.Values
}
