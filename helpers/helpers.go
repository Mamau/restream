package helpers

import (
	"encoding/json"
	"net/http"
)

func JsonRequestToMap(r *http.Request, s interface{}) error {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(s)
	if err != nil {
		return err
	}
	return nil
}
