package validator

import (
	"github.com/mamau/restream/routes/response"
	"github.com/mamau/restream/routes/validator/contraints"
	"net/http"
)

func Validate(w http.ResponseWriter, r *http.Request, constraint contraints.Validatable) bool {
	errBag := constraint.Validate(r)

	if len(errBag) > 0 {
		response.JsonStruct(w, errBag, http.StatusBadRequest)
		return false
	}
	return true
}
